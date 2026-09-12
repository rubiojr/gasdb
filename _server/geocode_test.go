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
				assert.Equal(t, "settlement", r.URL.Query().Get("featureType"))
				assert.Equal(t, "es", r.URL.Query().Get("countrycodes"))
				assert.Equal(t, "1", r.URL.Query().Get("limit"))
				assert.Contains(t, r.UserAgent(), "gasdb")
				fmt.Fprint(w, `[{"lat":"41.4796899","lon":"2.3118347"}]`)
			})
			lat, lng, err := g.lookup(context.Background(), "  "+location+"  ")
			require.NoError(t, err)
			assert.Equal(t, 41.4796899, lat)
			assert.Equal(t, 2.3118347, lng)
			cachedLat, cachedLng, err := g.lookup(context.Background(), strings.ToUpper(location))
			require.NoError(t, err)
			assert.Equal(t, lat, cachedLat)
			assert.Equal(t, lng, cachedLng)
			assert.Equal(t, 1, calls)
		})
	}
}

func TestLocationAddressFallback(t *testing.T) {
	var features []string
	g := testGeocoder(t, func(w http.ResponseWriter, r *http.Request) {
		feature := r.URL.Query().Get("featureType")
		features = append(features, feature)
		if feature == "settlement" {
			fmt.Fprint(w, `[]`)
			return
		}
		fmt.Fprint(w, `[{"lat":"40.4","lon":"-3.7"}]`)
	})
	lat, lng, err := g.lookup(context.Background(), "Calle Mayor, Madrid")
	require.NoError(t, err)
	assert.Equal(t, 40.4, lat)
	assert.Equal(t, -3.7, lng)
	assert.Equal(t, []string{"settlement", ""}, features)
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
			_, _, err := g.lookup(context.Background(), "masnou")
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
	_, _, err := g.lookup(context.Background(), " \t ")
	assert.ErrorIs(t, err, errLocationNotFound)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = g.lookup(ctx, "masnou")
	assert.ErrorIs(t, err, context.Canceled)
}
