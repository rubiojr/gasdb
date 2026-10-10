package templates

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
