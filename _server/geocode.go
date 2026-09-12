package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var errLocationNotFound = errors.New("location not found")

type locationGeocoder struct {
	client      *http.Client
	endpoint    string
	cache       *cache.Cache
	interval    time.Duration
	mu          sync.Mutex
	nextRequest time.Time
}

type coordinates struct {
	lat, lng float64
}

func (g *locationGeocoder) lookup(ctx context.Context, location string) (float64, float64, error) {
	location = strings.Join(strings.Fields(location), " ")
	if location == "" {
		return 0, 0, errLocationNotFound
	}
	key := strings.ToLower(location)
	if cached, ok := g.cache.Get(key); ok {
		point := cached.(coordinates)
		return point.lat, point.lng, nil
	}

	// Prefer towns over similarly named buildings; retain address searches.
	for _, featureType := range []string{"settlement", ""} {
		point, err := g.search(ctx, location, featureType)
		if errors.Is(err, errLocationNotFound) {
			continue
		}
		if err != nil {
			return 0, 0, err
		}
		g.cache.Set(key, point, cache.DefaultExpiration)
		return point.lat, point.lng, nil
	}
	return 0, 0, errLocationNotFound
}

func (g *locationGeocoder) search(ctx context.Context, location, featureType string) (coordinates, error) {
	query := url.Values{"q": {location}, "format": {"jsonv2"}, "countrycodes": {"es"}, "limit": {"1"}}
	if featureType != "" {
		query.Set("featureType", featureType)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return coordinates{}, err
	}
	req.Header.Set("User-Agent", "gasdb/1.0 (https://github.com/rubiojr/gasdb)")
	if err := g.wait(ctx); err != nil {
		return coordinates{}, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return coordinates{}, fmt.Errorf("geocoding request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return coordinates{}, fmt.Errorf("geocoding service returned HTTP %d", resp.StatusCode)
	}
	return decodeLocation(resp.Body)
}

// Nominatim's public service permits at most one request per second.
func (g *locationGeocoder) wait(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	timer := time.NewTimer(max(0, time.Until(g.nextRequest)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		g.nextRequest = time.Now().Add(g.interval)
		return nil
	}
}

func decodeLocation(body io.Reader) (coordinates, error) {
	const maxResponseSize = 1 << 20
	data, err := io.ReadAll(io.LimitReader(body, maxResponseSize+1))
	if err != nil {
		return coordinates{}, err
	}
	if len(data) > maxResponseSize {
		return coordinates{}, errors.New("geocoding response too large")
	}
	var results []struct{ Lat, Lon string }
	if err := json.Unmarshal(data, &results); err != nil {
		return coordinates{}, fmt.Errorf("decoding geocoding response: %w", err)
	}
	if len(results) == 0 {
		return coordinates{}, errLocationNotFound
	}
	lat, err := parseCoordinate(results[0].Lat, 90)
	if err != nil {
		return coordinates{}, err
	}
	lng, err := parseCoordinate(results[0].Lon, 180)
	if err != nil {
		return coordinates{}, err
	}
	return coordinates{lat: lat, lng: lng}, nil
}

func parseCoordinate(value string, limit float64) (float64, error) {
	n, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) > limit {
		return 0, fmt.Errorf("invalid geocoding coordinate %q", value)
	}
	return n, nil
}
