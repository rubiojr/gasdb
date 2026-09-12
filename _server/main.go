package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/httprate"
	"github.com/patrickmn/go-cache"
	"github.com/rubiojr/gasdb/_server/internal/search"
	"github.com/rubiojr/gasdb/_server/internal/version"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/tkrajina/gpxgo/gpx"
)

const DefaultRadius = 5.0 // km

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	dbPath := flag.String("db", "fuel_prices.db", "Path to the database file")
	showVersion := flag.Bool("version", false, "Print server version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("gasdb-server " + version.String())
		return
	}

	c := cache.New(30*time.Minute, 90*time.Minute)
	geocoder := locationGeocoder{
		client:   &http.Client{Timeout: 10 * time.Second},
		endpoint: "https://nominatim.openstreetmap.org/search",
		cache:    c,
		interval: time.Second,
	}

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
			log.Println("Updating prices")
			if err := storage.UpdateDB(ctx); err != nil {
				logger.Error("Error updating prices", "error", err)
			} else {
				logger.Info("Price update completed successfully")
			}
			log.Println("Prices deleting old records")
			if err := storage.DeleteOldRecords(ctx, 2); err != nil {
				logger.Error("Error deleting old records", "error", err)
			} else {
				logger.Info("Old records cleanup completed successfully")
			}
			log.Println("Prices vacuuming database")
			if err := storage.VacuumDatabase(ctx); err != nil {
				logger.Error("Error vacuuming database", "error", err)
			} else {
				logger.Info("Database vacuum completed successfully")
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
		// Get language from query parameter
		lang := translations.GetLanguageFromQuery(r.URL.Query().Get("lang"))
		t := translations.GetTranslations(lang)

		renderHome(w, r, storage, logger.Logger, t)
	})

	r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Get language from query parameter
		lang := translations.GetLanguageFromQuery(query.Get("lang"))
		t := translations.GetTranslations(lang)

		location := query.Get("location")
		fuelType := query.Get("fuel")

		latStr := query.Get("lat")
		lngStr := query.Get("lng")
		radiusStr := query.Get("radius")

		var lat, lng, radius float64
		var resolvedLocation string
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
			place, err := geocoder.lookup(r.Context(), location)
			if err != nil {
				status := http.StatusNotFound
				if !errors.Is(err, errLocationNotFound) {
					status = http.StatusServiceUnavailable
					t.LocationNotFound = t.LocationSearchUnavailable
					logger.Error("Location lookup failed", "error", err)
				}
				w.WriteHeader(status)
				templates.ResultsPage([]api.StationWithDistance{}, location, "", lat, lng, radius, err, nil, t).Render(r.Context(), w)
				return
			}
			lat, lng, resolvedLocation = place.lat, place.lng, place.name
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
				renderHome(w, r, storage, logger.Logger, t)
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

		var suggestions []search.RadiusSuggestion
		if len(stations) == 0 {
			// Use the cached snapshot instead of logging extra nearby searches.
			prices, err := storage.GetLastPrices(r.Context())
			if err != nil {
				logger.Error("Error getting wider-radius suggestions", "error", err)
			} else {
				suggestions = search.SuggestRadii(prices.ListaEESSPrecio, lat, lng, radius, query)
			}
		}
		templates.ResultsPage(stations, location, resolvedLocation, lat, lng, radius, nil, suggestions, t).Render(r.Context(), w)
	})

	// Start server
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	logger.Debug("Starting server on", "addr", addr)
	log.Fatal(http.ListenAndServe(addr, r))
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
