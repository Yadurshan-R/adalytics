package api

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

// gzipPool keeps gzip writers for reuse, so each answer does not allocate a
// new one.
var gzipPool = sync.Pool{New: func() any { w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed); return w }}

// gzipWriter compresses an answer when it is a successful JSON answer and the
// browser accepts gzip. Logos (already compressed PNG) and error answers are
// sent as they are.
type gzipWriter struct {
	http.ResponseWriter
	gz      *gzip.Writer
	decided bool
}

// WriteHeader decides, from the status and content type, whether to compress.
func (w *gzipWriter) WriteHeader(status int) {
	if !w.decided {
		w.decided = true
		h := w.Header()
		if status == http.StatusOK && strings.HasPrefix(h.Get("Content-Type"), "application/json") {
			h.Del("Content-Length")
			h.Set("Content-Encoding", "gzip")
			w.gz = gzipPool.Get().(*gzip.Writer)
			w.gz.Reset(w.ResponseWriter)
		}
	}
	w.ResponseWriter.WriteHeader(status)
}

// Write sends the bytes, compressed when compression was chosen.
func (w *gzipWriter) Write(b []byte) (int, error) {
	if !w.decided {
		w.WriteHeader(http.StatusOK)
	}
	if w.gz != nil {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// close finishes the compressed answer and returns the writer to the pool.
func (w *gzipWriter) close() {
	if w.gz != nil {
		w.gz.Close()
		gzipPool.Put(w.gz)
		w.gz = nil
	}
}

// withGzip compresses JSON answers for browsers that accept gzip, so the
// token list and charts reach visitors several times smaller and faster.
func withGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gw := &gzipWriter{ResponseWriter: w}
		defer gw.close()
		next.ServeHTTP(gw, r)
	})
}
