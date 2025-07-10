package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/httprate"
	"github.com/muesli/gominatim"
	"github.com/patrickmn/go-cache"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/tkrajina/gpxgo/gpx"
)

const DefaultRadius = 5.0 // km

func main() {
	c := cache.New(30*time.Minute, 90*time.Minute)
	port := flag.Int("port", 8080, "HTTP server port")
	dbPath := flag.String("db", "fuel_prices.db", "Path to the database file")
	flag.Parse()

	ctx := context.Background()

	logger := httplog.NewLogger("gasdb", httplog.Options{
		JSON:            false,
		LogLevel:        slog.LevelDebug,
		Concise:         true,
		QuietDownPeriod: 10 * time.Second,
	})

	// Initialize storage
	storage, err := gasdb.NewStorage(ctx, *dbPath, logger.Logger)
	if err != nil {
		log.Fatalf("Error initializing storage: %v", err)
	}
	defer storage.Close()

	// Spawn a goroutine to update daily prices 4 times per day
	go func() {
		updateInterval := 6 * 60 * 60 // 6 hours in seconds (4 times per day)
		ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)
		defer ticker.Stop()

		for {
			if err := storage.UpdateDB(ctx); err != nil {
				logger.Error("Error updating prices", "error", err)
			} else {
				logger.Info("Price update completed successfully")
			}
			<-ticker.C
		}
	}()

	// Create router
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(20, time.Minute))

	// Define routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		lastUpdate, err := storage.GetLastUpdateDate(r.Context())
		if err != nil {
			logger.Error("Error getting last update date", "error", err)
		}
		templates.Home(lastUpdate).Render(r.Context(), w)
	})

	r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		location := query.Get("location")
		fuelType := query.Get("fuel")

		latStr := query.Get("lat")
		lngStr := query.Get("lng")
		radiusStr := query.Get("radius")

		var lat, lng, radius float64
		var err error

		// Set default radius if not provided or invalid
		if radiusStr == "" {
			radius = DefaultRadius
		} else {
			radius, err = strconv.ParseFloat(radiusStr, 64)
			if err != nil || radius <= 0 {
				radius = DefaultRadius
			}
		}

		// Set default fuel type if not provided
		if fuelType == "" {
			fuelType = "gasolina95"
		}

		// Handle location search or direct coordinates
		if location != "" {
			lat, lng, err = geocodeLocation(location, c)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				templates.ResultsPage([]api.StationWithDistance{}, location, lat, lng, radius, err).Render(r.Context(), w)
				return
			}
		} else {
			// Try to parse latitude and longitude
			if latStr != "" && lngStr != "" {
				lat, err = strconv.ParseFloat(latStr, 64)
				if err != nil {
					http.Error(w, "Invalid latitude value", http.StatusBadRequest)
					return
				}

				lng, err = strconv.ParseFloat(lngStr, 64)
				if err != nil {
					http.Error(w, "Invalid longitude value", http.StatusBadRequest)
					return
				}
			} else {
				// If neither location nor coordinates are provided, show the home page
				lastUpdate, err := storage.GetLastUpdateDate(r.Context())
				if err != nil {
					logger.Error("Error getting last update date", "error", err)
				}
				templates.Home(lastUpdate).Render(r.Context(), w)
				return
			}
		}

		// Find nearby stations
		nearbyStations, err := storage.NearbyPrices(ctx, lat, lng, radius*1000)
		if err != nil {
			http.Error(w, "Error finding nearby stations: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate distances and sort by distance
		stations := make([]api.StationWithDistance, 0, len(nearbyStations))
		for _, station := range nearbyStations {
			stationLat, err := gasdb.ParseLatLong(station.Latitud)
			if err != nil {
				continue
			}

			stationLng, err := gasdb.ParseLatLong(station.Longitud)
			if err != nil {
				continue
			}

			distance := gpx.Distance2D(lat, lng, stationLat, stationLng, true)
			if distance <= radius*1000 {
				stations = append(stations, api.StationWithDistance{
					Station:  station,
					Distance: distance,
				})
			}
		}

		// Sort by price (cheapest first), then by distance
		sort.Slice(stations, func(i, j int) bool {
			priceI := getFuelPrice(stations[i].Station, fuelType)
			priceJ := getFuelPrice(stations[j].Station, fuelType)

			// If both have prices, sort by price
			if priceI > 0 && priceJ > 0 {
				return priceI < priceJ
			}
			// If only one has a price, prioritize it
			if priceI > 0 && priceJ == 0 {
				return true
			}
			if priceI == 0 && priceJ > 0 {
				return false
			}
			// If neither has a price, sort by distance
			return stations[i].Distance < stations[j].Distance
		})

		templates.ResultsPage(stations, location, lat, lng, radius, nil).Render(r.Context(), w)
	})

	// Start server
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	logger.Debug("Starting server on", "addr", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func gominatimResultToLatLon(result gominatim.SearchResult) (lat, lng float64, err error) {
	lat, err = strconv.ParseFloat(result.Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing latitude: %w", err)
	}

	lng, err = strconv.ParseFloat(result.Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing longitude: %w", err)
	}

	return lat, lng, nil
}

func geocodeLocation(location string, c *cache.Cache) (lat, lng float64, err error) {
	// Configure Nominatim geocoder
	gominatim.SetServer("https://nominatim.openstreetmap.org/")
	if cachedLocation, ok := c.Get(location); ok {
		result := cachedLocation.(gominatim.SearchResult)
		return gominatimResultToLatLon(result)
	}

	// URL encode the location query
	encodedLocation := url.QueryEscape(location)

	// Create search query
	query := gominatim.SearchQuery{
		Q: encodedLocation,
	}

	// Get results
	results, err := query.Get()
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding error: %w", err)
	}

	// Check if we have results
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no results found for location: %s", location)
	}
	c.Set(location, results[0], cache.DefaultExpiration)

	return gominatimResultToLatLon(results[0])
}

func getFuelPrice(station *api.GasStation, fuelType string) float64 {
	var priceStr string

	switch strings.ToLower(fuelType) {
	case "gasolina95", "gasolina95e5":
		priceStr = station.PrecioGasolina95E5
	case "gasolina95e10":
		priceStr = station.PrecioGasolina95E10
	case "gasolina98", "gasolina98e5":
		priceStr = station.PrecioGasolina98E5
	case "gasolina98e10":
		priceStr = station.PrecioGasolina98E10
	case "gasolina95premium":
		priceStr = station.PrecioGasolina95E5Prem
	case "gasoleo", "gasoleoA":
		priceStr = station.PrecioGasoleoA
	case "gasoleoB":
		priceStr = station.PrecioGasoleoB
	case "gasoleoPremium":
		priceStr = station.PrecioGasoleoPremium
	case "biodiesel":
		priceStr = station.PrecioBiodiesel
	case "bioetanol":
		priceStr = station.PrecioBioetanol
	case "glp", "gaseslicuados":
		priceStr = station.PrecioGasesLicuados
	case "gnc", "gasnatural":
		priceStr = station.PrecioGasNaturalComp
	case "gnl", "gasnaturallicuado":
		priceStr = station.PrecioGasNaturalLicuado
	case "hidrogeno":
		priceStr = station.PrecioHidrogeno
	default:
		priceStr = station.PrecioGasolina95E5
	}

	// Replace comma with dot for proper float parsing
	priceStr = strings.Replace(priceStr, ",", ".", 1)

	// Parse the price, return 0 if invalid or empty
	if priceStr == "" || priceStr == "-" {
		return 0
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0
	}

	return price
}
