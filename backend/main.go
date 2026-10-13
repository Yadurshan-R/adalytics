// Cardano token analytics backend: reads token prices from Koios (read-only)
// and serves them from memory.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cardano-analytics/internal/api"
	"cardano-analytics/internal/config"
	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/logx"
	"cardano-analytics/internal/market"
)

// SlowKoios is how long a Koios answer may take before it is logged as a warning.
const SlowKoios = 20 * time.Second

// main loads the settings, starts the Tracker that reads Cardano data from
// Koios in the background, and serves the API until Ctrl+C is pressed. Every
// API call, every Koios answer (with LOG_LEVEL=debug) and a summary each
// minute are written to the log.
func main() {
	cfg, err := config.Load()
	logger := logx.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		logger.Error("config: " + err.Error())
		os.Exit(1)
	}
	mainLog := logger.With(logx.Component, "main")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mainLog.Info(fmt.Sprintf("Tracking %d verified Cardano tokens, showing the top %d, from Koios (%s) every %s, with %d days of history",
		len(market.Tokens), market.TopN, cfg.KoiosBaseURL, cfg.PollInterval, cfg.HistoryDays), "log_level", cfg.LogLevel.String())

	client := koios.New(cfg.KoiosBaseURL, cfg.KoiosToken)
	tracker := market.NewTracker(client, market.Tokens, cfg.PollInterval, logx.StdLogger(logger, "market"))
	tracker.SetHistoryDays(cfg.HistoryDays)
	client.OnRetry = func(endpoint string, reason error, wait time.Duration, attempt int) {
		// Note also shows it on the loading page, so a slow start is explained.
		tracker.Note("Koios %s: %v, retrying in %s (attempt %d of 3)", endpoint, reason, wait, attempt)
	}
	client.OnResponse = logKoios(logger.With(logx.Component, "koios"))
	go tracker.Run(ctx)

	apiServer := api.New(tracker, logger.With(logx.Component, "api"))
	go apiServer.ReportEveryMinute(ctx, func() (int64, int64) { return client.Requests(), client.BytesReceived() })

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apiServer,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		mainLog.Info("API ready at http://localhost:" + cfg.Port + "/api/tokens")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			mainLog.Error("server: " + err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	mainLog.Info("Shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

// logKoios logs each Koios answer: the endpoint, status, time and how much
// was sent and received. Normal answers show with LOG_LEVEL=debug; failures
// and very slow answers always show.
func logKoios(log *slog.Logger) func(koios.Call) {
	return func(c koios.Call) {
		level := slog.LevelDebug
		if c.Err != nil || c.Status >= 400 {
			level = slog.LevelWarn
		} else if c.Took >= SlowKoios {
			level = slog.LevelWarn
		}
		status := fmt.Sprint(c.Status)
		if c.Status == 0 {
			status = "---"
		}
		msg := fmt.Sprintf("%-16s %s %7s  sent %-6s received %s", c.Endpoint, status, logx.Duration(c.Took), logx.Size(int64(c.Sent)), logx.Size(int64(c.Received)))
		args := []any{}
		if c.Attempt > 0 {
			args = append(args, "retry", c.Attempt)
		}
		if c.Err != nil {
			args = append(args, "error", c.Err.Error())
		}
		log.Log(context.Background(), level, msg, args...)
	}
}
