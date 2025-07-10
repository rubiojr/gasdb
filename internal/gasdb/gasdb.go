package gasdb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/patrickmn/go-cache"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/tkrajina/gpxgo/gpx"
)

type Storage struct {
	db    *sql.DB
	cache *cache.Cache
	log   *slog.Logger
}

func NewStorage(dbPath string, logger *slog.Logger) (*Storage, error) {
	db, err := sql.Open("sqlite3", "file:"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set PRAGMA options
	if _, err = db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting journal mode: %w", err)
	}
	if _, err = db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting busy timeout: %w", err)
	}
	if _, err = db.Exec("PRAGMA synchronous = NORMAL;"); err != nil { // or OFF for migrate
		db.Close()
		return nil, fmt.Errorf("error setting synchronous: %w", err)
	}
	if _, err = db.Exec("PRAGMA cache_size = 8000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting cache size: %w", err)
	}

	// Set optimal page size
	_, err = db.Exec("PRAGMA page_size = 4096;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting page size: %w", err)
	}

	// Configure memory mapping
	_, err = db.Exec("PRAGMA mmap_size = 67108864") // 64MB
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting mmap size: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS fuel_prices (
		date TEXT PRIMARY KEY,
		data JSON,
		fetched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating table: %w", err)
	}

	// Initialize the cache with default expiration of 5 minutes and cleanup interval of 10 minutes
	c := cache.New(10*time.Minute, 30*time.Minute)

	s := &Storage{
		db:    db,
		cache: c,
		log:   logger,
	}

	err = s.CreateHistoricPricesTable()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating historic_prices table: %w", err)
	}

	err = s.CreateTrigger()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating trigger: %w", err)
	}

	err = s.CreateLocationLogsTable()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating location_logs table: %w", err)
	}

	return s, nil
}

func NewStorageMigrate(dbPath string, logger *slog.Logger) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	// Set PRAGMA options
	if _, err = db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting journal mode: %w", err)
	}
	if _, err = db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting busy timeout: %w", err)
	}
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}
	if _, err = db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting synchronous mode: %w", err)
	}
	if _, err = db.Exec("PRAGMA cache_size = 1000000000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting cache size: %w", err)
	}
	if _, err = db.Exec("PRAGMA temp_store = memory"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting temp store: %w", err)
	}
	if _, err = db.Exec("PRAGMA mmap_size = 268435456"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error setting mmap size: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	c := cache.New(5*time.Minute, 10*time.Minute)

	s := &Storage{
		db:    db,
		cache: c,
		log:   logger,
	}

	// Create historic prices table
	if err := s.CreateHistoricPricesTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error creating historic prices table: %w", err)
	}

	return s, nil
}

