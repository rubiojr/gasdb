package main

import (
	"fmt"
	"log"

	"github.com/rubiojr/gasdb/pkg/api"
)

func main() {
	// Create a new API client
	apiClient := api.NewFuelPriceAPI()

	// Example coordinates (Madrid, Spain)
	lat := 40.4168
	lng := -3.7038
	distance := 5000.0 // 5km in meters

	fmt.Printf("Searching for gas stations near coordinates: %.4f, %.4f\n", lat, lng)
	fmt.Printf("Within %.1f km radius\n\n", distance/1000)

	// Fetch nearby prices
	stations, err := apiClient.NearbyPrices(lat, lng, distance)
	if err != nil {
		log.Fatalf("Error fetching nearby prices: %v", err)
	}

	fmt.Printf("Found %d gas stations within the specified radius:\n\n", len(stations))

	// Display the first 5 stations
	limit := 5
	if len(stations) < limit {
		limit = len(stations)
	}

	for i := 0; i < limit; i++ {
		station := stations[i]
		fmt.Printf("Station %d:\n", i+1)
		fmt.Printf("  Name: %s\n", station.Rotulo)
		fmt.Printf("  Address: %s, %s\n", station.Direccion, station.Localidad)
		fmt.Printf("  Coordinates: %s, %s\n", station.Latitud, station.Longitud)

		if station.PrecioGasolina95E5 != "" {
			fmt.Printf("  Gasoline 95: %s €/L\n", station.PrecioGasolina95E5)
		}
		if station.PrecioGasoleoA != "" {
			fmt.Printf("  Diesel: %s €/L\n", station.PrecioGasoleoA)
		}
		fmt.Println()
	}

	if len(stations) > limit {
		fmt.Printf("... and %d more stations\n", len(stations)-limit)
	}
}
