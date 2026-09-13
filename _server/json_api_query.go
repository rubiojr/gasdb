package main

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

type apiSearchQuery struct {
	Location         string  `json:"location,omitempty"`
	ResolvedLocation string  `json:"resolved_location,omitempty"`
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
	RadiusKM         float64 `json:"radius_km"`
	Fuel             string  `json:"fuel"`
	Limit            int     `json:"limit"`
	Language         string  `json:"lang"`
}

func parseAPIQuery(values url.Values) (apiSearchQuery, error) {
	q := apiSearchQuery{Language: "es"}
	if err := validateAPIKeys(values, "location", "lat", "lng", "radius", "fuel", "limit", "lang"); err != nil {
		return q, err
	}
	var err error
	if q.RadiusKM, err = apiRadius(values.Get("radius")); err != nil {
		return q, err
	}
	if q.Limit, err = apiLimit(values.Get("limit")); err != nil {
		return q, err
	}
	if q.Fuel, err = apiFuel(values.Get("fuel")); err != nil {
		return q, err
	}
	if q.Language, err = apiLanguage(values.Get("lang")); err != nil {
		return q, err
	}
	err = q.readLocation(values)
	return q, err
}

func validateAPIKeys(values url.Values, allowed ...string) error {
	for key, entries := range values {
		if !slices.Contains(allowed, key) {
			return fmt.Errorf("unsupported query parameter: %.64s", key)
		}
		if len(entries) != 1 {
			return fmt.Errorf("%s must not be repeated", key)
		}
	}
	return nil
}

func apiLanguage(raw string) (string, error) {
	switch raw {
	case "", "es":
		return "es", nil
	case "en":
		return "en", nil
	default:
		return "", fmt.Errorf("lang must be en or es")
	}
}

func (q *apiSearchQuery) readLocation(values url.Values) error {
	q.Location = strings.TrimSpace(values.Get("location"))
	if len(q.Location) > 512 || !utf8.ValidString(q.Location) {
		return fmt.Errorf("location must be valid UTF-8 and at most 512 bytes")
	}
	if q.Location != "" {
		return nil
	}
	if values.Get("lat") == "" || values.Get("lng") == "" {
		return fmt.Errorf("provide location or both lat and lng")
	}
	var err error
	if q.Lat, err = parseCoordinate(values.Get("lat"), 90); err != nil {
		return fmt.Errorf("lat must be a finite number between -90 and 90")
	}
	if q.Lng, err = parseCoordinate(values.Get("lng"), 180); err != nil {
		return fmt.Errorf("lng must be a finite number between -180 and 180")
	}
	return nil
}

func apiRadius(raw string) (float64, error) {
	if raw == "" {
		return DefaultRadius, nil
	}
	radius, err := strconv.ParseFloat(raw, 64)
	if err != nil || !(radius > 0 && radius <= 50) {
		return 0, fmt.Errorf("radius must be greater than 0 and at most 50 km")
	}
	return radius, nil
}

func apiLimit(raw string) (int, error) {
	if raw == "" {
		return 20, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > 100 {
		return 0, fmt.Errorf("limit must be an integer from 1 to 100")
	}
	return limit, nil
}

func apiFuel(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "gasolina95":
		return "gasolina95", nil
	case "gasolina98":
		return "gasolina98", nil
	case "gasoleo":
		return "gasoleo", nil
	case "gasoleopremium":
		return "gasoleoPremium", nil
	default:
		return "", fmt.Errorf("fuel must be gasolina95, gasolina98, gasoleo, or gasoleoPremium")
	}
}

func (q apiSearchQuery) values() url.Values {
	values := url.Values{
		"radius": {strconv.FormatFloat(q.RadiusKM, 'f', -1, 64)},
		"fuel":   {q.Fuel}, "limit": {strconv.Itoa(q.Limit)}, "lang": {q.Language},
	}
	if q.Location != "" {
		values.Set("location", q.Location)
	} else {
		values.Set("lat", strconv.FormatFloat(q.Lat, 'f', -1, 64))
		values.Set("lng", strconv.FormatFloat(q.Lng, 'f', -1, 64))
	}
	return values
}

func (q apiSearchQuery) webURL() string {
	values := q.values()
	values.Del("limit")
	return "/search?" + values.Encode()
}
