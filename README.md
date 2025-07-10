# GasDB

[![Go Reference](https://pkg.go.dev/badge/github.com/rubiojr/gasdb.svg)](https://pkg.go.dev/github.com/rubiojr/gasdb)
[![Test](https://github.com/rubiojr/gasdb/actions/workflows/test.yml/badge.svg)](https://github.com/rubiojr/gasdb/actions/workflows/test.yml)
[![CI](https://github.com/rubiojr/gasdb/actions/workflows/ci.yml/badge.svg)](https://github.com/rubiojr/gasdb/actions/workflows/ci.yml)

A Go library for fetching and filtering Spanish fuel station prices from the official government API.

## Features

- 🚗 Fetch current fuel prices from all Spanish gas stations
- 📅 Retrieve historical price data by date
- 📍 Find nearby stations by coordinates and distance
- 🏪 Access detailed station information (address, hours, fuel types)
- 🔄 Built-in distance calculation and filtering
- 📊 Clean, type-safe API

## Installation

```bash
go get github.com/rubiojr/gasdb/pkg/api
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/rubiojr/gasdb/pkg/api"
)

func main() {
    // Create API client
    client := api.NewFuelPriceAPI()

    // Find gas stations near Madrid within 5km
    lat := 40.4168
    lng := -3.7038
    distance := 5000.0 // meters

    stations, err := client.NearbyPrices(lat, lng, distance)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d stations near Madrid:\n\n", len(stations))

    for i, station := range stations {
        if i >= 3 { // Show first 3 stations
            break
        }
        
        fmt.Printf("🏪 %s\n", station.Rotulo)
        fmt.Printf("📍 %s, %s\n", station.Direccion, station.Localidad)
        
        if station.PrecioGasolina95E5 != "" {
            fmt.Printf("⛽ Gasoline 95: %s €/L\n", station.PrecioGasolina95E5)
        }
        if station.PrecioGasoleoA != "" {
            fmt.Printf("🚛 Diesel: %s €/L\n", station.PrecioGasoleoA)
        }
        fmt.Println()
    }
}
```

## API Reference

### Fetch Current Prices

```go
prices, err := client.FetchPrices()
```

### Fetch Historical Prices

```go
date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
prices, err := client.FetchPricesForDate(date)
```

### Find Nearby Stations

```go
// Barcelona coordinates, 10km radius
stations, err := client.NearbyPrices(41.3851, 2.1734, 10000)
```

## Data Structure

Each gas station includes:

- **Location**: Coordinates, address, municipality, province
- **Details**: Name (Rótulo), opening hours, station ID
- **Fuel Prices**: Gasoline 95/98, Diesel, Premium fuels, etc.
- **Additional**: Biofuel percentages, service type

## CLI Tool

The project includes a command-line tool:

```bash
# Build the CLI
cd cmd/gasdb && go build .

# Find nearby stations
./gasdb nearby --lat 40.4168 --lng -3.7038 --radius 5
```

## Web Server

A web interface is also available:

```bash
# Build and run the server
cd _server && go build . && ./server
```

Visit `http://localhost:8080` to search for gas stations via web interface.

## Testing

Run tests including integration tests with the real API:

```bash
# All tests
go test ./...

# Integration tests only
cd pkg/api && go test -v -run TestFuelPriceAPI

# Local test script
./scripts/test.sh
```

See [TESTING.md](TESTING.md) for detailed testing information.

## Data Source

This library fetches data from the official Spanish Ministry of Industry API:
- **Current prices**: `EstacionesTerrestres` endpoint
- **Historical prices**: `EstacionesTerrestresHist/{date}` endpoint

Data is provided by the Spanish government and updated regularly.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for any new functionality
4. Run the test suite: `./scripts/test.sh`
5. Submit a pull request

Integration tests are run automatically on all pull requests.