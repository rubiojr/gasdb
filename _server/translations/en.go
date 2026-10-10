package translations

// GetEnglishTranslations returns all English text strings
func GetEnglishTranslations() Translations {
	return Translations{
		Language: "en",
		// Page titles
		HomeTitle:    "GasDB — Fuel prices in Spain",
		ResultsTitle: "GasDB — Search results",

		// Home page
		HomeHeading:             "GasDB",
		BrandEyebrow:            "Fuel prices in Spain",
		LastUpdated:             "📅 Fuel prices last updated:",
		LocationLabel:           "Location",
		RadiusLabel:             "Radius",
		LocationPlaceholder:     "City or address",
		LocationExample:         "Example: Tibidabo, Barcelona",
		SearchButton:            "Search",
		UseLocationButton:       "Use My Location",
		GeolocationNotSupported: "Geolocation not supported",

		TodayPrices:               "Today's prices in Spain",
		LatestPrices:              "Latest prices in Spain",
		PricesUnavailable:         "Fuel prices are not available yet.",
		PriceSummaryDescription:   "Lowest and highest are individual station prices across Spain. The national average uses all stations reporting a valid price for each fuel. Province rankings below compare averages, not individual station prices.",
		PriceDataDate:             "Data published:",
		PriceSummaryMethod:        "How are prices calculated?",
		FuelLabel:                 "Fuel",
		LowPrice:                  "Lowest station price",
		AveragePrice:              "National average",
		HighPrice:                 "Highest station price",
		ProvincesTitle:            "Provinces",
		ProvincePricesDescription: "Province averages for the selected fuel · €/L",
		ProvincePricesUnavailable: "No province price data for this fuel.",
		CheapestProvinces:         "Lowest province averages",
		MostExpensiveProvinces:    "Highest province averages",

		// Geolocation messages
		RequestingLocation:  "Requesting your location...",
		LocationFound:       "Location found! Searching nearby stations...",
		PermissionDenied:    "Location permission denied.",
		LocationUnavailable: "Location information is unavailable.",
		LocationTimeout:     "Location request timed out.",
		UnknownError:        "An unknown error occurred.",

		// Results page
		NearbyStations:            "Nearby Fuel Stations",
		ResultsFor:                "Results for:",
		ResolvedLocation:          "Location found:",
		ResultsForCoords:          "Results for coordinates:",
		SearchRadius:              "Search radius:",
		NewSearchButton:           "New Search",
		NoStationsFound:           "No fuel stations found within",
		LocationNotFound:          "Location not found.",
		LocationSearchUnavailable: "Location search is temporarily unavailable. Please try again in a moment.",
		StationsFound:             "Found",
		StationsWithin:            "stations within",
		OfYourLocation:            "of your location.",
		ExpandRadiusTitle:         "Try a wider search radius",
		SearchWithinRadius:        "Search within %d km",
		OneStation:                "1 station",
		StationCount:              "%d stations",

		// Results map
		ShowMapButton:  "🗺️ Show map",
		HideMapButton:  "🗺️ Hide map",
		MapLabel:       "Map of nearby stations",
		MapLoading:     "Loading map...",
		MapUnavailable: "The map could not be loaded. Please try again later.",
		SearchCenter:   "Search location",

		// Station card
		MapButton:         "🗺️ OSM",
		GoogleMapsButton:  "📍 Google Maps",
		MapAltOSM:         "View on OpenStreetMap",
		MapAltGoogle:      "View on Google Maps",
		StationPageButton: "View map",
		StationPageAlt:    "View this station's page with its map",
		ShareButton:       "🔗 Share",
		ShareAlt:          "Share a link to this station",
		LinkCopied:        "Link copied to the clipboard",
		KmAway:            "km away",
		Gasoline95:        "Gasoline 95:",
		Gasoline98:        "Gasoline 98:",
		Diesel:            "Diesel:",
		PremiumDiesel:     "Premium Diesel:",
		NotAvailable:      "N/A",

		// Station page
		StationNotFoundTitle: "GasDB — Station not found",
		StationNotFound:      "This fuel station was not found. It may have closed or no longer report prices.",
		StationMapLabel:      "Station location",
		StationPrices:        "Prices",
		OpeningHours:         "Opening hours",

		// Fuels only listed on the station page, when sold
		Gasoline95E10:     "Gasoline 95 E10",
		Gasoline95Premium: "Gasoline 95 Premium",
		Gasoline98E10:     "Gasoline 98 E10",
		DieselB:           "Diesel B",
		Biodiesel:         "Biodiesel",
		Bioethanol:        "Bioethanol",
		LPG:               "LPG",
		CNG:               "CNG",
		LNG:               "LNG",
		Hydrogen:          "Hydrogen",

		// Footer
		FooterCopyright:    "Fuel Station Finder Spain",
		DarkMode:           "Dark mode",
		SwitchToLightTheme: "Switch to light mode",
		SwitchToDarkTheme:  "Switch to dark mode",
	}
}
