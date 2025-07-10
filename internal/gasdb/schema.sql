CREATE TABLE fuel_prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT UNIQUE NOT NULL,
		data BLOB NOT NULL
	);
CREATE TABLE sqlite_sequence(name,seq);
CREATE INDEX idx_fuel_prices_date ON fuel_prices(date);
CREATE TABLE historic_prices (
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
CREATE INDEX idx_historic_prices_date ON historic_prices(date);
CREATE INDEX idx_historic_prices_ideess ON historic_prices(ideess);
CREATE INDEX idx_historic_prices_latitud_longitud ON historic_prices(latitud, longitud);
CREATE TRIGGER insert_historic_prices
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
CREATE TABLE location_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		distance REAL NOT NULL,
		search_count INTEGER NOT NULL DEFAULT 1,
		search_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_search TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
CREATE INDEX idx_location_logs_coordinates ON location_logs (latitude, longitude);