func createTables(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS fuel_prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT UNIQUE NOT NULL,
		data BLOB NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_fuel_prices_date ON fuel_prices(date);
	`

	_, err := db.Exec(createTableSQL)
	return err
}

func (s *Storage) CreateTrigger() error {
	createTriggerSQL := `
	CREATE TRIGGER IF NOT EXISTS insert_historic_prices
	AFTER INSERT ON fuel_prices
	BEGIN
		INSERT OR REPLACE INTO historic_prices (
			date, ideess, cp, direccion, horario, latitud, localidad, longitud,
			margen, municipio, provincia, rotulo, tipo_venta, precio_biodiesel,
			precio_bioetanol, precio_gas_natural_comp, precio_gas_natural_licuado,
			precio_gases_licuados, precio_gasoleo_a, precio_gasoleo_b, precio_gasoleo_premium,
			precio_gasolina_95_e10, precio_gasolina_95_e5, precio_gasolina_95_e5_prem,
			precio_gasolina_98_e10, precio_gasolina_98_e5, precio_hidrogeno,
			porcentaje_bioetanol, porcentaje_ester_metilico, idmunicipio, idprovincia, idccaa
		)
		SELECT
			NEW.date,
			json_extract(station.value, '$.IDEESS'),
			json_extract(station.value, '$."C.P."'),
			json_extract(station.value, '$."Dirección"'),
			json_extract(station.value, '$.Horario'),
			json_extract(station.value, '$.Latitud'),
			json_extract(station.value, '$.Localidad'),
			json_extract(station.value, '$."Longitud (WGS84)"'),
			json_extract(station.value, '$.Margen'),
			json_extract(station.value, '$.Municipio'),
			json_extract(station.value, '$.Provincia'),
			json_extract(station.value, '$."Rótulo"'),
			json_extract(station.value, '$."Tipo Venta"'),
			json_extract(station.value, '$."Precio Biodiesel"'),
			json_extract(station.value, '$."Precio Bioetanol"'),
			json_extract(station.value, '$."Precio Gas Natural Comprimido"'),
			json_extract(station.value, '$."Precio Gas Natural Licuado"'),
			json_extract(station.value, '$."Precio Gases licuados del petróleo"'),
			json_extract(station.value, '$."Precio Gasoleo A"'),
			json_extract(station.value, '$."Precio Gasoleo B"'),
			json_extract(station.value, '$."Precio Gasoleo Premium"'),
			json_extract(station.value, '$."Precio Gasolina 95 E10"'),
			json_extract(station.value, '$."Precio Gasolina 95 E5"'),
			json_extract(station.value, '$."Precio Gasolina 95 E5 Premium"'),
			json_extract(station.value, '$."Precio Gasolina 98 E10"'),
			json_extract(station.value, '$."Precio Gasolina 98 E5"'),
			json_extract(station.value, '$."Precio Hidrogeno"'),
			json_extract(station.value, '$."% BioEtanol"'),
			json_extract(station.value, '$."% Éster metílico"'),
			json_extract(station.value, '$.IDMunicipio'),
			json_extract(station.value, '$.IDProvincia'),
			json_extract(station.value, '$.IDCCAA')
		FROM json_each(json_extract(NEW.data, '$.ListaEESSPrecio')) AS station;
	END;
	`

	_, err := s.db.Exec(createTriggerSQL)
	if err != nil {
		return fmt.Errorf("error creating trigger: %w", err)
	}

	return nil
}

func (s *Storage) MigrateToHistoricPrices() error {
	s.log.Debug("Migrating to historic_prices table")
	rows, err := s.db.Query("SELECT date, data FROM fuel_prices ORDER BY date")
	if err != nil {
		return fmt.Errorf("error querying fuel_prices: %w", err)
	}
	defer rows.Close()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO historic_prices (
			date, ideess, cp, direccion, horario, latitud, localidad, longitud,
			margen, municipio, provincia, rotulo, tipo_venta, precio_biodiesel,
			precio_bioetanol, precio_gas_natural_comp, precio_gas_natural_licuado,
			precio_gases_licuados, precio_gasoleo_a, precio_gasoleo_b, precio_gasoleo_premium,
			precio_gasolina_95_e10, precio_gasolina_95_e5, precio_gasolina_95_e5_prem,
			precio_gasolina_98_e10, precio_gasolina_98_e5, precio_hidrogeno,
			porcentaje_bioetanol, porcentaje_ester_metilico, idmunicipio, idprovincia, idccaa
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	for rows.Next() {
		s.log.Debug("Processing row...")
		var dateStr string
		var jsonData []byte
		if err := rows.Scan(&dateStr, &jsonData); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		var stationList api.GasStationList
		if err := json.Unmarshal(jsonData, &stationList); err != nil {
			s.log.Warn("Warning: error unmarshaling data for date", "date", dateStr, "error", err)
			continue
		}

		for _, station := range stationList.ListaEESSPrecio {
			_, err := stmt.Exec(
				dateStr, station.IDEESS, station.CP, station.Direccion, station.Horario,
				station.Latitud, station.Localidad, station.Longitud, station.Margen,
				station.Municipio, station.Provincia, station.Rotulo, station.TipoVenta,
				station.PrecioBiodiesel, station.PrecioBioetanol, station.PrecioGasNaturalComp,
				station.PrecioGasNaturalLicuado, station.PrecioGasesLicuados, station.PrecioGasoleoA,
				station.PrecioGasoleoB, station.PrecioGasoleoPremium, station.PrecioGasolina95E10,
				station.PrecioGasolina95E5, station.PrecioGasolina95E5Prem, station.PrecioGasolina98E10,
				station.PrecioGasolina98E5, station.PrecioHidrogeno, station.PorcentajeBioEtanol,
				station.PorcentajeEsterMetilico, station.IDMunicipio, station.IDProvincia, station.IDCCAA,
			)
			if err != nil {
				s.log.Warn("Warning: error inserting station", "ideess", station.IDEESS, "error", err)
				continue
			}
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	s.log.Debug("Migration completed successfully")
	return nil
}

func (s *Storage) CreateHistoricPricesTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS historic_prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		ideess TEXT NOT NULL,
		cp TEXT,
		direccion TEXT,
		horario TEXT,
		latitud TEXT,
		localidad TEXT,
		longitud TEXT,
		margen TEXT,
		municipio TEXT,
		provincia TEXT,
		rotulo TEXT,
		tipo_venta TEXT,
		precio_biodiesel TEXT,
		precio_bioetanol TEXT,
		precio_gas_natural_comp TEXT,
		precio_gas_natural_licuado TEXT,
		precio_gases_licuados TEXT,
		precio_gasoleo_a TEXT,
		precio_gasoleo_b TEXT,
		precio_gasoleo_premium TEXT,
		precio_gasolina_95_e10 TEXT,
		precio_gasolina_95_e5 TEXT,
		precio_gasolina_95_e5_prem TEXT,
		precio_gasolina_98_e10 TEXT,
		precio_gasolina_98_e5 TEXT,
		precio_hidrogeno TEXT,
		porcentaje_bioetanol TEXT,
		porcentaje_ester_metilico TEXT,
		idmunicipio TEXT,
		idprovincia TEXT,
		idccaa TEXT,
		UNIQUE(date, ideess)
	);
	CREATE INDEX IF NOT EXISTS idx_historic_prices_date ON historic_prices(date);
	CREATE INDEX IF NOT EXISTS idx_historic_prices_ideess ON historic_prices(ideess);
	CREATE INDEX IF NOT EXISTS idx_historic_prices_latitud_longitud ON historic_prices(latitud, longitud);
	`

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating historic_prices table: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	// Clear the cache before closing
	if s.cache != nil {
		s.cache.Flush()
	}
	return s.db.Close()
}

