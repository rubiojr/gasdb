package search

import (
	"math"
	"net/url"
	"testing"

	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestRadiiCountsAndPreservesQuery(t *testing.T) {
	stations := []api.GasStation{
		{Latitud: "0,06", Longitud: "0"}, // roughly 6.7 km from the origin
		{Latitud: "0.08", Longitud: "0"}, // roughly 8.9 km
		{Latitud: "0.18", Longitud: "0"}, // roughly 20 km
		{Latitud: "0.40", Longitud: "0"}, // roughly 44 km
		{Latitud: "1", Longitud: "0"},    // outside all offered radii
		{Latitud: "invalid", Longitud: "0"},
		{Latitud: "0", Longitud: "invalid"},
		{Latitud: "NaN", Longitud: "0"},
		{Latitud: "0", Longitud: "+Inf"},
		{Latitud: "360", Longitud: "0"},
	}
	query := url.Values{
		"location": {"L'Hospitalet & Barcelona"},
		"radius":   {"5"}, "fuel": {"gasoleo"}, "lang": {"en"},
		"lat": {"0"}, "lng": {"0"}, "extra": {"one", "two"},
	}
	original := query.Encode()
	suggestions := SuggestRadii(stations, 0, 0, 5, query)
	require.Len(t, suggestions, 3)
	for i, radius := range []int{10, 25, 50} {
		assert.Equal(t, radius, suggestions[i].Radius)
		assert.Equal(t, i+2, suggestions[i].Count)
		link, err := url.Parse(suggestions[i].URL)
		require.NoError(t, err)
		assert.Empty(t, link.Scheme)
		assert.Empty(t, link.Host)
		assert.Equal(t, "/search", link.Path)
		for _, key := range []string{"location", "fuel", "lang", "lat", "lng", "extra"} {
			assert.Equal(t, query[key], link.Query()[key])
		}
		assert.NotEqual(t, "5", link.Query().Get("radius"))
	}
	assert.Equal(t, original, query.Encode(), "suggestions must not modify the current request")
}

func TestStationCoordinates(t *testing.T) {
	lat, lng, ok := StationCoordinates(&api.GasStation{Latitud: "41,4", Longitud: "-2,3"})
	assert.True(t, ok)
	assert.Equal(t, 41.4, lat)
	assert.Equal(t, -2.3, lng)
	for _, station := range []*api.GasStation{
		nil,
		{Latitud: "91", Longitud: "0"},
		{Latitud: "-91", Longitud: "0"},
		{Latitud: "0", Longitud: "181"},
		{Latitud: "0", Longitud: "-181"},
	} {
		_, _, ok := StationCoordinates(station)
		assert.False(t, ok)
	}
}

func TestSuggestRadiiSkipsEmptyAndRedundantOptions(t *testing.T) {
	stations := []api.GasStation{{Latitud: "0.18", Longitud: "0"}}
	suggestions := SuggestRadii(stations, 0, 0, 5, nil)
	assert.Equal(t, []RadiusSuggestion{{Radius: 25, Count: 1, URL: "/search?radius=25"}}, suggestions)
	assert.Empty(t, SuggestRadii(nil, 0, 0, 5, nil))
	assert.Empty(t, SuggestRadii([]api.GasStation{{Latitud: "1", Longitud: "0"}}, 0, 0, 5, nil))
}

func TestSuggestRadiiOnlyOffersWiderBoundedSearches(t *testing.T) {
	stations := []api.GasStation{{Latitud: "0.4", Longitud: "0"}}
	for _, current := range []float64{5, 10, 25, 49.9} {
		assert.Equal(t, []RadiusSuggestion{{Radius: 50, Count: 1, URL: "/search?radius=50"}}, SuggestRadii(stations, 0, 0, current, nil))
	}
	for _, current := range []float64{0, -1, 50, 100, math.NaN(), math.Inf(1)} {
		assert.Empty(t, SuggestRadii(stations, 0, 0, current, nil))
	}
}
