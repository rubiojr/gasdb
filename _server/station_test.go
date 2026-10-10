package main

import (
	"context"
	"encoding/json"
	"html"
	"io"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/internal/gasdb"
	"github.com/rubiojr/gasdb/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStationPage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	storage, err := gasdb.NewStorage(ctx, filepath.Join(t.TempDir(), "prices.db"), logger)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	const style = "https://maps.example/positron"
	get := func(id, lang string) (int, string) {
		t.Helper()
		response := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "/station/"+id+"?lang="+lang, nil)
		renderStation(response, request, storage, logger, id, style, translations.GetTranslations(lang))
		return response.Code, response.Body.String()
	}

	code, page := get("4375", "en")
	assert.Equal(t, 503, code, "no prices saved yet")
	assert.Contains(t, page, translations.GetEnglishTranslations().PricesUnavailable)

	data := api.GasStationList{
		Fecha: "10/10/2026 14:30:00",
		ListaEESSPrecio: []api.GasStation{
			{
				IDEESS: "4375", Rotulo: "BALLENOIL <B>", Direccion: "RONDA SEGOVIA, 37", CP: "28005",
				Localidad: "MADRID", Provincia: "MADRID", Horario: "L-D: 24H",
				Latitud: "40,410250", Longitud: "-3,717528",
				PrecioGasolina95E5: "1,729", PrecioGasoleoA: "1,859", PrecioGasesLicuados: "0,999",
			},
			{IDEESS: "99", Rotulo: "NOWHERE", Latitud: "", Longitud: ""},
		},
	}
	encoded, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, storage.SavePrices(ctx, time.Now(), encoded))

	for _, lang := range []string{"es", "en"} {
		tr := translations.GetTranslations(lang)
		code, page := get("4375", lang)
		require.Equal(t, 200, code)
		assert.Contains(t, page, `<title>GasDB — BALLENOIL &lt;B&gt;, MADRID</title>`)
		assert.Contains(t, page, `<h1>BALLENOIL &lt;B&gt;</h1>`)
		assert.Contains(t, page, "28005 MADRID", "the province matches the town, so it's left out")
		assert.NotContains(t, page, "(MADRID)")
		assert.Contains(t, page, `<dd>L-D: 24H</dd>`)
		assert.Contains(t, page, tr.PriceDataDate+" "+data.Fecha)
		assert.Contains(t, page, `<dd>1.729 €</dd>`)
		assert.Contains(t, page, "<dt>"+tr.LPG+"</dt><dd>0.999 €</dd>", "extra fuels the station sells are listed")
		assert.NotContains(t, page, "<dt>"+tr.CNG+"</dt>", "fuels it doesn't sell are not")
		assert.Contains(t, page, `<meta property="og:description" content="RONDA SEGOVIA, 37, 28005 MADRID · `+strings.TrimSuffix(tr.Gasoline95, ":")+`: 1.729 € · `)
		assert.Contains(t, page, `href="/station/4375?lang=`+lang+`"`)
		assert.Contains(t, page, `data-share data-share-title="BALLENOIL &lt;B&gt;, MADRID"`)
		assert.Contains(t, page, `href="/search?lang=`+lang+`&amp;lat=40.410250&amp;lng=-3.717528"`)
		assert.Contains(t, page, `id="toast" class="toast" role="status" data-copied="`+tr.LinkCopied+`"`)
		var stationMap templates.StationMap
		require.NoError(t, json.Unmarshal([]byte(jsonScript(t, page, "station-map-data")), &stationMap))
		assert.Equal(t, templates.StationMap{Style: style, Lng: -3.717528, Lat: 40.41025}, stationMap)
	}

	for _, id := range []string{"404", "99"} {
		code, page := get(id, "en")
		assert.Equal(t, 404, code, id)
		assert.Contains(t, page, translations.GetEnglishTranslations().StationNotFound, id)
	}
}

func TestResultCardsLinkStationPage(t *testing.T) {
	stations := []api.StationWithDistance{
		{Station: &api.GasStation{IDEESS: "4375", Rotulo: "BALLENOIL", Localidad: "MADRID", Latitud: "40,41", Longitud: "-3,71"}},
		{Station: &api.GasStation{Rotulo: "NO ID", Latitud: "40,42", Longitud: "-3,72"}},
	}
	tr := translations.GetEnglishTranslations()
	var output strings.Builder
	require.NoError(t, templates.ResultsPage(stations, "query", "", 40.4, -3.7, 5, nil, nil, templates.ResultsMap{}, tr).Render(context.Background(), &output))
	page := output.String()
	assert.Contains(t, page, `<a href="/station/4375?lang=en" class="btn btn-map" aria-label="`+html.EscapeString(tr.StationPageAlt)+`">`+tr.StationPageButton+`</a>`)
	assert.Equal(t, 1, strings.Count(page, `href="/station/`), "stations without an ID have no page")
	assert.NotContains(t, page, "data-share-title", "result cards open the station page; sharing happens there")
	assert.NotContains(t, page, `<meta name="description"`, "only station pages describe themselves")
}
