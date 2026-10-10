package main

import (
	"fmt"
	"net/url"

	"github.com/rubiojr/gasdb/_server/internal/search"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/pkg/api"
)

// DefaultMapStyle is the light OpenFreeMap style the results map uses.
const DefaultMapStyle = "https://tiles.openfreemap.org/styles/positron"

func validateMapStyle(style string) error {
	u, err := url.Parse(style)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("invalid map style URL %q: want an http(s) URL", style)
	}
	return nil
}

// newResultsMap describes the search results for the results page map,
// keeping their order so markers match the station cards.
func newResultsMap(stations []api.StationWithDistance, lat, lng, radius float64, fuelType, style string) templates.ResultsMap {
	m := templates.ResultsMap{
		Style:    style,
		Center:   [2]float64{lng, lat},
		RadiusKM: radius,
		Stations: make([]templates.MapStation, 0, len(stations)),
	}
	for _, s := range stations {
		// Results only include stations with valid coordinates.
		stationLat, stationLng, _ := search.StationCoordinates(s.Station)
		station := templates.MapStation{Lng: stationLng, Lat: stationLat, Name: s.Station.Rotulo}
		if price := getFuelPrice(s.Station, fuelType); price > 0 {
			station.Price = &price
		}
		m.Stations = append(m.Stations, station)
	}
	return m
}
