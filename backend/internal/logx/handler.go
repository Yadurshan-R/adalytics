// Package logx prints the backend's log as short, readable lines:
//
//	13:20:04 INFO  api     GET /api/tokens/night/chart  200  4ms  18KB  visitor=203.0.113.7  id=a1f4
//
// time, level, the part of the backend that wrote it, the message, then any
// details as key=value. It uses Go's standard log/slog, so every part of the
// backend logs the same way. Logs only go to the terminal; nothing is saved.
package logx

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Component is the key that names the part of the backend a line comes from.
const Component = "component"

// New returns a logger that prints lines at level and above to w.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(&handler{w: w, level: level, mu: &sync.Mutex{}})
}

// ParseLevel turns LOG_LEVEL ("debug", "info", "warn", "error") into a level.
func ParseLevel(s string) (slog.Level, error) {
	var l slog.Level
	err := l.UnmarshalText([]byte(strings.TrimSpace(s)))
	return l, err
}

// StdLogger gives parts of the code that use the older log.Logger (the
// Tracker) a logger that writes through slog at INFO, under a component name.
func StdLogger(l *slog.Logger, component string) *log.Logger {
	return slog.NewLogLogger(l.With(Component, component).Handler(), slog.LevelInfo)
}

// handler is a slog.Handler that writes one plain line per record.
type handler struct {
	w         io.Writer
	level     slog.Level
	component string
	attrs     []slog.Attr
	mu        *sync.Mutex // shared by every handler made from this one
}

// Enabled reports whether lines at level are printed.
func (h *handler) Enabled(_ context.Context, level slog.Level) bool { return level >= h.level }

// Handle writes one record as a line.
func (h *handler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Time.Format("15:04:05"))
	fmt.Fprintf(&b, " %-5s %-7s %s", levelName(r.Level), h.component, r.Message)
	write := func(a slog.Attr) bool {
		if a.Key == Component {
			return true
		}
		v := a.Value.Resolve().String()
		if strings.ContainsAny(v, " \t") {
			v = fmt.Sprintf("%q", v)
		}
		fmt.Fprintf(&b, "  %s=%s", a.Key, v)
		return true
	}
	for _, a := range h.attrs {
		write(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == Component {
			return true
		}
		return write(a)
	})
	b.WriteByte('\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, b.String())
	return err
}

// WithAttrs returns a handler that adds attrs to every line. A "component"
// attr becomes the line's component column instead.
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	n := *h
	n.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	for _, a := range attrs {
		if a.Key == Component {
			n.component = a.Value.String()
		}
	}
	return &n
}

// WithGroup is not used by the backend; groups are written flat.
func (h *handler) WithGroup(string) slog.Handler { return h }

// levelName is the level as shown in the log.
func levelName(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "ERROR"
	case l >= slog.LevelWarn:
		return "WARN"
	case l >= slog.LevelInfo:
		return "INFO"
	}
	return "DEBUG"
}

// Size shows a number of bytes as 512B, 18KB or 2.4MB.
func Size(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%dKB", (n+512)>>10)
	}
	return fmt.Sprintf("%dB", n)
}

// Count shows a whole number with thousands separators: 50000 → "50,000".
func Count(n int64) string {
	s := fmt.Sprint(n)
	if n < 0 {
		return "-" + Count(-n)
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// Duration shows how long something took: "<1ms", "4ms" or "1.2s".
func Duration(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return "<1ms"
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
