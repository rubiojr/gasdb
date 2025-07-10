# GasDB Web Server

A web-based interface for finding nearby fuel stations in Spain with current prices.

## Features

- 🌐 Clean web interface for searching fuel stations
- 📍 Location-based search (address or coordinates)
- 🧭 Geolocation support for finding nearby stations
- 📊 Real-time fuel price data from Spanish government API
- 🔄 Automatic price updates (4 times daily)
- ⚡ In-memory caching for geocoding results
- 🛡️ Rate limiting and request logging

## Quick Start

### Build and Run

```bash
cd _server
go build -o gasdb-server .
./gasdb-server
```

The server will start on `http://127.0.0.1:8080` by default.

### Command Line Options

```bash
./gasdb-server -port 8080 -db fuel_prices.db
```

- `-port`: HTTP server port (default: 8080)
- `-db`: Path to SQLite database file (default: fuel_prices.db)

## Usage

### Web Interface

1. Visit `http://localhost:8080` in your browser
2. Search by location name (e.g., "Tibidabo, Barcelona") or use your current location
3. View nearby fuel stations with current prices and distances

### Search Options

- **Location Search**: Enter city, neighborhood, or address
- **Geolocation**: Click "Use My Location" for GPS-based search
- **Radius**: Default 3km search radius (can be customized via URL parameters)

### URL Parameters

The search endpoint supports direct URL access:

```
/search?location=Madrid
/search?lat=40.4168&lng=-3.7038&radius=5
```

Parameters:
- `location`: Location name to geocode
- `lat`, `lng`: Direct coordinates (decimal degrees)
- `radius`: Search radius in kilometers (default: 3)

## Architecture

### Components

- **HTTP Server**: Chi router with middleware for logging, rate limiting, and recovery
- **Template Engine**: Templ-based HTML templates for the UI
- **Geocoding**: OpenStreetMap Nominatim API for location search
- **Caching**: In-memory cache for geocoding results (30-minute TTL)
- **Database**: SQLite storage with automatic price updates

### Automatic Updates

The server runs background price updates every 6 hours to keep fuel prices current. Updates are logged and any errors are reported.

### Rate Limiting

Built-in rate limiting allows 20 requests per minute per IP address to prevent abuse.

## Templates

The web interface uses server-side rendered templates:

- `base.templ`: Common layout and styling
- `home.templ`: Search form with geolocation support
- `results.templ`: Station listings with prices and distances

Templates are compiled to Go code for fast rendering.

## Dependencies

- **Chi**: HTTP router and middleware
- **Templ**: Type-safe HTML templates
- **Gominatim**: OpenStreetMap geocoding client
- **Go-cache**: In-memory caching
- **GPX**: Distance calculations
- **GasDB**: Core fuel station database and API

## Development

### Building Templates

If you modify `.templ` files, regenerate the Go code:

```bash
templ generate
```

### Database

The server expects a SQLite database with Spanish fuel station data. The database is automatically updated from the official government API.

## Configuration

The server is configured for Spanish fuel stations and uses:

- OpenStreetMap Nominatim for geocoding
- Spanish Ministry of Industry API for fuel prices
- Local SQLite database for fast queries

## License

MIT License - see parent project LICENSE file for details.