func (s *Storage) SavePrices(date time.Time, data []byte) error {
	dateStr := date.Format("2006-01-02")

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT OR REPLACE INTO fuel_prices (date, data) VALUES (?, ?)", dateStr, data)
	if err != nil {
		return fmt.Errorf("error inserting data: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	// Clear cache for this date and related keys
	s.cache.Delete("last_price")
	s.cache.Flush()

	return nil
}

func (s *Storage) HasDate(date time.Time) (bool, error) {
	dateStr := date.Format("2006-01-02")
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM fuel_prices WHERE date = ?", dateStr).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking date existence: %w", err)
	}
	return count > 0, nil
}

func (s *Storage) GetLastPrice() (*api.GasStationList, error) {
	// Use a static cache key for the last price
	const cacheKey = "last_price"

	// Try to get data from cache
	if cachedData, found := s.cache.Get(cacheKey); found {
		// Return the cached data if found
		s.log.Debug("Using cached data", "key", cacheKey)
		return cachedData.(*api.GasStationList), nil
	}

	// If not in cache, fetch from database
	var jsonData []byte
	err := s.db.QueryRow("SELECT data FROM fuel_prices ORDER BY date DESC LIMIT 1").Scan(&jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no data available")
		}
		return nil, fmt.Errorf("error querying database: %w", err)
	}

	var pricesResponse api.GasStationList
	if err := json.Unmarshal(jsonData, &pricesResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling data: %w", err)
	}

	// Store the result in cache for future use
	s.cache.Set(cacheKey, &pricesResponse, cache.DefaultExpiration)

	return &pricesResponse, nil
}

