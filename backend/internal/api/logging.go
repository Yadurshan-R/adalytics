package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"cardano-analytics/internal/logx"
)

// SlowRequest is how long an API answer may take before it is logged as a warning.
const SlowRequest = time.Second

// recorder remembers the status, size and (for errors) the start of the body
// of an answer, so the log can show them.
type recorder struct {
	http.ResponseWriter
	status int
	size   int64
	wrote  bool
	body   bytes.Buffer // the first part of an error answer
}

// WriteHeader keeps the status before passing it on.
func (r *recorder) WriteHeader(status int) {
	if !r.wrote {
		r.status, r.wrote = status, true
		r.ResponseWriter.WriteHeader(status)
	}
}

// Write counts the bytes sent and keeps the start of an error answer.
func (r *recorder) Write(b []byte) (int, error) {
	if !r.wrote {
		r.WriteHeader(http.StatusOK)
	}
	if r.status >= 400 && r.body.Len() < 512 {
		r.body.Write(b[:min(len(b), 512-r.body.Len())])
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += int64(n)
	return n, err
}

// withLogging logs one line per API call: what was asked, the status, how
// long it took, how much was sent, who asked and a request ID. The ID is also
// sent back in the X-Request-ID header, so a problem a visitor sees can be
// found in the log. A request that panics is logged with its stack, answered
// with 500, and the server keeps running.
func withLogging(next http.Handler, log *slog.Logger, st *stats) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := newID()
		w.Header().Set("X-Request-ID", id)
		rec := &recorder{ResponseWriter: w}
		defer func() {
			if p := recover(); p != nil {
				log.Error("request crashed", "path", r.URL.Path, "error", fmt.Sprint(p), "id", id, "stack", firstLines(debug.Stack(), 8))
				if !rec.wrote {
					writeJSON(rec, http.StatusInternalServerError, map[string]string{"error": "internal error, request id " + id})
				}
				rec.status = http.StatusInternalServerError
			}
			if !rec.wrote {
				rec.status = http.StatusOK
			}
			took := time.Since(start)
			visitor := visitorOf(r)
			st.add(visitor, rec.status, rec.size, took, r.Method+" "+r.URL.Path)

			msg := fmt.Sprintf("%-4s %-34s %3d %6s %6s", r.Method, r.URL.Path, rec.status, logx.Duration(took), logx.Size(rec.size))
			args := []any{"visitor", visitor, "id", id}
			if reason := errorOf(rec); reason != "" {
				args = append(args, "error", reason)
			}
			log.Log(r.Context(), levelOf(r, rec.status, took), msg, args...)
		}()
		next.ServeHTTP(rec, r)
	})
}

// levelOf decides how loud a request's log line is: errors and slow answers
// stand out, and logos and health checks (many, and boring) only show with
// LOG_LEVEL=debug.
func levelOf(r *http.Request, status int, took time.Duration) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400 || took >= SlowRequest:
		return slog.LevelWarn
	case r.Method == http.MethodOptions || strings.HasSuffix(r.URL.Path, "/logo") || r.URL.Path == "/api/health":
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// visitorOf is the visitor's IP address. Behind Nginx and ngrok the request
// comes from this laptop, and the real address is the first one in
// X-Forwarded-For; that header is only trusted when the request is local.
func visitorOf(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			first, _, _ := strings.Cut(fwd, ",")
			return strings.TrimSpace(first)
		}
		return "local"
	}
	return host
}

// errorOf is the "error" message of an error answer, for the log.
func errorOf(rec *recorder) string {
	if rec.status < 400 || rec.body.Len() == 0 {
		return ""
	}
	var body struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(rec.body.Bytes(), &body) == nil && body.Error != "" {
		return body.Error
	}
	return strings.TrimSpace(rec.body.String())
}

// newID is a short random request ID, e.g. "a1f40c".
func newID() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// firstLines keeps the first n lines of a stack trace, joined with " | ".
func firstLines(stack []byte, n int) string {
	lines := strings.Split(strings.TrimSpace(string(stack)), "\n")
	return strings.Join(lines[:min(n, len(lines))], " | ")
}
