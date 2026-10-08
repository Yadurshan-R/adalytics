// Package api serves prices and charts to the frontend. It only reads from memory.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cardano-analytics/internal/market"
)

func New(tracker *market.Tracker) http.Handler {
	mux := http.NewServeMux()

	// The list page: every token with price, changes, volume and a small price
	// line, plus how much history is loaded (for the 7D/30D buttons).
	mux.HandleFunc("GET /api/tokens", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"currency": "ADA",
			"tokens":   tracker.Quotes(),
			"history":  tracker.History(),
			"now":      time.Now(),
		})
	})

	// A token page chart: /api/tokens/snek/chart?range=7D&type=candles.
	// range is 24H, 7D or 30D; type is line (the default) or candles.
	mux.HandleFunc("GET /api/tokens/{ticker}/chart", func(w http.ResponseWriter, r *http.Request) {
		rangeKey := r.URL.Query().Get("range")
		if rangeKey == "" {
			rangeKey = "24H"
		}
		kind := r.URL.Query().Get("type")
		if kind == "" {
			kind = "line"
		}
		if kind != "line" && kind != "candles" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type must be line or candles"})
			return
		}
		quote, rng, points, candles, err := tracker.Chart(r.PathValue("ticker"), rangeKey, kind == "candles")
		switch {
		case errors.Is(err, market.ErrRangeLoading):
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "history": tracker.History()})
			return
		case err != nil && strings.HasPrefix(err.Error(), "unknown token"):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		case err != nil:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		interval := rng.Interval
		if kind == "candles" {
			interval = rng.CandleInterval
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ticker":           quote.Ticker,
			"range":            rng.Key,
			"type":             kind,
			"interval_seconds": int(interval / time.Second),
			"price_ada":        quote.PriceADA,
			"points":           points,
			"candles":          candles,
		})
	})

	// A token page's recent trades, newest first: /api/tokens/snek/trades?limit=5.
	mux.HandleFunc("GET /api/tokens/{ticker}/trades", func(w http.ResponseWriter, r *http.Request) {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		quote, trades, err := tracker.Trades(r.PathValue("ticker"), limit)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ticker": quote.Ticker, "trades": trades})
	})

	// A token's logo from the Cardano token registry (PNG). Browsers cache it for a day.
	mux.HandleFunc("GET /api/tokens/{ticker}/logo", func(w http.ResponseWriter, r *http.Request) {
		png, ok := tracker.Logo(r.PathValue("ticker"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(png)
	})

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return withCORS(mux)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// withCORS lets the Vue dev server (another port) read the API. Read-only: GET only.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
