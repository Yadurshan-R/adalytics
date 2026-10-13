// Tests for the request log, the crash guard and the minute summary.
package api

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cardano-analytics/internal/logx"
)

// serve runs one request through the logging middleware and returns the
// answer and the log it wrote.
func serve(t *testing.T, h http.HandlerFunc, r *http.Request) (*httptest.ResponseRecorder, string, *stats) {
	t.Helper()
	var out bytes.Buffer
	st := &stats{}
	w := httptest.NewRecorder()
	withLogging(h, logx.New(&out, slog.LevelDebug).With(logx.Component, "api"), st).ServeHTTP(w, r)
	return w, out.String(), st
}

// TestLogsEveryRequest checks the log line: method, path, status, size, the
// visitor behind Nginx and a request ID that is also sent to the browser.
func TestLogsEveryRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/tokens/snek/chart?range=7D", nil)
	r.RemoteAddr = "127.0.0.1:5000"
	r.Header.Set("X-Forwarded-For", "203.0.113.7, 127.0.0.1")
	w, log, st := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"hello": strings.Repeat("x", 2000)})
	}, r)
	id := w.Header().Get("X-Request-ID")
	for _, want := range []string{" INFO ", " api ", "GET  /api/tokens/snek/chart?range=7D", " 200 ", "2KB", "visitor=203.0.113.7", "id=" + id} {
		if !strings.Contains(log, want) {
			t.Errorf("log %q has no %q", log, want)
		}
	}
	if len(id) != 6 || st.requests != 1 || !st.visitors["203.0.113.7"] {
		t.Errorf("id %q, counts %+v", id, st.counts)
	}
}

// TestLogsErrorReason checks that a 404 is a warning with the reason from the answer.
func TestLogsErrorReason(t *testing.T) {
	_, log, _ := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": `unknown token "xyz"`})
	}, httptest.NewRequest("GET", "/api/tokens/xyz/chart", nil))
	if !strings.Contains(log, " WARN ") || !strings.Contains(log, " 404 ") || !strings.Contains(log, `error="unknown token \"xyz\""`) {
		t.Errorf("log: %q", log)
	}
}

// TestCrashIsCaught checks that a request that panics is answered with 500,
// logged as an error, and does not stop the server.
func TestCrashIsCaught(t *testing.T) {
	w, log, st := serve(t, func(http.ResponseWriter, *http.Request) { panic("boom") },
		httptest.NewRequest("GET", "/api/tokens", nil))
	if w.Code != http.StatusInternalServerError || !strings.Contains(log, "request crashed") || !strings.Contains(log, "error=boom") || st.errors != 1 {
		t.Errorf("code %d, errors %d, log %q", w.Code, st.errors, log)
	}
}

// TestVisitorIsOnlyTrustedLocally checks that X-Forwarded-For is ignored when
// the request does not come through Nginx on this laptop.
func TestVisitorIsOnlyTrustedLocally(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "198.51.100.9:443"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	if got := visitorOf(r); got != "198.51.100.9" {
		t.Errorf("visitor %q", got)
	}
	r.RemoteAddr = "127.0.0.1:5000"
	r.Header.Del("X-Forwarded-For")
	if got := visitorOf(r); got != "local" {
		t.Errorf("local visitor %q", got)
	}
}

// TestSummary checks the minute summary line.
func TestSummary(t *testing.T) {
	st := &stats{}
	st.add("a", 200, 40_000, 3*time.Millisecond, "GET /api/tokens")
	st.add("b", 500, 2_000, 9*time.Millisecond, "GET /api/tokens/x/chart")
	got := summary(st.take(), 4, 210_000, 2500)
	want := "last minute: 2 requests from 2 visitors, 1 error, 41KB sent, slowest 9ms (GET /api/tokens/x/chart) | koios: 4 requests, 205KB received, today 2,500 of 50,000 (5.0%)"
	if got != want {
		t.Errorf("summary\n got  %s\n want %s", got, want)
	}
	if st.take().requests != 0 {
		t.Error("take should start counting from zero")
	}
}
