// Package api serves prices and charts to the frontend. It only reads from memory.
package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cardano-analytics/internal/market"
)

// Server is the API: every route, wrapped in the request log (logging.go),
// plus the counts for the minute summary (stats.go).
type Server struct {
	handler http.Handler
	stats   *stats
	log     *slog.Logger
}

// ServeHTTP answers one request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.handler.ServeHTTP(w, r) }

// New builds the API with every route. Each route only reads from the
// Tracker's memory, so it answers instantly and never calls Koios. Every
// request is logged to log.
func New(tracker *market.Tracker, log *slog.Logger) *Server {
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

	// A token page chart: /api/tokens/snek/chart?range=7D for the line, or
	// /api/tokens/snek/chart?type=candles&size=1h for candles. range is 24H,
	// 7D or 30D; size is 15m, 1h, 4h or 1D (see candleChart).
	mux.HandleFunc("GET /api/tokens/{ticker}/chart", func(w http.ResponseWriter, r *http.Request) {
		if size := r.URL.Query().Get("size"); size != "" && r.URL.Query().Get("type") == "candles" {
			candleChart(w, tracker, r.PathValue("ticker"), size)
			return
		}
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
		writeJSON(w, http.StatusOK, map[string]any{"ticker": quote.Ticker, "trades": trades, "whale_ada": market.WhaleADA(quote)})
	})

	// The Whale Trades page: one page of big trades across every token,
	// newest first, with the token list and how much history is loaded:
	// /api/whales?hours=24&min=10000&max=50000&token=snek&wallet=stake1…&offset=200.
	mux.HandleFunc("GET /api/whales", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		hours, _ := strconv.Atoi(q.Get("hours"))
		minADA, _ := strconv.ParseFloat(q.Get("min"), 64)
		maxADA, _ := strconv.ParseFloat(q.Get("max"), 64)
		offset, _ := strconv.Atoi(q.Get("offset"))
		f := market.WhaleFilter{Hours: hours, MinADA: minADA, MaxADA: maxADA, Token: q.Get("token"), Wallet: q.Get("wallet"), Offset: offset}
		trades, more := tracker.WhaleTrades(f)
		writeJSON(w, http.StatusOK, map[string]any{"trades": trades, "more": more, "tokens": tracker.WhaleTokens(),
			"ranges": tracker.History().Ranges, "share": market.WhaleShare * 100})
	})

	// What one wallet traded in the last 24 hours, token by token, from the
	// trades in memory: /api/wallets/stake1u8….
	mux.HandleFunc("GET /api/wallets/{wallet}", func(w http.ResponseWriter, r *http.Request) {
		wallet := r.PathValue("wallet")
		if !validWallet(wallet) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "not a Cardano stake or payment address"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"wallet": wallet, "tokens": tracker.Wallet(wallet)})
	})

	// A token's trading over the last 5m, 1h, 4h and 24h: price change, volume,
	// buys and sells, for the activity box: /api/tokens/snek/activity.
	mux.HandleFunc("GET /api/tokens/{ticker}/activity", func(w http.ResponseWriter, r *http.Request) {
		quote, periods, err := tracker.Periods(r.PathValue("ticker"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ticker": quote.Ticker, "periods": periods})
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

	s := &Server{stats: &stats{}, log: log}
	s.handler = withLogging(withGzip(withCORS(mux)), log, s.stats)
	return s
}

// candleChart answers a candlestick chart at one candle size, with all the
// history loaded so far (like TradingView, the visitor scrolls back through it).
func candleChart(w http.ResponseWriter, tracker *market.Tracker, ticker, sizeKey string) {
	quote, size, candles, err := tracker.Candles(ticker, sizeKey)
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
	writeJSON(w, http.StatusOK, map[string]any{
		"ticker":           quote.Ticker,
		"size":             size.Key,
		"type":             "candles",
		"interval_seconds": int(size.Interval / time.Second),
		"price_ada":        quote.PriceADA,
		"candles":          candles,
	})
}

// writeJSON sends v as JSON with the given HTTP status, and tells the browser
// not to cache it, so prices are always fresh.
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

// validWallet checks that a wallet looks like a mainnet stake address
// (stake1…) or payment address (addr1…): lowercase letters and digits only.
func validWallet(s string) bool {
	if len(s) < 50 || len(s) > 120 || !(strings.HasPrefix(s, "stake1") || strings.HasPrefix(s, "addr1")) {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
