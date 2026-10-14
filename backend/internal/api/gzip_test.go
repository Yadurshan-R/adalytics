// Tests for gzip compression of API answers.
package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// get asks h for a page, with or without accepting gzip.
func get(h http.Handler, gz bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "/api/tokens", nil)
	if gz {
		r.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// TestGzipJSON checks that JSON is compressed for browsers that accept gzip,
// comes back unchanged when unpacked, and is sent plain to those that do not.
func TestGzipJSON(t *testing.T) {
	body := `{"tokens":[` + strings.Repeat(`{"ticker":"SNEK","price_ada":0.0021},`, 500) + `{}]}`
	h := withGzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, rawJSON(body))
	}))

	w := get(h, true)
	if w.Header().Get("Content-Encoding") != "gzip" || w.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("headers: %v", w.Header())
	}
	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	plain, _ := io.ReadAll(zr)
	if !strings.Contains(string(plain), `"SNEK"`) || len(plain) < 10*w.Body.Len() {
		t.Errorf("unpacked %d bytes from %d, want at least 10 times smaller", len(plain), w.Body.Len())
	}

	w = get(h, false)
	if w.Header().Get("Content-Encoding") != "" || !strings.Contains(w.Body.String(), `"SNEK"`) {
		t.Errorf("without gzip: %v %q", w.Header(), w.Body.String()[:40])
	}
}

// TestGzipSkipsLogosAndErrors checks that PNG logos and error answers are sent
// as they are, so the log can still read the error.
func TestGzipSkipsLogosAndErrors(t *testing.T) {
	logo := withGzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("\x89PNG"))
	}))
	if w := get(logo, true); w.Header().Get("Content-Encoding") != "" || w.Body.String() != "\x89PNG" {
		t.Errorf("logo was compressed: %v", w.Header())
	}
	missing := withGzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown token"})
	}))
	if w := get(missing, true); w.Header().Get("Content-Encoding") != "" || !strings.Contains(w.Body.String(), "unknown token") {
		t.Errorf("error answer was compressed: %v", w.Header())
	}
}

// rawJSON lets a test send a ready-made JSON text through writeJSON unchanged.
type rawJSON string

func (j rawJSON) MarshalJSON() ([]byte, error) { return []byte(j), nil }
