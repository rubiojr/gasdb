package search

import (
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
)

// StationCoordinates applies the same validity rules to HTML results, JSON
// results, and wider-radius counts. Comparisons also reject NaN and infinities.
func StationCoordinates(station *api.GasStation) (float64, float64, bool) {
	if station == nil {
		return 0, 0, false
	}
	lat, err := gasdb.ParseLatLong(station.Latitud)
	if err != nil {
		return 0, 0, false
	}
	lng, err := gasdb.ParseLatLong(station.Longitud)
	if err != nil {
		return 0, 0, false
	}
	return lat, lng, lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
