package translations

// GetEnglishTranslations returns all English text strings
func GetEnglishTranslations() Translations {
	return Translations{
		// Page titles
		HomeTitle:    "Fuel Station Finder - Home",
		ResultsTitle: "Fuel Station Results",

		// Home page
		HomeHeading:             "🔎 Nearby Fuel Stations",
		LastUpdated:             "📅 Fuel prices last updated:",
		LocationLabel:           "Location",
		LocationPlaceholder:     "City, province (Spain only)",
		LocationExample:         "Example: Tibidabo, Barcelona",
		SearchButton:            "Search",
		UseLocationButton:       "Use My Location",
		GeolocationNotSupported: "Geolocation not supported",

		// Geolocation messages
		RequestingLocation:  "Requesting your location...",
		LocationFound:       "Location found! Searching nearby stations...",
		PermissionDenied:    "Location permission denied.",
		LocationUnavailable: "Location information is unavailable.",
		LocationTimeout:     "Location request timed out.",
		UnknownError:        "An unknown error occurred.",

		// Results page
		NearbyStations:   "Nearby Fuel Stations",
		ResultsFor:       "Results for:",
		ResultsForCoords: "Results for coordinates:",
		SearchRadius:     "Search radius:",
		NewSearchButton:  "New Search",
		NoStationsFound:  "No fuel stations found within",
		LocationNotFound: "Location not found.",
		StationsFound:    "Found",
		StationsWithin:   "stations within",
		OfYourLocation:   "of your location.",

		// Station card
		MapButton:        "🗺️ OSM",
		GoogleMapsButton: "📍 Google Maps",
		MapAltOSM:        "View on OpenStreetMap",
		MapAltGoogle:     "View on Google Maps",
		KmAway:           "km away",
		Gasoline95:       "Gasoline 95:",
		Gasoline98:       "Gasoline 98:",
		Diesel:           "Diesel:",
		PremiumDiesel:    "Premium Diesel:",
		NotAvailable:     "N/A",

		// Footer
		FooterCopyright: "Fuel Station Finder Spain",
	}
}
