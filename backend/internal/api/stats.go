package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cardano-analytics/internal/logx"
)

// KoiosDailyLimit is the free Koios tier's requests per day, for the summary.
const KoiosDailyLimit = 50_000

// counts is the API traffic between two minute summaries.
type counts struct {
	requests int
	errors   int // answers with status 500 or more
	sent     int64
	visitors map[string]bool
	slowest  time.Duration
	slowPath string
}

// stats keeps the counts safe while requests are answered at the same time.
type stats struct {
	mu sync.Mutex
	counts
}

// add counts one answered request.
func (s *stats) add(visitor string, status int, size int64, took time.Duration, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.visitors == nil {
		s.visitors = map[string]bool{}
	}
	s.requests++
	if status >= 500 {
		s.errors++
	}
	s.sent += size
	s.visitors[visitor] = true
	if took > s.slowest {
		s.slowest, s.slowPath = took, path
	}
}

// take returns the counts so far and starts counting again from zero.
func (s *stats) take() counts {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.counts
	s.counts = counts{}
	return out
}

// KoiosUsage returns the Koios requests sent and bytes received since start.
type KoiosUsage func() (requests, bytes int64)

// ReportEveryMinute logs one summary line a minute: the API traffic (requests,
// visitors, errors, data sent, the slowest answer) and the Koios traffic
// (requests and data received this minute, and requests today against the
// free tier's daily limit). "Today" is the UTC day, counted by this backend
// since it started.
func (s *Server) ReportEveryMinute(ctx context.Context, koios KoiosUsage) {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	lastReq, lastBytes := koios()
	day, dayStartReq := time.Now().UTC().YearDay(), lastReq
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-tick.C:
			req, bytes := koios()
			if d := now.UTC().YearDay(); d != day {
				day, dayStartReq = d, lastReq
			}
			api := s.stats.take()
			s.log.Info(summary(api, req-lastReq, bytes-lastBytes, req-dayStartReq))
			lastReq, lastBytes = req, bytes
		}
	}
}

// summary is the text of one minute summary line.
func summary(api counts, koiosReq, koiosBytes, koiosToday int64) string {
	line := fmt.Sprintf("last minute: %d request%s from %d visitor%s, %d error%s, %s sent",
		api.requests, plural(api.requests), len(api.visitors), plural(len(api.visitors)), api.errors, plural(api.errors), logx.Size(api.sent))
	if api.requests > 0 {
		line += fmt.Sprintf(", slowest %s (%s)", logx.Duration(api.slowest), api.slowPath)
	}
	return line + fmt.Sprintf(" | koios: %s request%s, %s received, today %s of %s (%.1f%%)",
		logx.Count(koiosReq), plural(int(koiosReq)), logx.Size(koiosBytes), logx.Count(koiosToday), logx.Count(KoiosDailyLimit), float64(koiosToday)*100/KoiosDailyLimit)
}

// plural returns "s" unless n is 1.
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
