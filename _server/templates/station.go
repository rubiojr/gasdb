package templates

import (
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/pkg/api"
)

// StationMap is embedded in the station page as JSON and drawn by its map.
type StationMap struct {
	Style string  `json:"style"` // MapLibre style URL, light in every theme
	Lng   float64 `json:"lng"`
	Lat   float64 `json:"lat"`
}

type fuelPrice struct {
	Label string
	Price string
}

// stationURL is the station page's path, in the current language.
func stationURL(station *api.GasStation, t translations.Translations) templ.SafeURL {
	return templ.SafeURL("/station/" + url.PathEscape(station.IDEESS) + "?lang=" + t.Language)
}

// nearbyURL searches for stations around this one.
func nearbyURL(station *api.GasStation, t translations.Translations) templ.SafeURL {
	query := url.Values{
		"lat":  {formatDecimal(station.Latitud)},
		"lng":  {formatDecimal(station.Longitud)},
		"lang": {t.Language},
	}
	return templ.SafeURL("/search?" + query.Encode())
}

// stationTitle names a station by its brand and town.
func stationTitle(station *api.GasStation) string {
	if station.Localidad == "" {
		return station.Rotulo
	}
	return station.Rotulo + ", " + station.Localidad
}

// stationLocality is the postal code and town, plus the province when it
// differs from the town.
func stationLocality(station *api.GasStation) string {
	locality := strings.TrimSpace(station.CP + " " + station.Localidad)
	if station.Provincia == "" || strings.EqualFold(station.Provincia, station.Localidad) {
		return locality
	}
	return strings.TrimSpace(locality + " (" + station.Provincia + ")")
}

func mainPrices(station *api.GasStation, t translations.Translations) []fuelPrice {
	return []fuelPrice{
		{t.Gasoline95, station.PrecioGasolina95E5},
		{t.Gasoline98, station.PrecioGasolina98E5},
		{t.Diesel, station.PrecioGasoleoA},
		{t.PremiumDiesel, station.PrecioGasoleoPremium},
	}
}

// extraPrices lists the fuels result cards leave out that the station sells.
func extraPrices(station *api.GasStation, t translations.Translations) []fuelPrice {
	var prices []fuelPrice
	for _, price := range []fuelPrice{
		{t.Gasoline95E10, station.PrecioGasolina95E10},
		{t.Gasoline95Premium, station.PrecioGasolina95E5Prem},
		{t.Gasoline98E10, station.PrecioGasolina98E10},
		{t.DieselB, station.PrecioGasoleoB},
		{t.Biodiesel, station.PrecioBiodiesel},
		{t.Bioethanol, station.PrecioBioetanol},
		{t.LPG, station.PrecioGasesLicuados},
		{t.CNG, station.PrecioGasNaturalComp},
		{t.LNG, station.PrecioGasNaturalLicuado},
		{t.Hydrogen, station.PrecioHidrogeno},
	} {
		if price.Price != "" {
			prices = append(prices, price)
		}
	}
	return prices
}

// stationDescription summarizes the station for link previews.
func stationDescription(station *api.GasStation, t translations.Translations) string {
	parts := []string{strings.Trim(station.Direccion+", "+stationLocality(station), ", ")}
	for _, price := range mainPrices(station, t) {
		if price.Price != "" {
			parts = append(parts, strings.TrimSuffix(price.Label, ":")+": "+formatPrice(price.Price, ""))
		}
	}
	return strings.Join(parts, " · ")
}
