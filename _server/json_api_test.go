package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/httprate"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func apiFixture(t *testing.T, seed bool) (apiHandler, *api.GasStationList) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	storage, err := gasdb.NewStorage(context.Background(), filepath.Join(t.TempDir(), "prices.db"), logger)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	geocoder := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("q") {
		case "missing":
			fmt.Fprint(w, `[]`)
		case "down":
			w.WriteHeader(429)
		default:
			fmt.Fprint(w, `[{"lat":"0","lon":"0","address":{"town":"Testville","state_district":"Testshire"}}]`)
		}
	})
	data := &api.GasStationList{Fecha: "01/01/2020 0:00:00", ListaEESSPrecio: []api.GasStation{
		{IDEESS: "gasoline-cheap", Rotulo: "Gasoline cheap", Provincia: "Alpha", Latitud: "0", Longitud: "0", PrecioGasolina95E5: "1,100", PrecioGasolina98E5: "NaN", PrecioGasoleoA: "1,900", PrecioGasoleoPremium: "1,900"},
		{IDEESS: "premium-far", Rotulo: "Premium far", Provincia: "Beta", Latitud: "0,02", Longitud: "0", PrecioGasolina95E5: "1,700", PrecioGasoleoA: "1,600", PrecioGasoleoPremium: "1,200"},
		{IDEESS: "premium-near", Rotulo: "Premium near", Provincia: "Gamma", Latitud: "0,01", Longitud: "0", PrecioGasolina95E5: "1,800", PrecioGasoleoA: "1,500", PrecioGasoleoPremium: "1,200"},
		{IDEESS: "missing", Rotulo: "No premium", Provincia: "Alpha", Latitud: "0,001", Longitud: "0", PrecioGasolina95E5: "1,000", PrecioGasoleoPremium: "0"},
		{IDEESS: "invalid", Rotulo: "Invalid coordinates", Provincia: "Alpha", Latitud: "360", Longitud: "0", PrecioGasolina95E5: "2,100"},
		{IDEESS: "outside", Rotulo: "Outside", Provincia: "Beta", Latitud: "0,18", Longitud: "0", PrecioGasolina95E5: "1,800"},
	}}
	if seed {
		saveAPIFixture(t, storage, data)
	}
	return apiHandler{storage: storage, geocoder: geocoder, logger: logger}, data
}

func saveAPIFixture(t *testing.T, storage *gasdb.Storage, data *api.GasStationList) {
	t.Helper()
	encoded, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, storage.SavePrices(context.Background(), time.Now(), encoded))
}

func apiCall(t *testing.T, handler http.HandlerFunc, path string, response any) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest("GET", path, nil))
	require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), response), "%s", w.Body.String())
	return w
}

func TestJSONSearchSortsBeforeLimiting(t *testing.T) {
	h, _ := apiFixture(t, true)
	var response apiSearchResponse
	w := apiCall(t, h.search, "/api/search?lat=0&lng=0&fuel=gasoleoPremium&limit=1&lang=en", &response)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "ok", response.State)
	assert.Equal(t, 4, response.TotalStations)
	assert.Equal(t, 1, response.ReturnedStations)
	assert.True(t, response.Truncated)
	require.Len(t, response.Stations, 1)
	assert.Equal(t, "premium-near", response.Stations[0].ID)
	require.NotNil(t, response.Stations[0].Prices.PremiumDiesel)
	assert.Equal(t, 1.2, *response.Stations[0].Prices.PremiumDiesel)
	assert.InDelta(t, 1.11, response.Stations[0].DistanceKM, 0.01)
	assert.Contains(t, response.Stations[0].Maps.OSM, "mlat=0.01")
	assert.Contains(t, response.WebURL, "fuel=gasoleoPremium")
	assert.NotContains(t, response.WebURL, "limit=")
	assert.Equal(t, "01/01/2020 0:00:00", response.Meta.PublishedAt)
	assert.Equal(t, "Europe/Madrid", response.Meta.PublicationTimezone)
	assert.False(t, response.Meta.GeneratedAt.IsZero())
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	apiCall(t, h.search, "/api/search?lat=0&lng=0&fuel=gasoleoPremium&limit=100", &response)
	require.Len(t, response.Stations, 4)
	assert.Nil(t, response.Stations[3].Prices.PremiumDiesel)
	assert.False(t, response.Truncated)
	assert.Empty(t, response.RadiusSuggestions)
	for _, tt := range []struct{ fuel, firstID string }{
		{"gasolina95", "missing"},
		{"gasolina98", "gasoline-cheap"},
		{"gasoleo", "premium-near"},
		{"GASOLEOPREMIUM", "premium-near"},
	} {
		apiCall(t, h.search, "/api/search?lat=0&lng=0&fuel="+tt.fuel, &response)
		require.NotEmpty(t, response.Stations)
		assert.Equal(t, tt.firstID, response.Stations[0].ID)
	}
}