func (s *Storage) NearbyPrices(lat, lng, distance float64) ([]*api.GasStation, error) {
	// Create a cache key based on the parameters
	cacheKey := fmt.Sprintf("nearby_prices_%f_%f_%f", lat, lng, distance)

	// Log the search location
	newLat, newLong := reduceLocationPrecision(lat, lng, 2)
	err := s.LogSearchLocation(newLat, newLong, distance)
	if err != nil {
		// Log error but don't fail the search if logging fails
		s.log.Error("Failed to log search location", "error", err)
	} else {
		s.log.Debug("Search location logged", "latitude", lat, "longitude", lng)
	}

	// Try to get data from cache
	if cachedData, found := s.cache.Get(cacheKey); found {
		// Return the cached data if found
		s.log.Debug("Using cached data", "key", cacheKey)
		return cachedData.([]*api.GasStation), nil
	}
	s.log.Debug("Fetching data from database, cached data not found", "key", cacheKey)

	// If not in cache, fetch from database
	pricesResponse, err := s.GetLastPrice()
	if err != nil {
		return nil, fmt.Errorf("error getting last price: %w", err)
	}

	var nearbyStations []*api.GasStation
	for _, station := range pricesResponse.ListaEESSPrecio {
		stationLat, err := ParseLatLong(station.Latitud)
		if err != nil {
			continue
		}

		stationLng, err := ParseLatLong(station.Longitud)
		if err != nil {
			continue
		}

		calculatedDistance := gpx.Distance2D(lat, lng, stationLat, stationLng, true)
		if calculatedDistance <= distance {
			nearbyStations = append(nearbyStations, &station)
		}
	}

	// Store the result in cache for future use
	s.cache.Set(cacheKey, nearbyStations, cache.DefaultExpiration)

	return nearbyStations, nil
}

func (s *Storage) GetPricesForDate(date time.Time) (*api.GasStationList, error) {
	dateStr := date.Format("2006-01-02")

	var jsonData []byte
	err := s.db.QueryRow("SELECT data FROM fuel_prices WHERE date = ?", dateStr).Scan(&jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no data available for date %s", dateStr)
		}
		return nil, fmt.Errorf("error querying database: %w", err)
	}

	var pricesResponse api.GasStationList
	if err := json.Unmarshal(jsonData, &pricesResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling data: %w", err)
	}

	return &pricesResponse, nil
}

func (s *Storage) UpdateDB() error {
	api := api.NewFuelPriceAPI()
	pricesResponse, err := api.FetchPrices()
	if err != nil {
		return err
	}

	if pricesResponse.ResultadoConsulta != "OK" {
		return fmt.Errorf("API returned non-OK result: %s", pricesResponse.ResultadoConsulta)
	}

	jsonData, err := json.Marshal(pricesResponse)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %w", err)
	}

	if err := s.SavePrices(time.Now(), jsonData); err != nil {
		return fmt.Errorf("Error saving data for today: %w", err)
	}

	s.log.Debug("Successfully saved data for today")
	return nil
}

