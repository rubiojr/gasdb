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
		LocationPlaceholder:     "Ciudad, provincia (solo España)",
		LocationExample:         "Ejemplo: Tibidabo, Barcelona",
		SearchButton:            "Buscar",
		UseLocationButton:       "Usar Mi Ubicación",
		GeolocationNotSupported: "Geolocalización no soportada",

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
		ResultsForCoords:          "Resultados para coordenadas:",
		SearchRadius:              "Radio de búsqueda:",
		NewSearchButton:           "Nueva Búsqueda",
		NoStationsFound:           "No se encontraron gasolineras en un radio de",
		LocationNotFound:          "Ubicación no encontrada.",
		LocationSearchUnavailable: "El servicio de búsqueda de ubicaciones no está disponible temporalmente. Inténtalo de nuevo en unos momentos.",
		StationsFound:             "Se encontraron",
		StationsWithin:            "estaciones en un radio de",
		OfYourLocation:            "de tu ubicación.",

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
		FooterCopyright: "Buscador de Gasolineras España",
	}
}
