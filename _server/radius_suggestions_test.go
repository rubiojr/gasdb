package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/rubiojr/gasdb/_server/internal/search"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRadiusSuggestionsRendering(t *testing.T) {
	suggestions := []search.RadiusSuggestion{
		{Radius: 10, Count: 1, URL: "/search?location=Duruelo&radius=10"},
		{Radius: 25, Count: 7, URL: "/search?location=Duruelo&radius=25"},
	}
	for _, lang := range []string{"es", "en"} {
		t.Run(lang, func(t *testing.T) {
			tr := translations.GetTranslations(lang)
			var output strings.Builder
			err := templates.ResultsPage(nil, "Duruelo", "Duruelo de la Sierra, Soria", 41.954, -2.932, 5, nil, suggestions, tr).Render(context.Background(), &output)
			require.NoError(t, err)
			assert.Contains(t, output.String(), tr.ExpandRadiusTitle)
			assert.Contains(t, output.String(), fmt.Sprintf(tr.SearchWithinRadius, 10))
			assert.Contains(t, output.String(), tr.OneStation)
			assert.Contains(t, output.String(), fmt.Sprintf(tr.StationCount, 7))
			assert.Contains(t, output.String(), `href="/search?location=Duruelo&amp;radius=25"`)
		})
	}
}

func TestRadiusSuggestionsOnlyAppearForEmptySuccessfulSearches(t *testing.T) {
	suggestion := []search.RadiusSuggestion{{Radius: 10, Count: 1, URL: "/search?radius=10"}}
	station := api.StationWithDistance{Station: &api.GasStation{Rotulo: "Example", Latitud: "41", Longitud: "2"}}
	for _, tt := range []struct {
		name        string
		stations    []api.StationWithDistance
		err         error
		suggestions []search.RadiusSuggestion
	}{
		{"lookup failed", nil, errors.New("service unavailable"), suggestion},
		{"stations already found", []api.StationWithDistance{station}, nil, suggestion},
		{"no wider matches", nil, nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			err := templates.ResultsPage(tt.stations, "query", "", 41, 2, 5, tt.err, tt.suggestions, translations.GetEnglishTranslations()).Render(context.Background(), &output)
			require.NoError(t, err)
			assert.NotContains(t, output.String(), `class="radius-suggestions"`)
		})
	}
}
