// Package config loads settings from the environment and from a .env file.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KoiosBaseURL string
	KoiosToken   string
	Port         string
	PollInterval time.Duration
	HistoryDays  int // days of history to load and keep, 1 to 30
}

// Load reads .env (from the current folder, or the parent folder) and then the
// environment. Values already set in the environment win over the file.
func Load() (Config, error) {
	loadDotEnv(".env", "../.env")

	cfg := Config{
		KoiosBaseURL: strings.TrimRight(getenv("KOIOS_BASE_URL", "https://api.koios.rest/api/v1"), "/"),
		KoiosToken:   strings.TrimSpace(os.Getenv("KOIOS_API_TOKEN")),
		Port:         getenv("PORT", "8080"),
	}

	interval, err := time.ParseDuration(getenv("POLL_INTERVAL", "60s"))
	if err != nil {
		return cfg, fmt.Errorf("POLL_INTERVAL %q is not a valid duration (use e.g. 60s): %w", os.Getenv("POLL_INTERVAL"), err)
	}
	if interval < 10*time.Second {
		return cfg, errors.New("POLL_INTERVAL must be at least 10s to stay inside the Koios free tier")
	}
	cfg.PollInterval = interval

	days, err := strconv.Atoi(getenv("HISTORY_DAYS", "30"))
	if err != nil || days < 1 || days > 30 {
		return cfg, fmt.Errorf("HISTORY_DAYS %q must be a whole number from 1 to 30", os.Getenv("HISTORY_DAYS"))
	}
	cfg.HistoryDays = days

	if cfg.KoiosToken == "" {
		return cfg, errors.New("KOIOS_API_TOKEN is missing: add it to the .env file")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv reads the first file that exists. Lines look like KEY=VALUE;
// blank lines and lines starting with # are skipped.
func loadDotEnv(paths ...string) {
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.Trim(strings.TrimSpace(value), `"'`)
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, value)
			}
		}
		return
	}
}
