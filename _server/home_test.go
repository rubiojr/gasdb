package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomePriceSummary(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	storage, err := gasdb.NewStorage(ctx, filepath.Join(t.TempDir(), "prices.db"), logger)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	render := func(lang string) string {
		t.Helper()
		response := httptest.NewRecorder()
		renderHome(response, httptest.NewRequest("GET", "/?lang="+lang, nil), storage, logger, translations.GetTranslations(lang))
		assert.Equal(t, 200, response.Code)
		return response.Body.String()
	}
	assert.Contains(t, render("en"), translations.GetEnglishTranslations().PricesUnavailable)
	data := api.GasStationList{
		Fecha: "01/01/2020 12:00:00",
		ListaEESSPrecio: []api.GasStation{
			{IDEESS: "1", Provincia: "Soria", PrecioGasolina95E5: "1,200"},
			{IDEESS: "2", Provincia: "Barcelona", PrecioGasolina95E5: "1,600"},
		},
	}
	save := func() {
		t.Helper()
		encoded, err := json.Marshal(data)
		require.NoError(t, err)
		require.NoError(t, storage.SavePrices(ctx, time.Now(), encoded))
	}
	save()
	for _, lang := range []string{"es", "en"} {
		page := render(lang)
		tr := translations.GetTranslations(lang)
		assert.Contains(t, page, `<html lang="`+lang+`">`)
		assert.Contains(t, page, `name="lang" value="`+lang+`"`)
		assert.Contains(t, page, tr.LatestPrices)
		assert.NotContains(t, page, tr.TodayPrices)
		assert.Contains(t, page, data.Fecha)
		assert.Contains(t, page, "<td>1.200</td><td>1.400</td><td>1.600</td>")
		assert.Contains(t, page, "<td>"+tr.NotAvailable+"</td>")
		assert.Contains(t, page, tr.CheapestProvinces)
		assert.Contains(t, page, tr.MostExpensiveProvinces)
		assert.Contains(t, page, `class="province-name">SORIA</span>`)
		assert.Contains(t, page, `class="province-amount">1.200 €</strong>`)
		assert.Contains(t, page, `name="radius"`)
	}
	// Replacing today's snapshot must refresh the displayed summary, too.
	data.ListaEESSPrecio[0].PrecioGasolina95E5 = "1,400"
	save()
	page := render("en")
	assert.Contains(t, page, "<td>1.400</td><td>1.500</td><td>1.600</td>")
	assert.Contains(t, page, `class="province-amount">1.400 €</strong>`)
}
