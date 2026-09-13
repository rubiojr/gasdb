package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rubiojr/gasdb/_server/internal/pricesummary"
	"github.com/rubiojr/gasdb/_server/internal/search"
	"github.com/rubiojr/gasdb/_server/internal/version"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/tkrajina/gpxgo/gpx"
)

type apiHandler struct {
	storage  *gasdb.Storage
	geocoder *locationGeocoder
	logger   *slog.Logger
}

func (h apiHandler) search(w http.ResponseWriter, r *http.Request) {
	query, err := parseAPIQuery(r.URL.Query())
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if query.Location != "" {
		place, err := h.geocoder.lookup(r.Context(), query.Location)
		if err != nil {
			h.lookupError(w, err)
			return
		}
		query.Lat, query.Lng, query.ResolvedLocation = place.lat, place.lng, place.name
	}
	data, meta, err := h.snapshot(r.Context())
	if err != nil {
		h.writeError(w, http.StatusServiceUnavailable, "prices_unavailable", "Fuel price data is unavailable.")
		return
	}
	nearby, err := h.storage.NearbyPrices(r.Context(), query.Lat, query.Lng, query.RadiusKM*1000)
	if err != nil {
		h.logger.Error("JSON nearby search failed", "error", err)
		h.writeError(w, http.StatusInternalServerError, "search_failed", "Unable to search fuel stations.")
		return
	}
	stations := apiStations(nearby, query)
	response := apiSearchResponse{
		State: "ok", Meta: meta, Query: query,
		TotalStations: len(stations), Truncated: len(stations) > query.Limit,
		Stations:          stations[:min(len(stations), query.Limit)],
		RadiusSuggestions: []apiRadiusSuggestion{}, WebURL: query.webURL(),
	}
	response.ReturnedStations = len(response.Stations)
	if len(stations) == 0 {
		response.State = "no_stations"
		for _, suggestion := range search.SuggestRadii(data.ListaEESSPrecio, query.Lat, query.Lng, query.RadiusKM, query.values()) {
			response.RadiusSuggestions = append(response.RadiusSuggestions, apiRadiusSuggestion{
				RadiusKM: suggestion.Radius, StationCount: suggestion.Count,
				URL: "/api/search" + strings.TrimPrefix(suggestion.URL, "/search"),
			})
		}
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h apiHandler) stats(w http.ResponseWriter, r *http.Request) {
	if err := validateAPIKeys(r.URL.Query(), "lang"); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if _, err := apiLanguage(r.URL.Query().Get("lang")); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	data, meta, err := h.snapshot(r.Context())
	if err != nil {
		h.writeError(w, http.StatusServiceUnavailable, "prices_unavailable", "Fuel price data is unavailable.")
		return
	}
	summary := pricesummary.Build(data, time.Now())
	response := apiStatsResponse{
		State: "ok", Meta: meta, DataIsToday: summary.Today,
		NationalPrices: map[string]apiPriceStats{
			"gasolina95": priceAPIStats(summary.Gasoline95), "gasolina98": priceAPIStats(summary.Gasoline98),
			"gasoleo": priceAPIStats(summary.Diesel), "gasoleoPremium": priceAPIStats(summary.PremiumDiesel),
		},
		ProvinceAverages: map[string]apiProvinceRanking{
			"gasolina95": provinceAPIRanking(summary.Provinces.Gasoline95), "gasolina98": provinceAPIRanking(summary.Provinces.Gasoline98),
			"gasoleo": provinceAPIRanking(summary.Provinces.Diesel), "gasoleoPremium": provinceAPIRanking(summary.Provinces.PremiumDiesel),
		},
	}
	if !summary.Available {
		response.State = "no_prices"
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h apiHandler) snapshot(ctx context.Context) (*api.GasStationList, apiMetadata, error) {
	data, err := h.storage.GetLastPrices(ctx)
	if err != nil {
		h.logger.Error("JSON price snapshot unavailable", "error", err)
		return nil, apiMetadata{}, err
	}
	meta := apiMetadata{APIVersion: jsonAPIVersion, Version: version.String(), GeneratedAt: time.Now().UTC(), PublishedAt: data.Fecha, PublicationTimezone: "Europe/Madrid"}
	date, err := h.storage.GetLastUpdateDate(ctx)
	if err != nil {
		h.logger.Error("JSON snapshot storage date unavailable", "error", err)
	} else if date != nil {
		meta.StorageDate = date.Format(time.DateOnly)
	}
	return data, meta, nil
}

func (h apiHandler) lookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, errLocationNotFound) {
		h.writeError(w, http.StatusNotFound, "location_not_found", "Location not found. Try adding the province.")
		return
	}
	h.logger.Error("JSON location lookup failed", "error", err)
	h.writeError(w, http.StatusServiceUnavailable, "geocoding_unavailable", "Location search is temporarily unavailable.")
}

func (h apiHandler) rateLimited(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.writeError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests. Wait before retrying.")
		return
	}
	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
}

func (h apiHandler) writeError(w http.ResponseWriter, status int, state, message string) {
	h.writeJSON(w, status, apiErrorResponse{APIVersion: jsonAPIVersion, State: state, Error: message})
}

func (h apiHandler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		h.logger.Error("Writing JSON response failed", "error", err)
	}
}

func apiStations(stations []*api.GasStation, query apiSearchQuery) []apiStation {
	result := make([]apiStation, 0, len(stations))
	for _, station := range stations {
		lat, lng, ok := search.StationCoordinates(station)
		if !ok {
			continue
		}
		distance := gpx.Distance2D(query.Lat, query.Lng, lat, lng, true) / 1000
		if distance > query.RadiusKM {
			continue
		}
		latStr, lngStr := strconv.FormatFloat(lat, 'f', -1, 64), strconv.FormatFloat(lng, 'f', -1, 64)
		result = append(result, apiStation{
			ID: station.IDEESS, Name: station.Rotulo, Address: station.Direccion,
			Locality: station.Localidad, Municipality: station.Municipio, Province: station.Provincia,
			PostalCode: station.CP, Hours: station.Horario, Lat: lat, Lng: lng, DistanceKM: distance,
			Prices: stationAPIPrices(station),
			Maps:   apiMapLinks{OSM: fmt.Sprintf("https://www.openstreetmap.org/?mlat=%s&mlon=%s&zoom=16", latStr, lngStr), GoogleMaps: fmt.Sprintf("https://www.google.com/maps?q=%s,%s", latStr, lngStr)},
		})
	}
	sort.SliceStable(result, func(i, j int) bool { return apiStationLess(result[i], result[j], query.Fuel) })
	return result
}

func apiStationLess(a, b apiStation, fuel string) bool {
	priceA, priceB := a.Prices.forFuel(fuel), b.Prices.forFuel(fuel)
	if (priceA == nil) != (priceB == nil) {
		return priceA != nil
	}
	if priceA != nil && *priceA != *priceB {
		return *priceA < *priceB
	}
	return a.DistanceKM < b.DistanceKM
}
