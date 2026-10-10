package templates

import "github.com/rubiojr/gasdb/_server/translations"

// MapLibre GL JS, pinned to Subresource Integrity hashes.
const (
	mapLibreModule              = "https://cdn.jsdelivr.net/npm/maplibre-gl@6.13.0/dist/maplibre-gl.mjs"
	mapLibreModuleIntegrity     = "sha384-tG0qQdL7veKvVoyu6PbnmddDls/OVkTjaK9oVY5eq9afBA6zYptFui69Dx57duAh"
	mapLibreStylesheet          = "https://cdn.jsdelivr.net/npm/maplibre-gl@6.13.0/dist/maplibre-gl.css"
	mapLibreStylesheetIntegrity = "sha384-yFNc3bs28S14jOJo1ZO9SnLiAmtyY2a3H9be8GkLMpDGlO5r1zZNP8fYu+fOCE/w"
)

// importMap lets page scripts import "maplibre-gl", checked against its hash.
type importMap struct {
	Imports   map[string]string `json:"imports"`
	Integrity map[string]string `json:"integrity"`
}

var mapLibreImports = importMap{
	Imports:   map[string]string{"maplibre-gl": mapLibreModule},
	Integrity: map[string]string{mapLibreModule: mapLibreModuleIntegrity},
}

// mapText holds the strings map scripts show, MapLibre's own included.
type mapText struct {
	Show         string            `json:"show"`
	Hide         string            `json:"hide"`
	Loading      string            `json:"loading"`
	Unavailable  string            `json:"unavailable"`
	SearchCenter string            `json:"search_center"`
	NotAvailable string            `json:"not_available"`
	Locale       map[string]string `json:"locale,omitempty"`
}

func newMapText(t translations.Translations) mapText {
	return mapText{
		Show:         t.ShowMapButton,
		Hide:         t.HideMapButton,
		Loading:      t.MapLoading,
		Unavailable:  t.MapUnavailable,
		SearchCenter: t.SearchCenter,
		NotAvailable: t.NotAvailable,
		Locale:       t.MapLocale,
	}
}
