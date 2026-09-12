package translations

// GetSpanishTranslations returns all Spanish text strings
func GetSpanishTranslations() Translations {
	return Translations{
		// Page titles
		HomeTitle:    "Buscador de Gasolineras - Inicio",
		ResultsTitle: "Resultados de Gasolineras",

		// Home page
		HomeHeading:             "🔎 Gasolineras Cercanas",
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
		PriceSummaryDescription:   "Precios mínimos, medios y máximos en España, en €/L. La media se calcula entre las gasolineras que publican un precio para cada carburante.",
		PriceDataDate:             "Datos publicados:",
		PriceSummaryMethod:        "¿Cómo se calculan los precios?",
		FuelLabel:                 "Carburante",
		LowPrice:                  "Mínimo",
		AveragePrice:              "Media",
		HighPrice:                 "Máximo",
		ProvincesTitle:            "Provincias",
		ProvincePricesDescription: "Precio medio por provincia · €/L",
		ProvincePricesUnavailable: "Sin datos provinciales para este carburante.",
		CheapestProvinces:         "Más baratas",
		MostExpensiveProvinces:    "Más caras",

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
