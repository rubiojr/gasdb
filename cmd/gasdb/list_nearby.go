package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/muesli/gominatim"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/tkrajina/gpxgo/gpx"
	"github.com/urfave/cli/v2"
)

const (
	defaultRadiusKm = 5.0
	metersPerKm     = 1000.0
)

func listNearbyCommand() *cli.Command {
	return &cli.Command{
		Name:  "list-nearby",
		Usage: "List nearby gas stations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "location",
				Usage:    "Location to search",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "db",
				Usage:    "Database file",
				Required: false,
				Value:    "fuel_prices.db",
			},
			&cli.Float64Flag{
				Name:  "lat",
				Usage: "Latitude of the location",
			},
			&cli.Float64Flag{
				Name:  "long",
				Usage: "Longitude of the location",
			},
			&cli.Float64Flag{
				Name:    "radius",
				Aliases: []string{"r"},
				Usage:   "Search radius in kilometers",
				Value:   defaultRadiusKm,
			},
			&cli.StringFlag{
				Name:     "date",
				Usage:    "Date to search",
				Required: false,
				Value:    time.Now().Format("2006-01-02"),
			},
		},
		Action: listNearbyAction,
	}
}

func listNearbyAction(c *cli.Context) error {
	lat := c.Float64("lat")
	lng := c.Float64("long")
	radius := c.Float64("radius")
	loc := c.String("location")

	if loc != "" {
		return listNearbyByName(c.String("db"), loc, radius)
	}

	if lat == 0 && lng == 0 {
		return errors.New("location or latitude and longitude are required")
	}

	return listNearbyStations(c.String("db"), lat, lng, radius)
}

func listNearbyByName(dbPath, name string, distanceKm float64) error {
	gominatim.SetServer("https://nominatim.openstreetmap.org/")
	qry := gominatim.SearchQuery{
		Q: name,
	}

	resp, err := qry.Get()
	if err != nil {
		return err
	}
	fmt.Println("Location found:", resp[0].DisplayName)

	lat, err1 := strconv.ParseFloat(resp[0].Lat, 64)
	if err1 != nil {
		return err1
	}
	lon, err2 := strconv.ParseFloat(resp[0].Lon, 64)
	if err2 != nil {
		return err2
	}
	return listNearbyStations(dbPath, lat, lon, distanceKm)
}

func listNearbyStations(dbPath string, lat, lng, radius float64) error {
	storage, err := gasdb.NewStorage(dbPath, slog.New(slog.DiscardHandler))
	if err != nil {
		return fmt.Errorf("error initializing storage: %w", err)
	}
	defer storage.Close()

	fmt.Println("Filtering stations within\n", radius, "km radius...")

	nearbyStations, err := storage.NearbyPrices(lat, lng, radius*metersPerKm)
	if err != nil {
		return fmt.Errorf("error fetching nearby stations: %w", err)
	}

	for i, station := range nearbyStations {
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
			fmt.Printf("%d. %s (%s)\n", i+1, station.Rotulo, station.Direccion)
			fmt.Printf("   Municipio: %s\n", station.Municipio)
			fmt.Printf("   Distance: %.2f km\n", distance/metersPerKm)
			fmt.Printf("   Gasoline 95: %s €\n", formatDecimal(station.PrecioGasolina95E5))
			fmt.Printf("   Diesel: %s €\n", formatDecimal(station.PrecioGasoleoA))
			fmt.Printf("   Premium Diesel: %s €\n", formatDecimal(station.PrecioGasoleoPremium))
			fmt.Printf("   Coordinates: %s, %s\n\n", formatDecimal(station.Latitud), formatDecimal(station.Longitud))
		}
	}

	fmt.Printf("Found %d stations within %g km radius\n\n", len(nearbyStations), radius)

	return nil
}

func formatDecimal(value string) string {
	return strings.Replace(value, ",", ".", 1)
}
