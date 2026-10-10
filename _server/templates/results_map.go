package templates

import "github.com/rubiojr/gasdb/_server/translations"

// ResultsMap is embedded in the results page as JSON and drawn by its map.
type ResultsMap struct {
	Style    string       `json:"style"`  // MapLibre style URL, light in every theme
	Center   [2]float64   `json:"center"` // [lng, lat], MapLibre's order
	RadiusKM float64      `json:"radius_km"`
	Stations []MapStation `json:"stations"` // in result order, matching the cards
}

// MapStation is one station marker.
type MapStation struct {
	Lng  float64 `json:"lng"`
	Lat  float64 `json:"lat"`
	Name string  `json:"name"`
	// Price is the searched fuel's price; nil when the station doesn't sell it.
	Price *float64 `json:"price"`
}

type resultsMapText struct {
	Show         string            `json:"show"`
	Hide         string            `json:"hide"`
	Loading      string            `json:"loading"`
	Unavailable  string            `json:"unavailable"`
	SearchCenter string            `json:"search_center"`
	NotAvailable string            `json:"not_available"`
	Locale       map[string]string `json:"locale,omitempty"`
}

func newResultsMapText(t translations.Translations) resultsMapText {
	return resultsMapText{
		Show:         t.ShowMapButton,
		Hide:         t.HideMapButton,
		Loading:      t.MapLoading,
		Unavailable:  t.MapUnavailable,
		SearchCenter: t.SearchCenter,
		NotAvailable: t.NotAvailable,
		Locale:       t.MapLocale,
	}
}
