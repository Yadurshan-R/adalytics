package koios

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// retryAfter is a rate limit error that says how long to wait before retrying.
type retryAfter struct {
	wait time.Duration
	msg  string
}

// Error returns the message, so retryAfter can be used as an error.
func (r retryAfter) Error() string { return r.msg }

// parseRetryAfter reads the Retry-After header (in seconds). Without one it
// waits 60 seconds, about how long Koios blocks after too many requests.
func parseRetryAfter(h string) time.Duration {
	if s, err := strconv.Atoi(h); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	return 60 * time.Second // Koios blocks an IP for about 60s after a burst
}

// snippet shortens an error response to 200 characters for the log.
func snippet(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "…"
	}
	return string(b)
}

// sleep waits for d, or returns early with an error when ctx is cancelled.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// limiter allows at most max requests in any sliding window.
type limiter struct {
	mu     sync.Mutex
	times  []time.Time
	max    int
	window time.Duration
}

// wait blocks until another request is allowed: it forgets requests older
// than the window, and if max are still in it, sleeps until the oldest leaves.
func (l *limiter) wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := time.Now()
		kept := l.times[:0]
		for _, t := range l.times {
			if now.Sub(t) < l.window {
				kept = append(kept, t)
			}
		}
		l.times = kept
		if len(l.times) < l.max {
			l.times = append(l.times, now)
			l.mu.Unlock()
			return nil
		}
		wait := l.window - now.Sub(l.times[0])
		l.mu.Unlock()
		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}
