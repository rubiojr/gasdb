package main

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/rubiojr/gasdb/_server/internal/search"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
)

// renderStation shows the shareable page of the station with the given ID,
// using the latest price snapshot.
func renderStation(w http.ResponseWriter, r *http.Request, storage *gasdb.Storage, logger *slog.Logger, id, mapStyle string, t translations.Translations) {
	render := func(status int, page templ.Component) {
		w.WriteHeader(status)
		if err := page.Render(r.Context(), w); err != nil {
			logger.Error("Error rendering station page", "error", err)
		}
	}
	// GetLastPrices caches the decoded snapshot and invalidates it on updates.
	prices, err := storage.GetLastPrices(r.Context())
	if err != nil {
		logger.Error("Error getting prices for station page", "error", err)
		render(http.StatusServiceUnavailable, templates.StationMissingPage(t.HomeTitle, t.PricesUnavailable, t))
		return
	}
	station := findStation(prices.ListaEESSPrecio, id)
	// Search results skip stations without valid coordinates, so their pages
	// can't be shared either.
	lat, lng, ok := search.StationCoordinates(station)
	if !ok {
		render(http.StatusNotFound, templates.StationMissingPage(t.StationNotFoundTitle, t.StationNotFound, t))
		return
	}
	stationMap := templates.StationMap{Style: mapStyle, Lng: lng, Lat: lat}
	render(http.StatusOK, templates.StationPage(station, prices.Fecha, stationMap, t))
}

func findStation(stations []api.GasStation, id string) *api.GasStation {
	for i := range stations {
		if stations[i].IDEESS == id {
			return &stations[i]
		}
	}
	return nil
}
