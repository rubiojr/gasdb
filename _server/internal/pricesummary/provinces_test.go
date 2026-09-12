package pricesummary

import (
	"testing"
	"time"

	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvinceRankingsUsePerFuelAverages(t *testing.T) {
	data := &api.GasStationList{ListaEESSPrecio: []api.GasStation{
		{Provincia: "Soria", PrecioGasolina95E5: "1,200", PrecioGasolina98E5: "2,000", PrecioGasoleoA: "1,000", PrecioGasoleoPremium: "1,600"},
		{Provincia: " soria ", PrecioGasolina95E5: "1,600", PrecioGasolina98E5: "-", PrecioGasoleoA: "1,200", PrecioGasoleoPremium: "NaN"},
		{Provincia: "Barcelona", PrecioGasolina95E5: "1,500", PrecioGasolina98E5: "1,100", PrecioGasoleoA: "1,400", PrecioGasoleoPremium: "1,800"},
		{Provincia: "Madrid", PrecioGasolina95E5: "1,800", PrecioGasolina98E5: "2,000", PrecioGasoleoA: "1,800"},
		{Provincia: "Burgos", PrecioGasolina95E5: "1,600", PrecioGasolina98E5: "1,900"},
		{Provincia: "Palencia", PrecioGasolina95E5: "1,900"},
		{Provincia: "", PrecioGasolina95E5: "0,500"},
		{Provincia: "INVALID", PrecioGasolina95E5: "0", PrecioGasolina98E5: "NaN", PrecioGasoleoA: "-1", PrecioGasoleoPremium: "+Inf"},
	}}
	summary := Build(data, time.Now())
	ranking := summary.Provinces.Gasoline95
	require.Len(t, ranking.Cheapest, 2)
	require.Len(t, ranking.MostExpensive, 2)
	assert.Equal(t, "SORIA", ranking.Cheapest[0].Province)
	assert.InDelta(t, 1.4, ranking.Cheapest[0].Average, 1e-12)
	assert.Equal(t, 2, ranking.Cheapest[0].Count)
	assert.Equal(t, ProvincePrice{Province: "BARCELONA", Average: 1.5, Count: 1}, ranking.Cheapest[1])
	assert.Equal(t, []ProvincePrice{
		{Province: "PALENCIA", Average: 1.9, Count: 1},
		{Province: "MADRID", Average: 1.8, Count: 1},
	}, ranking.MostExpensive)
	assert.Equal(t, "BARCELONA", summary.Provinces.Gasoline98.Cheapest[0].Province)
	assert.Equal(t, "MADRID", summary.Provinces.Gasoline98.MostExpensive[0].Province)
	assert.Equal(t, "SORIA", summary.Provinces.Gasoline98.MostExpensive[1].Province)
	assert.InDelta(t, 1.1, summary.Provinces.Diesel.Cheapest[0].Average, 1e-12)
	assert.Equal(t, 2, summary.Provinces.Diesel.Cheapest[0].Count)
	assert.Equal(t, "BARCELONA", summary.Provinces.PremiumDiesel.MostExpensive[0].Province)
	assert.Equal(t, 1, summary.Provinces.PremiumDiesel.Cheapest[0].Count)
	// A missing province excludes a record only from the province comparison.
	assert.Equal(t, 0.5, summary.Gasoline95.Low)
	assert.Equal(t, 7, summary.Gasoline95.Count)
	assert.Equal(t, " soria ", data.ListaEESSPrecio[1].Provincia, "the cached source snapshot must not be mutated")
}

func TestProvinceRankingsHandleSparseData(t *testing.T) {
	assert.Equal(t, ProvinceRankings{}, Build(nil, time.Now()).Provinces)
	data := &api.GasStationList{ListaEESSPrecio: []api.GasStation{
		{Provincia: "", PrecioGasolina95E5: "1,000"},
		{Provincia: "Soria", PrecioGasolina95E5: "-"},
	}}
	assert.Equal(t, ProvinceRankings{}, Build(data, time.Now()).Provinces)
	data.ListaEESSPrecio = append(data.ListaEESSPrecio, api.GasStation{Provincia: "Soria", PrecioGasolina95E5: "1,250"})
	ranking := Build(data, time.Now()).Provinces.Gasoline95
	want := []ProvincePrice{{Province: "SORIA", Average: 1.25, Count: 1}}
	assert.Equal(t, want, ranking.Cheapest)
	assert.Equal(t, want, ranking.MostExpensive)
}

func TestProvinceRankingDoesNotRoundBeforeSorting(t *testing.T) {
	data := &api.GasStationList{ListaEESSPrecio: []api.GasStation{
		{Provincia: "Z", PrecioGasolina95E5: "1,233"},
		{Provincia: "Z", PrecioGasolina95E5: "1,234"},
		{Provincia: "Z", PrecioGasolina95E5: "1,235"},
		{Provincia: "A", PrecioGasolina95E5: "1,234"},
		{Provincia: "A", PrecioGasolina95E5: "1,234"},
		{Provincia: "A", PrecioGasolina95E5: "1,235"},
	}}
	ranking := Build(data, time.Now()).Provinces.Gasoline95
	assert.Equal(t, "Z", ranking.Cheapest[0].Province)
	assert.Equal(t, "A", ranking.MostExpensive[0].Province)
}

func TestProvinceRankingBreaksExactTiesAlphabetically(t *testing.T) {
	data := &api.GasStationList{}
	for _, name := range []string{"D", "B", "C", "A"} {
		data.ListaEESSPrecio = append(data.ListaEESSPrecio, api.GasStation{Provincia: name, PrecioGasolina95E5: "1,500"})
	}
	ranking := Build(data, time.Now()).Provinces.Gasoline95
	assert.Equal(t, "A", ranking.Cheapest[0].Province)
	assert.Equal(t, "B", ranking.Cheapest[1].Province)
	assert.Equal(t, "A", ranking.MostExpensive[0].Province)
	assert.Equal(t, "B", ranking.MostExpensive[1].Province)
}
