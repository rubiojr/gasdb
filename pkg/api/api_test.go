package api

import (
	"testing"
	"time"
)

func TestFuelPriceAPI_FetchPrices(t *testing.T) {
	api := NewFuelPriceAPI()

	prices, err := api.FetchPrices()
	if err != nil {
		t.Fatalf("FetchPrices() failed: %v", err)
	}

	if prices == nil {
		t.Fatal("FetchPrices() returned nil prices")
	}

	if prices.ResultadoConsulta != ApiResultOK {
		t.Errorf("Expected ResultadoConsulta to be 'OK', got '%s'", prices.ResultadoConsulta)
	}

	if len(prices.ListaEESSPrecio) == 0 {
		t.Error("Expected at least one gas station in the response")
	}

	// Verify structure of first station
	if len(prices.ListaEESSPrecio) > 0 {
		station := prices.ListaEESSPrecio[0]
		if station.IDEESS == "" {
			t.Error("Expected station to have IDEESS")
		}
		if station.Latitud == "" {
			t.Error("Expected station to have Latitud")
		}
		if station.Longitud == "" {
			t.Error("Expected station to have Longitud")
		}
		if station.Rotulo == "" {
			t.Error("Expected station to have Rotulo")
		}
	}
}

func TestFuelPriceAPI_FetchPricesForDate(t *testing.T) {
	api := NewFuelPriceAPI()

	// Test with a recent date (yesterday)
	yesterday := time.Now().AddDate(0, 0, -1)

	prices, err := api.FetchPricesForDate(yesterday)
	if err != nil {
		t.Fatalf("FetchPricesForDate() failed: %v", err)
	}

	if prices == nil {
		t.Fatal("FetchPricesForDate() returned nil prices")
	}

	if prices.ResultadoConsulta != ApiResultOK {
		t.Errorf("Expected ResultadoConsulta to be 'OK', got '%s'", prices.ResultadoConsulta)
	}

	// Should have some stations
	if len(prices.ListaEESSPrecio) == 0 {
		t.Error("Expected at least one gas station in the response")
	}
}

func TestFuelPriceAPI_NearbyPrices(t *testing.T) {
	api := NewFuelPriceAPI()

	// Test with Madrid coordinates
	lat := 40.4168
	lng := -3.7038
	distance := 5000.0 // 5km

	stations, err := api.NearbyPrices(lat, lng, distance)
	if err != nil {
		t.Fatalf("NearbyPrices() failed: %v", err)
	}

	if stations == nil {
		t.Fatal("NearbyPrices() returned nil stations")
	}

	// Should find at least some stations in Madrid within 5km
	if len(stations) == 0 {
		t.Error("Expected to find at least one gas station near Madrid")
	}

	// Verify all returned stations are within the specified distance
	for i, station := range stations {
		if station == nil {
			t.Errorf("Station %d is nil", i)
			continue
		}

		// Parse station coordinates
		stationLat, err := parseLatLong(station.Latitud)
		if err != nil {
			t.Errorf("Failed to parse station %d latitude: %v", i, err)
			continue
		}

		stationLng, err := parseLatLong(station.Longitud)
		if err != nil {
			t.Errorf("Failed to parse station %d longitude: %v", i, err)
			continue
		}

		// Check if station has required fields
		if station.IDEESS == "" {
			t.Errorf("Station %d missing IDEESS", i)
		}
		if station.Rotulo == "" {
			t.Errorf("Station %d missing Rotulo", i)
		}

		// Calculate distance to verify it's within range
		// Simple distance check (not exact but good enough for test)
		latDiff := lat - stationLat
		lngDiff := lng - stationLng
		distanceSquared := latDiff*latDiff + lngDiff*lngDiff

		// Very rough check - within reasonable bounds for 5km
		if distanceSquared > 0.05 { // Approximately 5km in degrees squared
			t.Logf("Warning: Station %d (%s) seems far from search center", i, station.Rotulo)
		}
	}

	// Test with smaller radius
	smallerStations, err := api.NearbyPrices(lat, lng, 1000.0) // 1km
	if err != nil {
		t.Fatalf("NearbyPrices() with smaller radius failed: %v", err)
	}

	// Should have fewer or equal stations with smaller radius
	if len(smallerStations) > len(stations) {
		t.Error("Smaller radius search returned more stations than larger radius")
	}
}

func TestFuelPriceAPI_InvalidCoordinates(t *testing.T) {
	api := NewFuelPriceAPI()

	// Test with coordinates in the ocean (should return no results)
	lat := 0.0
	lng := 0.0
	distance := 1000.0

	stations, err := api.NearbyPrices(lat, lng, distance)
	if err != nil {
		t.Fatalf("NearbyPrices() with ocean coordinates failed: %v", err)
	}

	// Should return empty slice for coordinates in the ocean
	if len(stations) > 0 {
		t.Logf("Found %d stations at ocean coordinates (this might be expected)", len(stations))
	}
}

func TestParseLatLong(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"40.4168", 40.4168, false},
		{"40,4168", 40.4168, false}, // Spanish decimal format
		{"-3.7038", -3.7038, false},
		{"-3,7038", -3.7038, false}, // Spanish decimal format
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		result, err := parseLatLong(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("parseLatLong(%q) expected error but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseLatLong(%q) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("parseLatLong(%q) = %f, expected %f", test.input, result, test.expected)
			}
		}
	}
}

func BenchmarkFuelPriceAPI_FetchPrices(b *testing.B) {
	api := NewFuelPriceAPI()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := api.FetchPrices()
		if err != nil {
			b.Fatalf("FetchPrices() failed: %v", err)
		}
	}
}

func BenchmarkFuelPriceAPI_NearbyPrices(b *testing.B) {
	api := NewFuelPriceAPI()
	lat := 40.4168
	lng := -3.7038
	distance := 5000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := api.NearbyPrices(lat, lng, distance)
		if err != nil {
			b.Fatalf("NearbyPrices() failed: %v", err)
		}
	}
}
