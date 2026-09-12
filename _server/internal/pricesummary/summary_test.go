package pricesummary

import (
	"testing"
	"time"

	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestBuildPriceRanges(t *testing.T) {
	data := &api.GasStationList{
		Fecha: "12/09/2026 12:00:00",
		ListaEESSPrecio: []api.GasStation{
			{PrecioGasolina95E5: "1,200", PrecioGasolina98E5: "1,800", PrecioGasoleoA: "1,100", PrecioGasoleoPremium: "1,500"},
			{PrecioGasolina95E5: " 1.600 ", PrecioGasolina98E5: "", PrecioGasoleoA: "1,300"},
			{PrecioGasolina95E5: "1,800", PrecioGasolina98E5: "2,000", PrecioGasoleoA: "-", PrecioGasoleoPremium: "0"},
		},
	}
	summary := Build(data, time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC))
	assert.True(t, summary.Today)
	assert.True(t, summary.Available)
	assert.Equal(t, data.Fecha, summary.Date)
	assert.Equal(t, 3, summary.Gasoline95.Count)
	assert.Equal(t, 1.2, summary.Gasoline95.Low)
	assert.InDelta(t, 4.6/3, summary.Gasoline95.Average, 1e-12)
	assert.Equal(t, 1.8, summary.Gasoline95.High)
	assert.Equal(t, 2, summary.Gasoline98.Count)
	assert.InDelta(t, 1.9, summary.Gasoline98.Average, 1e-12)
	assert.Equal(t, 2, summary.Diesel.Count)
	assert.InDelta(t, 1.2, summary.Diesel.Average, 1e-12)
	assert.Equal(t, Prices{Low: 1.5, Average: 1.5, High: 1.5, Count: 1}, summary.PremiumDiesel)
}

func TestBuildSkipsInvalidPrices(t *testing.T) {
	data := &api.GasStationList{}
	for _, raw := range []string{"", " ", "-", "0", "0,000", "-1,2", "broken", "NaN", "+Inf", "-Inf", "1e999"} {
		data.ListaEESSPrecio = append(data.ListaEESSPrecio, api.GasStation{PrecioGasolina95E5: raw})
	}
	summary := Build(data, time.Now())
	assert.False(t, summary.Available)
	assert.Equal(t, Prices{}, summary.Gasoline95)
	data.ListaEESSPrecio = append(data.ListaEESSPrecio, api.GasStation{PrecioGasolina95E5: "1,499"})
	summary = Build(data, time.Now())
	assert.Equal(t, Prices{Low: 1.499, Average: 1.499, High: 1.499, Count: 1}, summary.Gasoline95)
	assert.Equal(t, Summary{}, Build(nil, time.Now()))
}

func TestSummaryUsesSpanishPublicationDate(t *testing.T) {
	// It is already September 13 in mainland Spain, but still September 12 UTC.
	now := time.Date(2026, 9, 12, 22, 30, 0, 0, time.UTC)
	for _, tt := range []struct {
		date  string
		today bool
	}{
		{"13/09/2026 00:15:00", true},
		{"12/09/2026 23:00:00", false},
		{"14/09/2026 00:00:00", false},
		{"invalid", false},
		{"", false},
	} {
		t.Run(tt.date, func(t *testing.T) {
			summary := Build(&api.GasStationList{Fecha: tt.date}, now)
			assert.Equal(t, tt.today, summary.Today)
		})
	}
}
