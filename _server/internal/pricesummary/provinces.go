package pricesummary

import (
	"sort"
	"strings"

	"github.com/rubiojr/gasdb/pkg/api"
)

type ProvincePrice struct {
	Province string
	Average  float64 // euros per litre, averaged across reporting stations
	Count    int
}

// ProvinceRanking contains up to two provinces at each end of the ranking.
type ProvinceRanking struct {
	Cheapest, MostExpensive []ProvincePrice
}

type ProvinceRankings struct {
	Gasoline95, Gasoline98, Diesel, PremiumDiesel ProvinceRanking
}

func buildProvinceRankings(stations []api.GasStation) ProvinceRankings {
	var fuels [4]map[string]*Prices
	for i := range fuels {
		fuels[i] = make(map[string]*Prices)
	}
	for _, station := range stations {
		province := strings.ToUpper(strings.Join(strings.Fields(station.Provincia), " "))
		if province == "" {
			continue
		}
		values := [...]string{station.PrecioGasolina95E5, station.PrecioGasolina98E5, station.PrecioGasoleoA, station.PrecioGasoleoPremium}
		for i, price := range values {
			if fuels[i][province] == nil {
				fuels[i][province] = &Prices{}
			}
			fuels[i][province].add(price)
		}
	}
	return ProvinceRankings{
		Gasoline95:    rankProvinces(fuels[0]),
		Gasoline98:    rankProvinces(fuels[1]),
		Diesel:        rankProvinces(fuels[2]),
		PremiumDiesel: rankProvinces(fuels[3]),
	}
}

func rankProvinces(provinces map[string]*Prices) ProvinceRanking {
	var prices []ProvincePrice
	for name, stats := range provinces {
		if stats.Count > 0 {
			prices = append(prices, ProvincePrice{Province: name, Average: stats.Average, Count: stats.Count})
		}
	}
	if len(prices) == 0 {
		return ProvinceRanking{}
	}
	// Rank using unrounded averages; alphabetical order makes exact ties stable.
	sort.Slice(prices, func(i, j int) bool {
		if prices[i].Average == prices[j].Average {
			return prices[i].Province < prices[j].Province
		}
		return prices[i].Average < prices[j].Average
	})
	limit := min(2, len(prices))
	cheapest := append([]ProvincePrice(nil), prices[:limit]...)
	sort.SliceStable(prices, func(i, j int) bool { return prices[i].Average > prices[j].Average })
	return ProvinceRanking{
		Cheapest:      cheapest,
		MostExpensive: append([]ProvincePrice(nil), prices[:limit]...),
	}
}
