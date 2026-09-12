// Package search provides helpers for nearby station searches.
package search

import (
	"net/url"
	"strconv"

	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/tkrajina/gpxgo/gpx"
)

type RadiusSuggestion struct {
	Radius int // kilometres
	Count  int
	URL    string
}

var suggestedRadii = [...]int{5, 10, 25, 50}

// SuggestRadii counts stations in one pass, using the same distance calculation
// as the results page. It offers only wider radii with additional stations.
func SuggestRadii(stations []api.GasStation, lat, lng, currentRadius float64, query url.Values) []RadiusSuggestion {
	if !(currentRadius > 0 && currentRadius < 50) {
		return nil
	}
	counts := countRadii(stations, lat, lng)
	var suggestions []RadiusSuggestion
	previousCount := 0
	for i, radius := range suggestedRadii {
		if float64(radius) <= currentRadius || counts[i] <= previousCount {
			continue
		}
		suggestions = append(suggestions, RadiusSuggestion{
			Radius: radius,
			Count:  counts[i],
			URL:    radiusURL(query, radius),
		})
		previousCount = counts[i]
	}
	return suggestions
}

func countRadii(stations []api.GasStation, lat, lng float64) [len(suggestedRadii)]int {
	var counts [len(suggestedRadii)]int
	for i := range stations {
		station := &stations[i]
		stationLat, err := gasdb.ParseLatLong(station.Latitud)
		if err != nil {
			continue
		}
		stationLng, err := gasdb.ParseLatLong(station.Longitud)
		if err != nil {
			continue
		}
		distance := gpx.Distance2D(lat, lng, stationLat, stationLng, true)
		for i, radius := range suggestedRadii {
			if distance <= float64(radius)*1000 {
				counts[i]++
			}
		}
	}
	return counts
}

func radiusURL(query url.Values, radius int) string {
	values := make(url.Values, len(query)+1)
	for key, entries := range query {
		values[key] = append([]string(nil), entries...)
	}
	values.Set("radius", strconv.Itoa(radius))
	return "/search?" + values.Encode()
}
