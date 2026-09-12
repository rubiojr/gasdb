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

type geocodedLocation struct {
	lat, lng float64
	name     string
}

func (g *locationGeocoder) lookup(ctx context.Context, location string) (geocodedLocation, error) {
	location = strings.Join(strings.Fields(location), " ")
	if location == "" {
		return geocodedLocation{}, errLocationNotFound
	}
	key := strings.ToLower(location)
	if cached, ok := g.cache.Get(key); ok {
		return cached.(geocodedLocation), nil
	}

	// Prefer cities/towns over namesake provinces and buildings. Nominatim's
	// broader "settlement" filter includes provinces (for example, Soria).
	for _, featureType := range []string{"city", ""} {
		point, err := g.search(ctx, location, featureType)
		if errors.Is(err, errLocationNotFound) {
			continue
		}
		if err != nil {
			return geocodedLocation{}, err
		}
		g.cache.Set(key, point, cache.DefaultExpiration)
		return point, nil
	}
	return geocodedLocation{}, errLocationNotFound
}

func (g *locationGeocoder) search(ctx context.Context, location, featureType string) (geocodedLocation, error) {
	query := url.Values{"q": {location}, "format": {"jsonv2"}, "countrycodes": {"es"}, "limit": {"1"}, "addressdetails": {"1"}}
	if featureType != "" {
		query.Set("featureType", featureType)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return geocodedLocation{}, err
	}
	req.Header.Set("User-Agent", "gasdb/1.0 (https://github.com/rubiojr/gasdb)")
	if err := g.wait(ctx); err != nil {
		return geocodedLocation{}, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return geocodedLocation{}, fmt.Errorf("geocoding request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return geocodedLocation{}, fmt.Errorf("geocoding service returned HTTP %d", resp.StatusCode)
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

func decodeLocation(body io.Reader) (geocodedLocation, error) {
	const maxResponseSize = 1 << 20
	data, err := io.ReadAll(io.LimitReader(body, maxResponseSize+1))
	if err != nil {
		return geocodedLocation{}, err
	}
	if len(data) > maxResponseSize {
		return geocodedLocation{}, errors.New("geocoding response too large")
	}
	var results []geocodingResult
	if err := json.Unmarshal(data, &results); err != nil {
		return geocodedLocation{}, fmt.Errorf("decoding geocoding response: %w", err)
	}
	if len(results) == 0 {
		return geocodedLocation{}, errLocationNotFound
	}
	lat, err := parseCoordinate(results[0].Lat, 90)
	if err != nil {
		return geocodedLocation{}, err
	}
	lng, err := parseCoordinate(results[0].Lon, 180)
	if err != nil {
		return geocodedLocation{}, err
	}
	return geocodedLocation{lat: lat, lng: lng, name: results[0].label()}, nil
}

type geocodingResult struct {
	Lat, Lon, Name string
	DisplayName    string `json:"display_name"`
	Address        struct {
		City, Town, Village, Municipality, Hamlet, Province string
		StateDistrict                                       string `json:"state_district"`
	}
}

func (r geocodingResult) label() string {
	city := firstPlaceName(r.Address.City, r.Address.Town, r.Address.Village, r.Address.Municipality, r.Address.Hamlet)
	// Spanish provinces are commonly returned as state_district. County is
	// usually a comarca, so it must not be mistaken for the province.
	province := firstPlaceName(r.Address.Province, r.Address.StateDistrict)
	if city != "" && province != "" {
		if strings.EqualFold(city, province) {
			return city
		}
		return city + ", " + province
	}
	return firstPlaceName(r.DisplayName, city, province, r.Name)
}

func firstPlaceName(names ...string) string {
	for _, name := range names {
		if name = strings.TrimSpace(name); name != "" {
			return name
		}
	}
	return ""
}

func parseCoordinate(value string, limit float64) (float64, error) {
	n, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) > limit {
		return 0, fmt.Errorf("invalid geocoding coordinate %q", value)
	}
	return n, nil
}
