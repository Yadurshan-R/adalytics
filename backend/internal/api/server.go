// Package api serves prices and charts to the frontend. It only reads from memory.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cardano-analytics/internal/market"
)

func New(tracker *market.Tracker) http.Handler {
	mux := http.NewServeMux()

	// The list page: every token with price, changes, volume and a 7-day sparkline.
	mux.HandleFunc("GET /api/tokens", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"currency": "ADA",
			"tokens":   tracker.Quotes(),
			"now":      time.Now(),
		})
	})

	// A token page chart: /api/tokens/SNEK/chart?range=24H (or 7D).
	mux.HandleFunc("GET /api/tokens/{ticker}/chart", func(w http.ResponseWriter, r *http.Request) {
		rangeKey := r.URL.Query().Get("range")
		if rangeKey == "" {
			rangeKey = "24H"
		}
		quote, points, err := tracker.Chart(r.PathValue("ticker"), rangeKey)
		if err != nil {
			status := http.StatusBadRequest
			if strings.HasPrefix(err.Error(), "unknown token") {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		rng := market.Ranges[strings.ToUpper(rangeKey)]
		change := 0.0
		if len(points) > 0 && points[0].Price > 0 {
			change = (points[len(points)-1].Price/points[0].Price - 1) * 100
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ticker":           quote.Ticker,
			"range":            rng.Key,
			"interval_seconds": int(rng.Interval / time.Second),
			"change_pct":       change,
			"price_ada":        quote.PriceADA,
			"points":           points,
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
