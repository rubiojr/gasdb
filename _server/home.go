package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/rubiojr/gasdb/_server/internal/pricesummary"
	"github.com/rubiojr/gasdb/_server/templates"
	"github.com/rubiojr/gasdb/_server/translations"
	"github.com/rubiojr/gasdb/internal/gasdb"
)

func renderHome(w http.ResponseWriter, r *http.Request, storage *gasdb.Storage, logger *slog.Logger, t translations.Translations) {
	lastUpdate, err := storage.GetLastUpdateDate(r.Context())
	if err != nil {
		logger.Error("Error getting last update date", "error", err)
	}
	// GetLastPrices caches the decoded snapshot and invalidates it on updates.
	prices, err := storage.GetLastPrices(r.Context())
	if err != nil {
		logger.Error("Error getting daily price summary", "error", err)
	}
	summary := pricesummary.Build(prices, time.Now())
	if err := templates.Home(lastUpdate, summary, t).Render(r.Context(), w); err != nil {
		logger.Error("Error rendering home page", "error", err)
	}
}
