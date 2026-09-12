// Package pricesummary calculates nationwide fuel price ranges.
package pricesummary

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rubiojr/gasdb/pkg/api"
)

// Prices contains prices in euros per litre, counting only valid positive prices.
type Prices struct {
	Low, Average, High float64
	Count              int
}

// Summary describes the latest published snapshot, which may predate today.
type Summary struct {
	Date                                          string
	Today, Available                              bool
	Gasoline95, Gasoline98, Diesel, PremiumDiesel Prices
	Provinces                                     ProvinceRankings
}

func Build(data *api.GasStationList, now time.Time) Summary {
	if data == nil {
		return Summary{}
	}
	summary := Summary{Date: data.Fecha, Today: isToday(data.Fecha, now)}
	for _, station := range data.ListaEESSPrecio {
		summary.Gasoline95.add(station.PrecioGasolina95E5)
		summary.Gasoline98.add(station.PrecioGasolina98E5)
		summary.Diesel.add(station.PrecioGasoleoA)
		summary.PremiumDiesel.add(station.PrecioGasoleoPremium)
	}
	summary.Available = summary.Gasoline95.Count+summary.Gasoline98.Count+summary.Diesel.Count+summary.PremiumDiesel.Count > 0
	summary.Provinces = buildProvinceRankings(data.ListaEESSPrecio)
	return summary
}

func (p *Prices) add(raw string) {
	price, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(raw), ",", "."), 64)
	if err != nil || price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return
	}
	if p.Count == 0 {
		p.Low, p.High = price, price
	}
	p.Low = min(p.Low, price)
	p.High = max(p.High, price)
	p.Count++
	// An incremental mean avoids overflowing a sum of valid finite prices.
	p.Average += (price - p.Average) / float64(p.Count)
}

func isToday(date string, now time.Time) bool {
	zone, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		return false
	}
	published, err := time.ParseInLocation("02/01/2006 15:04:05", date, zone)
	return err == nil && published.Format(time.DateOnly) == now.In(zone).Format(time.DateOnly)
}