func TestAPIStationsRejectInvalidAndOutsideCoordinates(t *testing.T) {
	stations := []*api.GasStation{
		nil,
		{Latitud: "NaN", Longitud: "0"},
		{Latitud: "0", Longitud: "Inf"},
		{Latitud: "0.4", Longitud: "0"},
		{IDEESS: "valid", Latitud: "0.01", Longitud: "0"},
	}
	result := apiStations(stations, apiSearchQuery{RadiusKM: 5, Fuel: "gasolina95"})
	require.Len(t, result, 1)
	assert.Equal(t, "valid", result[0].ID)
}

func TestJSONSearchNamesAndErrors(t *testing.T) {
	h, _ := apiFixture(t, true)
	var response apiSearchResponse
	w := apiCall(t, h.search, "/api/search?location=Testville&lat=invalid&lng=invalid", &response)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "Testville", response.Query.Location)
	assert.Equal(t, "Testville, Testshire", response.Query.ResolvedLocation)
	assert.Equal(t, 4, response.TotalStations)
	for _, tt := range []struct {
		query, state string
		status       int
	}{
		{"missing", "location_not_found", 404},
		{"down", "geocoding_unavailable", 503},
	} {
		var failure apiErrorResponse
		w := apiCall(t, h.search, "/api/search?location="+tt.query, &failure)
		assert.Equal(t, tt.status, w.Code)
		assert.Equal(t, tt.state, failure.State)
		assert.NotEmpty(t, failure.Error)
		assert.NotContains(t, w.Body.String(), "radius_suggestions")
	}
}

func TestJSONRadiusSuggestionsMatchTotalCounts(t *testing.T) {
	h, _ := apiFixture(t, true)
	var response apiSearchResponse
	w := apiCall(t, h.search, "/api/search?lat=-0.06&lng=0&radius=5&limit=1&fuel=gasoleo&lang=en", &response)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "no_stations", response.State)
	assert.Empty(t, response.Stations)
	assert.Contains(t, w.Body.String(), `"stations":[]`)
	require.Len(t, response.RadiusSuggestions, 2)
	assert.Equal(t, 4, response.RadiusSuggestions[0].StationCount)
	for _, suggestion := range response.RadiusSuggestions {
		assert.True(t, strings.HasPrefix(suggestion.URL, "/api/search?"))
		var expanded apiSearchResponse
		w := apiCall(t, h.search, suggestion.URL, &expanded)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, suggestion.StationCount, expanded.TotalStations)
		assert.Equal(t, 1, expanded.ReturnedStations)
		assert.Equal(t, "gasoleo", expanded.Query.Fuel)
		assert.Equal(t, "en", expanded.Query.Language)
	}
}

