package translations

// Translations contains all text strings for the application
type Translations struct {
	// Page titles
	HomeTitle    string
	ResultsTitle string

	// Home page
	HomeHeading             string
	LastUpdated             string
	LocationLabel           string
	LocationPlaceholder     string
	LocationExample         string
	SearchButton            string
	UseLocationButton       string
	GeolocationNotSupported string

	// Geolocation messages
	RequestingLocation  string
	LocationFound       string
	PermissionDenied    string
	LocationUnavailable string
	LocationTimeout     string
	UnknownError        string

	// Results page
	NearbyStations   string
	ResultsFor       string
	ResultsForCoords string
	SearchRadius     string
	NewSearchButton  string
	NoStationsFound  string
	LocationNotFound string
	StationsFound    string
	StationsWithin   string
	OfYourLocation   string

	// Station card
	MapButton        string
	GoogleMapsButton string
	MapAltOSM        string
	MapAltGoogle     string
	KmAway           string
	Gasoline95       string
	Gasoline98       string
	Diesel           string
	PremiumDiesel    string
	NotAvailable     string

	// Footer
	FooterCopyright string
}

// GetTranslations returns translations for the specified language
func GetTranslations(lang string) Translations {
	switch lang {
	case "en", "english":
		return GetEnglishTranslations()
	default:
		return GetSpanishTranslations()
	}
}

// GetLanguageFromQuery extracts language from query parameter, defaults to Spanish
func GetLanguageFromQuery(langParam string) string {
	switch langParam {
	case "en", "english":
		return "en"
	default:
		return "es"
	}
}
