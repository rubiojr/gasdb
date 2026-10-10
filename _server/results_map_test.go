package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMapStyle(t *testing.T) {
	assert.NoError(t, validateMapStyle(DefaultMapStyle))
	assert.NoError(t, validateMapStyle("http://127.0.0.1:8080/maps/styles/positron"))
	for _, style := range []string{"", "tiles.example.com/styles/positron", "ftp://tiles.example.com", "https://", "javascript:alert(1)"} {
		assert.Error(t, validateMapStyle(style), style)
	}
}

func TestNewResultsMap(t *testing.T) {
	stations := []api.StationWithDistance{
		{Station: &api.GasStation{Rotulo: "CHEAP", Latitud: "41,5", Longitud: "2,25", PrecioGasoleoA: "1,359"}},
		{Station: &api.GasStation{Rotulo: "NO DIESEL", Latitud: "41,4", Longitud: "2,1", PrecioGasolina95E5: "1,499"}},
	}
	m := newResultsMap(stations, 41.45, 2.2, 10, "gasoleoa", "https://maps.example/light")
	assert.Equal(t, "https://maps.example/light", m.Style)
	assert.Equal(t, [2]float64{2.2, 41.45}, m.Center, "MapLibre takes [lng, lat]")
	assert.Equal(t, 10.0, m.RadiusKM)
	require.Len(t, m.Stations, 2)
	assert.Equal(t, "CHEAP", m.Stations[0].Name)
	assert.Equal(t, 2.25, m.Stations[0].Lng)
	assert.Equal(t, 41.5, m.Stations[0].Lat)
	require.NotNil(t, m.Stations[0].Price)
	assert.Equal(t, 1.359, *m.Stations[0].Price)
	assert.Nil(t, m.Stations[1].Price, "a station without the searched fuel has no price")
}

func TestResultsMapRendering(t *testing.T) {
	stations := []api.StationWithDistance{
		{Station: &api.GasStation{Rotulo: "FIRST </script>", Latitud: "41,5", Longitud: "2,25", PrecioGasolina95E5: "1,459"}},
		{Station: &api.GasStation{Rotulo: "SECOND", Latitud: "41,4", Longitud: "2,1"}},
	}
	resultsMap := newResultsMap(stations, 41.45, 2.2, 5, "gasolina95", "https://maps.example/light")
	for _, lang := range []string{"es", "en"} {
		t.Run(lang, func(t *testing.T) {
			tr := translations.GetTranslations(lang)
			var output strings.Builder
			require.NoError(t, templates.ResultsPage(stations, "query", "", 41.45, 2.2, 5, nil, nil, resultsMap, tr).Render(context.Background(), &output))
			page := output.String()
			// Without JavaScript, the toggle stays hidden.
			assert.Contains(t, page, `aria-controls="results-map" aria-expanded="false" hidden>`+tr.ShowMapButton+`</button>`)
			assert.Contains(t, page, `id="results-map" class="results-map" aria-label="`+tr.MapLabel+`" hidden>`)
			assert.Contains(t, page, `id="station-0" class="card station-card"`)
			assert.Contains(t, page, `id="station-1" class="card station-card"`)
			assert.NotContains(t, page, "FIRST </script>", "station names are escaped, in the JSON too")

			var data templates.ResultsMap
			require.NoError(t, json.Unmarshal([]byte(jsonScript(t, page, "results-map-data")), &data))
			assert.Equal(t, resultsMap, data)
			var text map[string]any
			require.NoError(t, json.Unmarshal([]byte(jsonScript(t, page, "results-map-text")), &text))
			assert.Equal(t, tr.HideMapButton, text["hide"])
			assert.Equal(t, tr.NotAvailable, text["not_available"])
			if lang == "es" {
				assert.Equal(t, "Acercar", text["locale"].(map[string]any)["NavigationControl.ZoomIn"])
			} else {
				assert.NotContains(t, text, "locale", "English uses MapLibre's built-in strings")
			}
		})
	}
	var output strings.Builder
	require.NoError(t, templates.ResultsPage(nil, "query", "", 41, 2, 5, nil, nil, templates.ResultsMap{}, translations.GetEnglishTranslations()).Render(context.Background(), &output))
	assert.NotContains(t, output.String(), `id="results-map`, "empty results have no map")
	assert.NotContains(t, output.String(), "maplibre-gl", "empty results don't reference MapLibre")
}

// jsonScript returns the contents of the JSON script element with the given id.
func jsonScript(t *testing.T, page, id string) string {
	t.Helper()
	_, after, found := strings.Cut(page, `<script id="`+id+`" type="application/json">`)
	require.True(t, found, "missing %s", id)
	contents, _, found := strings.Cut(after, "</script>")
	require.True(t, found, "unterminated %s", id)
	return contents
}