func TestJSONStatsAndSnapshotRefresh(t *testing.T) {
	h, data := apiFixture(t, true)
	var response apiStatsResponse
	w := apiCall(t, h.stats, "/api/stats", &response)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "ok", response.State)
	assert.False(t, response.DataIsToday)
	assert.Equal(t, 6, response.NationalPrices["gasolina95"].StationCount)
	require.NotNil(t, response.NationalPrices["gasolina95"].StationMin)
	assert.Equal(t, 1.0, *response.NationalPrices["gasolina95"].StationMin)
	assert.Nil(t, response.NationalPrices["gasolina98"].StationMin)
	assert.Equal(t, 0, response.NationalPrices["gasolina98"].StationCount)
	assert.Contains(t, w.Body.String(), `"lowest":[],"highest":[]`)
	assert.Len(t, response.ProvinceAverages["gasolina95"].Lowest, 2)
	data.ListaEESSPrecio[3].PrecioGasolina95E5 = "2,200"
	saveAPIFixture(t, h.storage, data)
	apiCall(t, h.stats, "/api/stats", &response)
	assert.Equal(t, 1.1, *response.NationalPrices["gasolina95"].StationMin)
}

func TestJSONUnavailableAndEmptyData(t *testing.T) {
	h, _ := apiFixture(t, false)
	for _, tt := range []struct {
		handler http.HandlerFunc
		path    string
	}{{h.stats, "/api/stats"}, {h.search, "/api/search?lat=0&lng=0"}} {
		var failure apiErrorResponse
		w := apiCall(t, tt.handler, tt.path, &failure)
		assert.Equal(t, 503, w.Code)
		assert.Equal(t, "prices_unavailable", failure.State)
	}
	saveAPIFixture(t, h.storage, &api.GasStationList{})
	var response apiStatsResponse
	w := apiCall(t, h.stats, "/api/stats", &response)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "no_prices", response.State)
}

func TestJSONValidation(t *testing.T) {
	h := apiHandler{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, query := range []string{
		"", "lat=41", "lat=91&lng=2", "lat=41&lng=181", "lat=NaN&lng=2", "lat=41&lng=Inf",
		"location=soria&radius=0", "location=soria&radius=-1", "location=soria&radius=51",
		"location=soria&radius=NaN", "location=soria&radius=Inf", "location=soria&radius=no",
		"location=soria&limit=0", "location=soria&limit=101", "location=soria&limit=1.5",
		"location=soria&fuel=diesel", "location=soria&lang=fr", "location=%FF",
		"location=soria&raduis=10", "location=soria&radius=5&radius=10",
		"location=" + strings.Repeat("a", 513),
	} {
		t.Run(query, func(t *testing.T) {
			var failure apiErrorResponse
			w := apiCall(t, h.search, "/api/search?"+query, &failure)
			assert.Equal(t, 400, w.Code)
			assert.Equal(t, "invalid_request", failure.State)
		})
	}
	query, err := parseAPIQuery(url.Values{"lat": {"41.4"}, "lng": {"2.3"}})
	require.NoError(t, err)
	assert.Equal(t, 41.4, query.Lat)
	assert.Equal(t, 2.3, query.Lng)
	assert.Equal(t, 5.0, query.RadiusKM)
	assert.Equal(t, 20, query.Limit)
	assert.Equal(t, "gasolina95", query.Fuel)
	for _, query := range []string{"province=Soria", "lang=fr", "lang=en&lang=es"} {
		var failure apiErrorResponse
		w := apiCall(t, h.stats, "/api/stats?"+query, &failure)
		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "invalid_request", failure.State)
	}
}

func TestJSONRateLimitResponse(t *testing.T) {
	h := apiHandler{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	limiter := httprate.Limit(1, time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP), httprate.WithLimitHandler(h.rateLimited))
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/api/stats", nil))
	assert.Equal(t, 429, w.Code)
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
	var failure apiErrorResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &failure))
	assert.Equal(t, "rate_limited", failure.State)
}

func TestHTMLDieselSortAliases(t *testing.T) {
	station := &api.GasStation{PrecioGasolina95E5: "1,900", PrecioGasoleoA: "1,100", PrecioGasoleoB: "0,900", PrecioGasoleoPremium: "1,200"}
	assert.Equal(t, 1.1, getFuelPrice(station, "gasoleoA"))
	assert.Equal(t, 0.9, getFuelPrice(station, "gasoleoB"))
	assert.Equal(t, 1.2, getFuelPrice(station, "gasoleoPremium"))
}
