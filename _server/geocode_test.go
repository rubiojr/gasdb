package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testGeocoder(t *testing.T, handler http.HandlerFunc) *locationGeocoder {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &locationGeocoder{client: server.Client(), endpoint: server.URL + "/search", cache: cache.New(time.Minute, 0)}
}

func TestLocationLookup(t *testing.T) {
	for _, location := range []string{"masnou", "el Masnou", "A Coruña", "L'Hospitalet de Llobregat", "Calle A & B, Madrid"} {
		t.Run(location, func(t *testing.T) {
			calls := 0
			g := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				assert.Equal(t, "/search", r.URL.Path)
				assert.Equal(t, location, r.URL.Query().Get("q"))
				assert.Equal(t, "city", r.URL.Query().Get("featureType"))
				assert.Equal(t, "es", r.URL.Query().Get("countrycodes"))
				assert.Equal(t, "1", r.URL.Query().Get("limit"))
				assert.Equal(t, "1", r.URL.Query().Get("addressdetails"))
				assert.Contains(t, r.UserAgent(), "gasdb")
				fmt.Fprint(w, `[{"lat":"41.4796899","lon":"2.3118347","display_name":"el Masnou, Maresme, Barcelona, Catalunya, España","address":{"town":"el Masnou","county":"Maresme","state_district":"Barcelona"}}]`)
			})
			place, err := g.lookup(context.Background(), "  "+location+"  ")
			require.NoError(t, err)
			assert.Equal(t, 41.4796899, place.lat)
			assert.Equal(t, 2.3118347, place.lng)
			assert.Equal(t, "el Masnou, Barcelona", place.name)
			cachedPlace, err := g.lookup(context.Background(), strings.ToUpper(location))
			require.NoError(t, err)
			assert.Equal(t, place, cachedPlace)
			assert.Equal(t, 1, calls)
		})
	}
}

func TestLocationAddressFallback(t *testing.T) {
	var features []string
	g := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
		feature := r.URL.Query().Get("featureType")
		features = append(features, feature)
		if feature == "city" {
			fmt.Fprint(w, `[]`)
			return
		}
		fmt.Fprint(w, `[{"lat":"40.4","lon":"-3.7","address":{"city":"Madrid","province":"Madrid"}}]`)
	})
	place, err := g.lookup(context.Background(), "Calle Mayor, Madrid")
	require.NoError(t, err)
	assert.Equal(t, 40.4, place.lat)
	assert.Equal(t, -3.7, place.lng)
	assert.Equal(t, "Madrid", place.name)
	assert.Equal(t, []string{"city", ""}, features)
}

func TestLocationPrefersCityOverProvince(t *testing.T) {
	g := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "soria", r.URL.Query().Get("q"))
		if r.URL.Query().Get("featureType") == "city" {
			fmt.Fprint(w, `[{"lat":"41.7633842","lon":"-2.4642041","address":{"city":"Soria","state_district":"Soria"}}]`)
			return
		}
		// The former settlement search returned this provincial centre, about
		// 28 km away from the city where the user expects to find stations.
		fmt.Fprint(w, `[{"lat":"41.6012505","lon":"-2.7219380","address":{"state_district":"Soria"}}]`)
	})
	place, err := g.lookup(context.Background(), "soria")
	require.NoError(t, err)
	assert.Equal(t, 41.7633842, place.lat)
	assert.Equal(t, -2.4642041, place.lng)
	assert.Equal(t, "Soria", place.name)
}

