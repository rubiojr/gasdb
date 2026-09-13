# GasDB JSON API

The read-only API uses `GET /api/search` and `GET /api/stats`. No API key is
required. Responses have `Content-Type: application/json` and `Cache-Control:
no-store`. These routes share the website's limit of **20 requests/minute/IP**.

The production base URL is `https://fuel.rbel.co`. Deploy a build containing the
JSON routes before using them there; older deployments serve only the HTML
interface at `/` and `/search`.

## Search

```bash
curl --silent --show-error --get 'https://fuel.rbel.co/api/search' \
  --data-urlencode 'location=el Masnou, Barcelona' \
  --data-urlencode 'fuel=gasolina95' \
  --data-urlencode 'radius=5' \
  --data-urlencode 'limit=5' \
  --data-urlencode 'lang=en'
```

| Parameter | Meaning |
|---|---|
| `location` | Town, city, landmark, or address in Spain; at most 512 UTF-8 bytes |
| `lat`, `lng` | Decimal coordinates; both required when location is absent |
| `radius` | Finite kilometres, greater than 0 and at most 50; default 5 |
| `fuel` | `gasolina95`, `gasolina98`, `gasoleo`, or `gasoleoPremium`; default `gasolina95` |
| `limit` | Integer 1–100; default 20 |
| `lang` | `es` (default) or `en`; controls linked web pages, not JSON field names |

Location takes precedence over coordinates. Latitude must be in `[-90, 90]` and
longitude in `[-180, 180]`. Fuel names are case-insensitive. Invalid parameters
return HTTP 400 rather than silently selecting a different fuel or radius.
Unknown or repeated query parameters also return HTTP 400.

The response contains:

- `state`: `ok` or `no_stations`.
- `meta`: snapshot and build metadata, described below.
- `query`: normalized parameters, resolved `lat`/`lng`, and
  `resolved_location` for named searches when the geocoder supplies a label.
- `total_stations`: all matching stations before applying `limit`.
- `returned_stations`, `truncated`: describe the returned subset.
- `stations`: sorted by the selected fuel's price, then straight-line distance.
  Missing prices sort last. **Sorting happens before limiting.**
- `radius_suggestions`: wider radii with actual total counts when the current
  search is empty. Each entry has `radius_km`, `station_count`, and a relative
  `/api/search?...` URL preserving the search options.
- `web_url`: a relative link to the equivalent browser search.

Each station has `id`, `name`, `address`, `locality`, `municipality`, `province`,
`postal_code`, `hours` (the source's raw schedule text), numeric `lat`, `lng`,
`distance_km`, `maps.osm`, and `maps.google_maps`. `prices_eur_per_l` contains:

```json
{
  "gasolina95": 1.499,
  "gasolina98": null,
  "gasoleo": 1.399,
  "gasoleoPremium": 1.529
}
```

The numbers above illustrate the shape, not current prices. All four fuel fields
are returned regardless of the sort fuel. Null means missing, invalid, or
nonpositive data, never free fuel. Distances are straight-line km, not route
distances. Format prices to three decimals and distances suitably when displaying
them; the API keeps numeric precision.

## Statistics

`GET /api/stats` returns all four fuels.

It has no province, location, radius, or fuel filter; only optional `lang=es|en`
is accepted for compatibility with clients using the HTML fallback.

- `state`: `ok`, or `no_prices` for a valid snapshot with no usable prices.
- `data_is_today`: whether the publication date matches today's date in Madrid.
- `national_prices.<fuel>`: `station_min`, `national_average`, `station_max`,
  and `station_count`. No valid prices produces null values and count zero.
- `province_averages.<fuel>.lowest` / `.highest`: up to two provinces at each
  end, with `province`, `average_eur_per_l`, and `station_count`.

Province rankings use unrounded province averages, not individual station
extremes. They expose only the two ends of each ranking, not a full province
dataset. An absent province is not evidence that it has no stations.

## Metadata and freshness

Successful responses contain `meta`:

| Field | Meaning |
|---|---|
| `api_version` | Contract version, currently `1` |
| `version` | Server build identifier |
| `generated_at` | UTC response-generation timestamp (RFC 3339) |
| `published_at` | Source publication timestamp, `DD/MM/YYYY H:MM:SS` |
| `publication_timezone` | `Europe/Madrid`, as used by the UI |
| `storage_date` | Database snapshot date, when available |

Use `published_at` for price freshness, not `generated_at`, the build version,
or the storage date. The latter can differ around midnight. Prices come from a
stored government snapshot, polled roughly every six hours, not live pump
telemetry. In-memory snapshot caching is invalidated when new prices are saved.

## Errors

Known API errors have this shape:

```json
{"api_version":1,"state":"location_not_found","error":"Location not found. Try adding the province."}
```

| HTTP | State | Meaning |
|---|---|---|
| 400 | `invalid_request` | Missing or invalid search parameters |
| 404 | `location_not_found` | Geocoder found no matching place |
| 429 | `rate_limited` | Shared IP budget exceeded; honor `Retry-After` |
| 500 | `search_failed` | Station search failed |
| 503 | `geocoding_unavailable` | Place lookup failed temporarily |
| 503 | `prices_unavailable` | No accessible price snapshot |

Unknown routes, unsupported methods, or intermediary failures can return other
formats. Inspect HTTP status and content type before treating a response as JSON.
An empty station search is **HTTP 200**, with `state: "no_stations"`, an empty
station array, and any useful radius suggestions. No result limit creates an
empty search; the minimum accepted limit is one.
