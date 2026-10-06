// Cardano token analytics backend: reads token prices from Koios (read-only)
// and serves them from memory.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cardano-analytics/internal/api"
	"cardano-analytics/internal/config"
	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/market"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ltime)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Printf("Tracking %d verified Cardano tokens, showing the top %d, from Koios (%s) every %s",
		len(market.Tokens), market.TopN, cfg.KoiosBaseURL, cfg.PollInterval)

	client := koios.New(cfg.KoiosBaseURL, cfg.KoiosToken)
	client.OnRetry = func(endpoint string, reason error, wait time.Duration, attempt int) {
		logger.Printf("Koios %s: %v, retrying in %s (attempt %d of 3)", endpoint, reason, wait, attempt)
	}
	tracker := market.NewTracker(client, market.Tokens, cfg.PollInterval, logger)
	go tracker.Run(ctx)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.New(tracker),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Printf("API ready at http://localhost:%s/api/tokens", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Println("Shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
