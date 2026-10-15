// Package koios is a small read-only client for the Koios REST API
// (https://koios.rest), which serves Cardano blockchain data.
//
// The package is split over five files:
//
//	client.go     the Client, and do(), which sends every request
//	types.go      the shapes of the data Koios returns
//	endpoints.go  one function per Koios endpoint we use
//	retry.go      retry timing and the rate limiter
//	parallel.go   runs batches of requests a few at a time
package koios

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// MaxBodyBytes is the largest request body Koios's free tier accepts.
const MaxBodyBytes = 5120

// ErrUnauthorized means Koios rejected the API token.
var ErrUnauthorized = errors.New("Koios rejected the API token: check KOIOS_API_TOKEN in .env, or renew it at https://koios.rest/Profile.html")

// Client talks to Koios. It adds the API token to every request, keeps under
// the free tier's rate limit and counts the requests it sends.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
	slow    *http.Client // for queries Koios takes minutes to answer
	limit   *limiter
	count   atomic.Int64
	bytes   atomic.Int64 // response bytes received since start

	// OnRetry, if set, is told whenever a request is retried.
	OnRetry func(endpoint string, reason error, wait time.Duration, attempt int)
	// OnResponse, if set, is told about every answer from Koios (and every
	// request that got none), so the backend can log what it downloads.
	OnResponse func(Call)
}

// Call describes one request to Koios and its answer, for the log.
type Call struct {
	Endpoint string        // e.g. /tx_info
	Status   int           // HTTP status; 0 if Koios did not answer
	Sent     int           // request body size in bytes
	Received int           // response size in bytes
	Took     time.Duration // from sending to the full answer
	Attempt  int           // 0 for the first try, 1 to 3 for retries
	Err      error         // why there was no usable answer, if so
}

// New creates a Client for the Koios API at baseURL, using the API token.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		// Koios sometimes needs a little over a minute for the big Minswap
		// queries; waiting 2 minutes keeps that answer instead of retrying.
		http: &http.Client{Timeout: 2 * time.Minute},
		slow: &http.Client{Timeout: 3 * time.Minute},
		// Free tier allows 100 requests per 10 seconds; keep a margin.
		limit: &limiter{max: 90, window: 10 * time.Second},
	}
}

// Requests is the number of requests sent to Koios since start, retries included.
func (c *Client) Requests() int64 { return c.count.Load() }

// BytesReceived is how much data Koios has sent back since start.
func (c *Client) BytesReceived() int64 { return c.bytes.Load() }

// report tells OnResponse about one request, if it is set.
func (c *Client) report(call Call) {
	if c.OnResponse != nil {
		c.OnResponse(call)
	}
}

// do sends a request and decodes the JSON answer into out. It waits for the
// rate limiter and retries when Koios is busy (429), slow or briefly down (5xx).
func (c *Client) do(ctx context.Context, hc *http.Client, method, path string, body any, out any) error {
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return err
		}
		if len(payload) >= MaxBodyBytes {
			return fmt.Errorf("request body is %d bytes, Koios accepts less than %d: send fewer items per request", len(payload), MaxBodyBytes)
		}
	}
	endpoint := path
	if i := bytes.IndexByte([]byte(path), '?'); i >= 0 {
		endpoint = path[:i]
	}

	backoff := []time.Duration{2 * time.Second, 5 * time.Second, 15 * time.Second}
	var lastErr error
	for attempt := 0; attempt <= len(backoff); attempt++ {
		if attempt > 0 {
			wait := backoff[attempt-1]
			if ra, ok := lastErr.(retryAfter); ok && ra.wait > wait {
				wait = ra.wait
			}
			if c.OnRetry != nil {
				c.OnRetry(endpoint, lastErr, wait, attempt)
			}
			if err := sleep(ctx, wait); err != nil {
				return err
			}
		}
		if err := c.limit.wait(ctx); err != nil {
			return err
		}

		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/json")
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		c.count.Add(1)
		sent := time.Now()
		resp, err := hc.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				lastErr = fmt.Errorf("Koios did not answer within %s", hc.Timeout)
			} else {
				lastErr = errors.New("could not reach Koios (check your internet connection)")
			}
			c.report(Call{Endpoint: endpoint, Sent: len(payload), Took: time.Since(sent), Attempt: attempt, Err: lastErr})
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
		resp.Body.Close()
		c.bytes.Add(int64(len(data)))
		c.report(Call{Endpoint: endpoint, Status: resp.StatusCode, Sent: len(payload), Received: len(data), Took: time.Since(sent), Attempt: attempt, Err: readErr})

		switch {
		case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
			return ErrUnauthorized
		case resp.StatusCode == http.StatusTooManyRequests:
			lastErr = retryAfter{wait: parseRetryAfter(resp.Header.Get("Retry-After")), msg: "rate limit reached (429)"}
			continue
		case resp.StatusCode == http.StatusGatewayTimeout:
			lastErr = errors.New("query took longer than 30s (504)")
			continue
		case resp.StatusCode >= 500:
			lastErr = fmt.Errorf("Koios returned %d: %s", resp.StatusCode, snippet(data))
			continue
		case resp.StatusCode >= 400:
			return fmt.Errorf("Koios returned %d: %s", resp.StatusCode, snippet(data))
		}
		if readErr != nil {
			lastErr = fmt.Errorf("reading Koios response: %w", readErr)
			continue
		}
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("unexpected Koios response: %w", err)
		}
		return nil
	}
	return lastErr
}
