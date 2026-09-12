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
# From the repository root:
scripts/build-server
./_server/gasdb-server
```

The server will start on `http://127.0.0.1:8080` by default.

### Command Line Options

```bash
./gasdb-server -port 8080 -db fuel_prices.db
```

- `-port`: HTTP server port (default: 8080)
- `-db`: Path to SQLite database file (default: fuel_prices.db)
- `-version`: Print the server version and exit without opening the database

### Versioning

The footer on every page displays the same version as `./gasdb-server --version`.
The version is the latest Git tag reachable from the current commit, selected
with `git describe --tags --abbrev=0` (for example, `v1.2.3`). It is embedded in
the binary, so the deployed server does not need Git or a checkout.

Both `scripts/build-server` and the local `scripts/deploy` helper read the tag
on every invocation. After creating a new tag, just build or deploy again.
There is no version file to update, generate, or commit. With no reachable tag,
the build uses the abbreviated commit hash.

The build wrapper uses a temporary Go source overlay to inject the version.
This works with Naked's own linker flags, leaves the checkout untouched, and
keeps concurrent builds independent. Deployment dry runs print the selected tag.

`scripts/build-server` accepts Go build flags, for example:

```bash
scripts/build-server -o /tmp/gasdb-server
```

Plain `go build` bypasses tag injection and uses Go's embedded commit metadata
instead (or `dev` when no metadata is available), so it never reports a stale
saved tag. Root repository tags are not automatically included in that metadata
because `_server` is a separate Go module.

To set a release version, build from `_server` with:

```bash
go build -ldflags "-X github.com/rubiojr/gasdb/_server/internal/version.Version=v1.2.3" -o gasdb-server .
./gasdb-server --version
```

An explicit linker version takes precedence over the injected tag.

## Usage

### Web Interface

1. Visit `http://localhost:8080` in your browser
2. Search by location name (e.g., "Tibidabo, Barcelona") or use your current location
3. View nearby fuel stations with current prices and distances

The interface follows the operating system's light or dark appearance by
default. The sun/moon button at the top right switches themes and remembers
your choice in this browser, including when navigating between search and results.

### Search Options

- **Location Search**: Enter city, neighborhood, or address
- **Resolved Location**: Results show the matched town/city and province alongside your original search. If structured address details are unavailable, the geocoder's full place name is shown instead.
- **Geolocation**: Click "Use My Location" for GPS-based search
- **Radius**: Choose 5, 10, 25, or 50 km in the search form (default: 5 km). Applies to both location-name and GPS searches.

When a location is found but there are no nearby stations, results offer wider
radii up to 50 km with actual station counts from the stored price snapshot.
Only options that add stations are shown. Each link preserves the location,
coordinates, language, and fuel selection while changing the radius. Lookup
failures do not produce radius suggestions.

### Daily Price Summary

The home page lists the lowest, average, and highest reported prices across
Spain for gasoline 95 E5, gasoline 98 E5, diesel A, and premium diesel, in €/L
with three decimal places. Each average uses only stations reporting a valid,
positive price for that fuel; missing prices are not treated as zero.

The summary uses the latest stored snapshot and displays its publication date.
It is labelled as today's prices only when that date matches the current date
in mainland Spain. Older snapshots are labelled as the latest available prices.

The province comparison shows the two cheapest and two most expensive provinces
for the selected fuel, ranked by their unrounded average station price. It shows
each average in €/L and the number of stations contributing to it. Missing or
invalid prices and records without a province are excluded from the province
ranking. Changing the fuel updates the comparison without submitting the search
form; without JavaScript, all four fuel comparisons remain visible.

### URL Parameters

The search endpoint supports direct URL access:

```
/search?location=Madrid
/search?lat=40.4168&lng=-3.7038&radius=5
```

Parameters:
- `location`: Location name to geocode
- `lat`, `lng`: Direct coordinates (decimal degrees)
- `radius`: Search radius in kilometers (default: 5)

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
