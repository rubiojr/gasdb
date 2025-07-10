// Package api provides types and functions to interact with the Spanish government
// fuel price API, fetch fuel station data, and perform geospatial queries.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

const (
	ApiResultOK    = "OK"
	DefaultTimeout = 30 * time.Second
)

// FuelPriceAPI provides methods to fetch fuel price data from the official API.
type FuelPriceAPI struct {
	baseURL    string
	httpClient *http.Client
}

// NewFuelPriceAPI creates a new FuelPriceAPI client with default settings.
func NewFuelPriceAPI() *FuelPriceAPI {
	return &FuelPriceAPI{
		baseURL: "https://sedeaplicaciones.minetur.gob.es/ServiciosRESTCarburantes/PreciosCarburantes/EstacionesTerrestresHist",
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// FetchPricesForDate fetches fuel station prices for a specific date.
func (api *FuelPriceAPI) FetchPricesForDate(date time.Time) (*GasStationList, error) {
	dateStr := date.Format("02-01-2006")
	url := fmt.Sprintf("%s/%s", api.baseURL, dateStr)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var pricesResponse GasStationList
	if err := json.Unmarshal(body, &pricesResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &pricesResponse, nil
}

// FetchPrices fetches the latest available fuel station prices.
func (api *FuelPriceAPI) FetchPrices() (*GasStationList, error) {
	url := strings.Replace(api.baseURL, "EstacionesTerrestresHist", "EstacionesTerrestres", 1)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var pricesResponse GasStationList
	if err := json.Unmarshal(body, &pricesResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &pricesResponse, nil
}

// NearbyPrices returns a list of gas stations within a given distance (meters) from the specified coordinates.
func (api *FuelPriceAPI) NearbyPrices(lat, lng, distance float64) ([]*GasStation, error) {
	prices, err := api.FetchPrices()
	if err != nil {
		return nil, fmt.Errorf("error fetching current prices: %w", err)
	}

	var nearbyStations []*GasStation
	for i := range prices.ListaEESSPrecio {
		station := &prices.ListaEESSPrecio[i]
		stationLat, err := parseLatLong(station.Latitud)
		if err != nil {
			continue
		}

		stationLng, err := parseLatLong(station.Longitud)
		if err != nil {
			continue
		}

		calculatedDistance := gpx.Distance2D(lat, lng, stationLat, stationLng, true)
		if calculatedDistance <= distance {
			nearbyStations = append(nearbyStations, station)
		}
	}

	return nearbyStations, nil
}

// parseLatLong parses a latitude or longitude string (with comma or dot) to float64.
func parseLatLong(s string) (float64, error) {
	s = strings.Replace(s, ",", ".", 1)
	m, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return m, nil
}