func TestLocationFailures(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		notFound bool
	}{
		{"no matches", 200, `[]`, true},
		{"rate limited", 429, `[]`, false},
		{"service failure", 503, `unavailable`, false},
		{"invalid JSON", 200, `<html>blocked</html>`, false},
		{"invalid latitude", 200, `[{"lat":"oops","lon":"2"}]`, false},
		{"invalid longitude", 200, `[{"lat":"41","lon":"181"}]`, false},
		{"nonfinite coordinate", 200, `[{"lat":"NaN","lon":"2"}]`, false},
		{"oversized response", 200, strings.Repeat(" ", (1<<20)+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			g := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			})
			_, err := g.lookup(context.Background(), "masnou")
			require.Error(t, err)
			if tt.notFound {
				assert.ErrorIs(t, err, errLocationNotFound)
				assert.Equal(t, 2, calls)
			} else {
				assert.NotErrorIs(t, err, errLocationNotFound)
				assert.Equal(t, 1, calls)
			}
			assert.Zero(t, g.cache.ItemCount(), "failures must not be cached")
		})
	}
}

func TestLocationCancelledAndBlank(t *testing.T) {
	g := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected geocoding request")
	})
	_, err := g.lookup(context.Background(), " \t ")
	assert.ErrorIs(t, err, errLocationNotFound)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = g.lookup(ctx, "masnou")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestResolvedPlaceNames(t *testing.T) {
	tests := []struct {
		name, fields, want string
	}{
		{"town", `"address":{"town":"el Masnou","county":"Maresme","state_district":"Barcelona"}`, "el Masnou, Barcelona"},
		{"city", `"address":{"city":"A Coruña","province":"A Coruña"}`, "A Coruña"},
		{"village", `"address":{"village":"Almarza","province":"Soria"}`, "Almarza, Soria"},
		{"municipality", `"address":{"municipality":"San Pedro Manrique","state_district":"Soria"}`, "San Pedro Manrique, Soria"},
		{"hamlet", `"address":{"hamlet":"Valdeavellano","province":"Soria"}`, "Valdeavellano, Soria"},
		{"explicit province", `"address":{"town":"el Masnou","province":"Barcelona","state_district":"Other"}`, "el Masnou, Barcelona"},
		{"county is not province", `"display_name":"el Masnou, Maresme, Barcelona, España","address":{"town":"el Masnou","county":"Maresme"}`, "el Masnou, Maresme, Barcelona, España"},
		{"full name fallback", `"display_name":"Calle Mayor, Madrid, España"`, "Calle Mayor, Madrid, España"},
		{"city only", `"address":{"city":"Madrid"}`, "Madrid"},
		{"province only", `"address":{"province":"Soria"}`, "Soria"},
		{"feature name", `"name":"Tibidabo"`, "Tibidabo"},
		{"no name", `"address":{}`, ""},
		{"whitespace", `"address":{"town":" el Masnou ","state_district":" Barcelona "}`, "el Masnou, Barcelona"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`[{"lat":"41.4","lon":"2.3",%s}]`, tt.fields)
			place, err := decodeLocation(strings.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, tt.want, place.name)
		})
	}
}

func TestResolvedLocationRendering(t *testing.T) {
	for _, lang := range []string{"es", "en"} {
		t.Run(lang, func(t *testing.T) {
			var output strings.Builder
			tr := translations.GetTranslations(lang)
			err := templates.ResultsPage(nil, "masnou", "el Masnou, Barcelona", 41.4797, 2.3118, 5, nil, nil, tr).Render(context.Background(), &output)
			require.NoError(t, err)
			assert.Contains(t, output.String(), ">masnou</strong>")
			assert.Contains(t, output.String(), tr.ResolvedLocation)
			assert.Contains(t, output.String(), ">el Masnou, Barcelona</strong>")
			assert.Contains(t, output.String(), `href="/?lang=`+lang+`"`)
		})
	}
	var output strings.Builder
	tr := translations.GetEnglishTranslations()
	require.NoError(t, templates.ResultsPage(nil, "query", "A & B <Village>", 41, 2, 5, nil, nil, tr).Render(context.Background(), &output))
	assert.Contains(t, output.String(), "A &amp; B &lt;Village&gt;")
	output.Reset()
	require.NoError(t, templates.ResultsPage(nil, "", "", 41, 2, 5, nil, nil, tr).Render(context.Background(), &output))
	assert.NotContains(t, output.String(), `class="results-location resolved-location"`)
}
