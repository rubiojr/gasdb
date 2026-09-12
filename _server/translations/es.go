package translations

// GetSpanishTranslations returns all Spanish text strings
func GetSpanishTranslations() Translations {
	return Translations{
		Language: "es",
		// Page titles
		HomeTitle:    "GasDB — Precios de carburantes en España",
		ResultsTitle: "GasDB — Resultados",

		// Home page
		HomeHeading:             "GasDB",
		BrandEyebrow:            "Carburantes en España",
		LastUpdated:             "📅 Precios actualizados el:",
		LocationLabel:           "Ubicación",
		RadiusLabel:             "Radio",
		LocationPlaceholder:     "Ciudad o dirección",
		LocationExample:         "Ejemplo: Tibidabo, Barcelona",
		SearchButton:            "Buscar",
		UseLocationButton:       "Usar Mi Ubicación",
		GeolocationNotSupported: "Geolocalización no soportada",

		TodayPrices:               "Precios de hoy en España",
		LatestPrices:              "Últimos precios en España",
		PricesUnavailable:         "Los precios de los carburantes todavía no están disponibles.",
		PriceSummaryDescription:   "El mínimo y el máximo son precios de gasolineras individuales de toda España. La media nacional usa todas las gasolineras que publican un precio válido para cada carburante. La comparación provincial de abajo usa medias, no precios individuales.",
		PriceDataDate:             "Datos publicados:",
		PriceSummaryMethod:        "¿Cómo se calculan los precios?",
		FuelLabel:                 "Carburante",
		LowPrice:                  "Mínimo por estación",
		AveragePrice:              "Media nacional",
		HighPrice:                 "Máximo por estación",
		ProvincesTitle:            "Provincias",
		ProvincePricesDescription: "Medias provinciales del carburante seleccionado · €/L",
		ProvincePricesUnavailable: "Sin datos provinciales para este carburante.",
		CheapestProvinces:         "Medias provinciales más bajas",
		MostExpensiveProvinces:    "Medias provinciales más altas",

		// Geolocation messages
		RequestingLocation:  "Obteniendo tu ubicación...",
		LocationFound:       "¡Ubicación encontrada! Buscando gasolineras cercanas...",
		PermissionDenied:    "Permiso de ubicación denegado.",
		LocationUnavailable: "Información de ubicación no disponible.",
		LocationTimeout:     "Tiempo de espera de ubicación agotado.",
		UnknownError:        "Ocurrió un error desconocido.",

		// Results page
		NearbyStations:            "Gasolineras Cercanas",
		ResultsFor:                "Resultados para:",
		ResolvedLocation:          "Ubicación encontrada:",
		ResultsForCoords:          "Resultados para coordenadas:",
		SearchRadius:              "Radio de búsqueda:",
		NewSearchButton:           "Nueva Búsqueda",
		NoStationsFound:           "No se encontraron gasolineras en un radio de",
		LocationNotFound:          "Ubicación no encontrada.",
		LocationSearchUnavailable: "El servicio de búsqueda de ubicaciones no está disponible temporalmente. Inténtalo de nuevo en unos momentos.",
		StationsFound:             "Se encontraron",
		StationsWithin:            "estaciones en un radio de",
		OfYourLocation:            "de tu ubicación.",
		ExpandRadiusTitle:         "Ampliar el radio de búsqueda",
		SearchWithinRadius:        "Buscar en %d km",
		OneStation:                "1 gasolinera",
		StationCount:              "%d gasolineras",

		// Station card
		MapButton:        "🗺️ OSM",
		GoogleMapsButton: "📍 Google Maps",
		MapAltOSM:        "Ver en OpenStreetMap",
		MapAltGoogle:     "Ver en Google Maps",
		KmAway:           "km de distancia",
		Gasoline95:       "Gasolina 95:",
		Gasoline98:       "Gasolina 98:",
		Diesel:           "Diésel:",
		PremiumDiesel:    "Diésel Premium:",
		NotAvailable:     "N/D",

		// Footer
		FooterCopyright:    "Buscador de Gasolineras España",
		DarkMode:           "Modo oscuro",
		SwitchToLightTheme: "Cambiar a modo claro",
		SwitchToDarkTheme:  "Cambiar a modo oscuro",
	}
}