func (s *Storage) GetPopularLocations(limit int) ([]map[string]interface{}, error) {
	query := `
	SELECT lat, lng, distance, count, last_search
	FROM search_locations
	ORDER BY count DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying popular locations: %w", err)
	}
	defer rows.Close()

	var popularLocations []map[string]interface{}
	for rows.Next() {
		var lat, lng, distance float64
		var count int
		var lastSearch string

		err := rows.Scan(&lat, &lng, &distance, &count, &lastSearch)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		location := map[string]interface{}{
			"lat":         lat,
			"lng":         lng,
			"distance":    distance,
			"count":       count,
			"last_search": lastSearch,
		}
		popularLocations = append(popularLocations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return popularLocations, nil
}

func ParseLatLong(s string) (float64, error) {
	s = strings.Replace(s, ",", ".", 1)
	m, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return m, nil
}

func (s *Storage) UpdateDBAll() error {
	api := api.NewFuelPriceAPI()

	startDate := time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Now().AddDate(0, 0, -1)

	for date := startDate; date.Before(endDate) || date.Equal(endDate); date = date.AddDate(0, 0, 1) {
		exists, err := s.HasDate(date)
		if err != nil {
			s.log.Debug("Error checking if date exists", "date", date.Format("2006-01-02"), "error", err)
			continue
		}

		if exists {
			continue
		}

		s.log.Debug("Fetching data for", "date", date.Format("2006-01-02"))

		pricesResponse, err := api.FetchPricesForDate(date)
		if err != nil {
			s.log.Debug("Error fetching data for", "date", date.Format("2006-01-02"), "error", err)
			continue
		}

		if pricesResponse.ResultadoConsulta != "OK" {
			s.log.Debug("API returned non-OK result for", "date", date.Format("2006-01-02"), "result", pricesResponse.ResultadoConsulta)
			continue
		}

		jsonData, err := json.Marshal(pricesResponse)
		if err != nil {
			s.log.Debug("Error marshaling JSON for", "date", date.Format("2006-01-02"), "error", err)
			continue
		}

		if err := s.SavePrices(date, jsonData); err != nil {
			s.log.Debug("Error saving data for", "date", date.Format("2006-01-02"), "error", err)
			continue
		}

		s.log.Debug("Successfully saved data for", "date", date.Format("2006-01-02"))

		time.Sleep(200 * time.Millisecond)
	}

	// Save today's prices
	pricesResponse, err := api.FetchPrices()
	if err != nil {
		return err
	}

	if pricesResponse.ResultadoConsulta != "OK" {
		return fmt.Errorf("API returned non-OK result: %s", pricesResponse.ResultadoConsulta)
	}

	jsonData, err := json.Marshal(pricesResponse)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %w", err)
	}

	if err := s.SavePrices(endDate.AddDate(0, 0, 1), jsonData); err != nil {
		return fmt.Errorf("Error saving data for today: %w", err)
	}

	log.Printf("Successfully saved data for today")
	return nil
}

func (s *Storage) CreateLocationLogsTable() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS location_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		distance REAL NOT NULL,
		search_count INTEGER NOT NULL DEFAULT 1,
		search_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_search TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Index for faster searches on location coordinates
	CREATE INDEX IF NOT EXISTS idx_location_logs_coordinates ON location_logs (latitude, longitude);
	`

	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating location_logs table: %w", err)
	}

	s.log.Debug("Location logs table created or verified")
	return nil
}

func reduceLocationPrecision(lat, lng float64, decimalPlaces int) (float64, float64) {
	factor := math.Pow(10, float64(decimalPlaces))
	return math.Round(lat*factor) / factor, math.Round(lng*factor) / factor
}

func (s *Storage) LogSearchLocation(latitude, longitude, distance float64) error {
	// First check if a similar location (with small tolerance) exists
	const tolerance = 0.0001 // Approximately 10 meters

	var id int64
	var count int

	newLat, newLong := reduceLocationPrecision(latitude, longitude, 2)
	err := s.db.QueryRow(`
		SELECT id, search_count FROM location_logs
		WHERE latitude = ?
		AND longitude = ?
		ORDER BY ABS(latitude - ?) + ABS(longitude - ?) ASC
		LIMIT 1
	`, newLat, newLong, newLat, newLong).Scan(&id, &count)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking for existing location: %w", err)
	}

	if err == sql.ErrNoRows {
		// Insert new location
		_, err := s.db.Exec(`
			INSERT INTO location_logs (latitude, longitude, distance)
			VALUES (?, ?, ?)
		`, latitude, longitude, distance)

		if err != nil {
			return fmt.Errorf("error logging search location: %w", err)
		}
	} else {
		// Update existing location
		_, err := s.db.Exec(`
			UPDATE location_logs
			SET search_count = search_count + 1, last_search = CURRENT_TIMESTAMP, distance = ?
			WHERE id = ?
		`, distance, id)

		if err != nil {
			return fmt.Errorf("error updating search location: %w", err)
		}
	}

	return nil
}

