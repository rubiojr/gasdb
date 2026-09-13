package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rubiojr/gasdb/_server/internal/pricesummary"
	"github.com/rubiojr/gasdb/pkg/api"
)

const jsonAPIVersion = 1

type apiMetadata struct {
	APIVersion          int       `json:"api_version"`
	Version             string    `json:"version"`
	GeneratedAt         time.Time `json:"generated_at"`
	PublishedAt         string    `json:"published_at"`
	PublicationTimezone string    `json:"publication_timezone"`
	StorageDate         string    `json:"storage_date,omitempty"`
}

type apiErrorResponse struct {
	APIVersion int    `json:"api_version"`
	State      string `json:"state"`
	Error      string `json:"error"`
}

type apiFuelPrices struct {
	Gasoline95    *float64 `json:"gasolina95"`
	Gasoline98    *float64 `json:"gasolina98"`
	Diesel        *float64 `json:"gasoleo"`
	PremiumDiesel *float64 `json:"gasoleoPremium"`
}

func (p apiFuelPrices) forFuel(fuel string) *float64 {
	switch fuel {
	case "gasolina98":
		return p.Gasoline98
	case "gasoleo":
		return p.Diesel
	case "gasoleoPremium":
		return p.PremiumDiesel
	default:
		return p.Gasoline95
	}
}

type apiStation struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Address      string        `json:"address"`
	Locality     string        `json:"locality"`
	Municipality string        `json:"municipality"`
	Province     string        `json:"province"`
	PostalCode   string        `json:"postal_code"`
	Hours        string        `json:"hours"`
	Lat          float64       `json:"lat"`
	Lng          float64       `json:"lng"`
	DistanceKM   float64       `json:"distance_km"`
	Prices       apiFuelPrices `json:"prices_eur_per_l"`
	Maps         apiMapLinks   `json:"maps"`
}

type apiMapLinks struct {
	OSM        string `json:"osm"`
	GoogleMaps string `json:"google_maps"`
}

type apiRadiusSuggestion struct {
	RadiusKM     int    `json:"radius_km"`
	StationCount int    `json:"station_count"`
	URL          string `json:"url"`
}

type apiSearchResponse struct {
	State             string                `json:"state"`
	Meta              apiMetadata           `json:"meta"`
	Query             apiSearchQuery        `json:"query"`
	TotalStations     int                   `json:"total_stations"`
	ReturnedStations  int                   `json:"returned_stations"`
	Truncated         bool                  `json:"truncated"`
	Stations          []apiStation          `json:"stations"`
	RadiusSuggestions []apiRadiusSuggestion `json:"radius_suggestions"`
	WebURL            string                `json:"web_url"`
}

type apiPriceStats struct {
	StationMin      *float64 `json:"station_min"`
	NationalAverage *float64 `json:"national_average"`
	StationMax      *float64 `json:"station_max"`
	StationCount    int      `json:"station_count"`
}

type apiProvinceAverage struct {
	Province     string  `json:"province"`
	Average      float64 `json:"average_eur_per_l"`
	StationCount int     `json:"station_count"`
}

type apiProvinceRanking struct {
	Lowest  []apiProvinceAverage `json:"lowest"`
	Highest []apiProvinceAverage `json:"highest"`
}

type apiStatsResponse struct {
	State            string                        `json:"state"`
	Meta             apiMetadata                   `json:"meta"`
	DataIsToday      bool                          `json:"data_is_today"`
	NationalPrices   map[string]apiPriceStats      `json:"national_prices"`
	ProvinceAverages map[string]apiProvinceRanking `json:"province_averages"`
}

func apiPrice(raw string) *float64 {
	n, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(raw), ",", "."), 64)
	if err != nil || n <= 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		return nil
	}
	return &n
}

func stationAPIPrices(station *api.GasStation) apiFuelPrices {
	return apiFuelPrices{
		Gasoline95:    apiPrice(station.PrecioGasolina95E5),
		Gasoline98:    apiPrice(station.PrecioGasolina98E5),
		Diesel:        apiPrice(station.PrecioGasoleoA),
		PremiumDiesel: apiPrice(station.PrecioGasoleoPremium),
	}
}

func priceAPIStats(prices pricesummary.Prices) apiPriceStats {
	if prices.Count == 0 {
		return apiPriceStats{}
	}
	return apiPriceStats{StationMin: &prices.Low, NationalAverage: &prices.Average, StationMax: &prices.High, StationCount: prices.Count}
}

func provinceAPIRanking(ranking pricesummary.ProvinceRanking) apiProvinceRanking {
	return apiProvinceRanking{Lowest: provinceAPIAverages(ranking.Cheapest), Highest: provinceAPIAverages(ranking.MostExpensive)}
}

func provinceAPIAverages(prices []pricesummary.ProvincePrice) []apiProvinceAverage {
	result := make([]apiProvinceAverage, 0, len(prices))
	for _, price := range prices {
		result = append(result, apiProvinceAverage{Province: price.Province, Average: price.Average, StationCount: price.Count})
	}
	return result
}