// LocationLog represents a row in the location_logs table
type LocationLog struct {
	ID          int64
	Latitude    float64
	Longitude   float64
	Distance    float64
	SearchCount int64
	SearchTime  time.Time
	LastSearch  time.Time
}

// GetLocationLogs retrieves location logs from the database
// limit: maximum number of rows to return (0 for all)
// orderBy: "count" for most searched or "time" for most recent
func (s *Storage) GetLocationLogs(limit int, orderBy string) ([]LocationLog, error) {
	query := `SELECT id, latitude, longitude, distance, search_count, search_time, last_search
			  FROM location_logs `

	// Add ordering
	switch orderBy {
	case "count":
		query += "ORDER BY search_count DESC "
	case "time":
		query += "ORDER BY last_search DESC "
	default:
		query += "ORDER BY id DESC "
	}

	// Add limit
	if limit > 0 {
		query += fmt.Sprintf("LIMIT %d", limit)
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error retrieving location logs: %w", err)
	}
	defer rows.Close()

	var logs []LocationLog
	for rows.Next() {
		var log LocationLog
		if err := rows.Scan(
			&log.ID,
			&log.Latitude,
			&log.Longitude,
			&log.Distance,
			&log.SearchCount,
			&log.SearchTime,
			&log.LastSearch,
		); err != nil {
			return nil, fmt.Errorf("error scanning location log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return logs, nil
}

// PopularLocation represents a clustered area of searches with its popularity
type PopularLocation struct {
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lng"`
	SearchCount int64   `json:"weight"` // Used as weight in heatmaps
	Radius      float64 `json:"radius"` // Estimated radius of the cluster in km
}

// GetPopularLocationHeatmap returns data suitable for generating a heatmap
// of popular search locations, with nearby searches clustered together
func (s *Storage) GetPopularLocationHeatmap() ([]PopularLocation, error) {
	// Get all location logs ordered by search count
	logs, err := s.GetLocationLogs(0, "count")
	if err != nil {
		return nil, err
	}

	// Clustering parameters
	const clusterDistance = 0.01 // Approximately 1km

	// Map to track processed logs
	processed := make(map[int64]bool)

	var popularLocations []PopularLocation

	// Process logs and create clusters
	for i, log := range logs {
		if processed[log.ID] {
			continue
		}

		// Mark this log as processed
		processed[log.ID] = true

		// Create a new cluster with this log as center
		cluster := PopularLocation{
			Latitude:    log.Latitude,
			Longitude:   log.Longitude,
			SearchCount: log.SearchCount,
			Radius:      log.Distance, // Start with search distance as radius
		}

		// Check for nearby logs to merge into this cluster
		for j, otherLog := range logs {
			if i == j || processed[otherLog.ID] {
				continue
			}

			// Calculate distance between points
			distance := math.Sqrt(
				math.Pow(log.Latitude-otherLog.Latitude, 2) +
					math.Pow(log.Longitude-otherLog.Longitude, 2))

			if distance <= clusterDistance {
				// Merge this log into the cluster
				processed[otherLog.ID] = true

				// Update cluster properties based on weighted average
				totalWeight := cluster.SearchCount + otherLog.SearchCount
				cluster.Latitude = (cluster.Latitude*float64(cluster.SearchCount) +
					otherLog.Latitude*float64(otherLog.SearchCount)) / float64(totalWeight)
				cluster.Longitude = (cluster.Longitude*float64(cluster.SearchCount) +
					otherLog.Longitude*float64(otherLog.SearchCount)) / float64(totalWeight)

				// Update search count and expand radius if needed
				cluster.SearchCount += otherLog.SearchCount
				if otherLog.Distance > cluster.Radius {
					cluster.Radius = otherLog.Distance
				}
			}
		}

		popularLocations = append(popularLocations, cluster)
	}

	// Sort by search count (most popular first)
	sort.Slice(popularLocations, func(i, j int) bool {
		return popularLocations[i].SearchCount > popularLocations[j].SearchCount
	})

	return popularLocations, nil
}
