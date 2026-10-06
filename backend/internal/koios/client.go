// Package koios is a small read-only client for the Koios REST API.
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
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Asset identifies a native token. An empty PolicyID means ADA.
type Asset struct {
	PolicyID  string `json:"policy_id"`
	AssetName string `json:"asset_name"`
}

func (a Asset) IsADA() bool { return a.PolicyID == "" && a.AssetName == "" }

type AssetAmount struct {
	PolicyID  string `json:"policy_id"`
	AssetName string `json:"asset_name"`
	Decimals  int    `json:"decimals"`
	Quantity  string `json:"quantity"`
}

// PlutusData is the decoded datum Koios returns in inline_datum.value.
type PlutusData struct {
	Constructor *int         `json:"constructor,omitempty"`
	Fields      []PlutusData `json:"fields,omitempty"`
	Int         *json.Number `json:"int,omitempty"`
	Bytes       *string      `json:"bytes,omitempty"`
}

type InlineDatum struct {
	Bytes string     `json:"bytes"`
	Value PlutusData `json:"value"`
}

type UTxO struct {
	TxHash      string        `json:"tx_hash"`
	TxIndex     int           `json:"tx_index"`
	Address     string        `json:"address"`
	Value       string        `json:"value"`
	BlockHeight int           `json:"block_height"`
	BlockTime   int64         `json:"block_time"`
	AssetList   []AssetAmount `json:"asset_list"`
	InlineDatum *InlineDatum  `json:"inline_datum"`
	IsSpent     bool          `json:"is_spent"`
}

// Ref is the UTxO reference "tx_hash#index".
func (u UTxO) Ref() string { return fmt.Sprintf("%s#%d", u.TxHash, u.TxIndex) }

// AddressTx is one transaction that touched an address.
type AddressTx struct {
	TxHash      string `json:"tx_hash"`
	BlockHeight int    `json:"block_height"`
	BlockTime   int64  `json:"block_time"`
}

// Tx is a transaction from tx_info, with only the fields we use.
type Tx struct {
	TxHash       string     `json:"tx_hash"`
	BlockHeight  int        `json:"block_height"`
	TxTimestamp  int64      `json:"tx_timestamp"`
	TxBlockIndex int        `json:"tx_block_index"`
	Outputs      []TxOutput `json:"outputs"`
}

type TxOutput struct {
	PaymentAddr struct {
		Bech32 string `json:"bech32"`
		Cred   string `json:"cred"`
	} `json:"payment_addr"`
	TxHash      string        `json:"tx_hash"`
	TxIndex     int           `json:"tx_index"`
	Value       string        `json:"value"`
	AssetList   []AssetAmount `json:"asset_list"`
	InlineDatum *InlineDatum  `json:"inline_datum"`
}

// UTxO converts a transaction output into the same shape as a UTxO lookup.
func (o TxOutput) UTxO(tx Tx) UTxO {
	return UTxO{
		TxHash: o.TxHash, TxIndex: o.TxIndex, Address: o.PaymentAddr.Bech32, Value: o.Value,
		BlockHeight: tx.BlockHeight, BlockTime: tx.TxTimestamp,
		AssetList: o.AssetList, InlineDatum: o.InlineDatum,
	}
}

// MaxBodyBytes is the largest request body Koios's free tier accepts.
const MaxBodyBytes = 5120

// ErrUnauthorized means Koios rejected the API token.
var ErrUnauthorized = errors.New("Koios rejected the API token: check KOIOS_API_TOKEN in .env, or renew it at https://koios.rest/Profile.html")

type Client struct {
	baseURL string
	token   string
	http    *http.Client
	slow    *http.Client // for queries Koios takes minutes to answer
	limit   *limiter
	count   atomic.Int64

	// OnRetry, if set, is told whenever a request is retried.
	OnRetry func(endpoint string, reason error, wait time.Duration, attempt int)
}

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

// Tip is the newest block Koios has indexed.
type Tip struct {
	BlockHeight int   `json:"block_height"`
	BlockTime   int64 `json:"block_time"`
}

func (c *Client) Tip(ctx context.Context) (Tip, error) {
	var out []Tip
	if err := c.do(ctx, c.http, http.MethodGet, "/tip?select=block_height,block_time", nil, &out); err != nil {
		return Tip{}, err
	}
	if len(out) == 0 {
		return Tip{}, errors.New("Koios returned no tip")
	}
	return out[0], nil
}

// UTxOInfo looks up UTxOs by reference ("tx_hash#index") and says if each is spent.
func (c *Client) UTxOInfo(ctx context.Context, refs []string) ([]UTxO, error) {
	body := map[string]any{"_utxo_refs": refs, "_extended": false}
	var out []UTxO
	err := c.do(ctx, c.http, http.MethodPost, "/utxo_info?select=tx_hash,tx_index,is_spent", body, &out)
	return out, err
}

// PoolUTxOs returns the unspent outputs at address that hold any of the given
// assets. Each Minswap pool holds its own LP token, so asking for the LP tokens
// at the pool address returns exactly those pools, traded recently or not.
// Requests are split so each body stays under Koios's size limit.
func (c *Client) PoolUTxOs(ctx context.Context, address string, assets []Asset, progress func(done, total int)) ([]UTxO, error) {
	const perRequest = 30 // 30 LP tokens make a body of about 3.9 KB
	q := url.Values{}
	q.Set("select", "tx_hash,tx_index,address,value,block_height,block_time,asset_list,inline_datum")
	q.Set("address", "eq."+address)
	var all []UTxO
	for start := 0; start < len(assets); start += perRequest {
		end := min(start+perRequest, len(assets))
		list := make([][2]string, 0, end-start)
		for _, a := range assets[start:end] {
			list = append(list, [2]string{a.PolicyID, a.AssetName})
		}
		body := map[string]any{"_asset_list": list, "_extended": true}
		var rows []UTxO
		if err := c.do(ctx, c.http, http.MethodPost, "/asset_utxos?"+q.Encode(), body, &rows); err != nil {
			return nil, err
		}
		all = append(all, rows...)
		if progress != nil {
			progress(end, len(assets))
		}
	}
	return all, nil
}

// AddressTxs lists every transaction that touched the addresses from a block
// height on (newest first). Koios returns at most 1000 rows per page, so it
// pages through; progress, if set, is called after each page.
func (c *Client) AddressTxs(ctx context.Context, addresses []string, afterBlockHeight int, progress func(rows int)) ([]AddressTx, error) {
	const page = 1000
	body := map[string]any{"_addresses": addresses, "_after_block_height": afterBlockHeight}
	var all []AddressTx
	seen := map[string]bool{}
	for offset := 0; ; offset += page {
		q := url.Values{}
		q.Set("order", "block_height.desc,tx_hash.asc")
		q.Set("limit", strconv.Itoa(page))
		q.Set("offset", strconv.Itoa(offset))
		var rows []AddressTx
		if err := c.do(ctx, c.http, http.MethodPost, "/address_txs?"+q.Encode(), body, &rows); err != nil {
			return nil, err
		}
		for _, r := range rows {
			if !seen[r.TxHash] {
				seen[r.TxHash] = true
				all = append(all, r)
			}
		}
		if progress != nil {
			progress(len(all))
		}
		if len(rows) < page {
			return all, nil
		}
	}
}

// AssetInfo is a token's supply, identity and Cardano Token Registry entry.
type AssetInfo struct {
	PolicyID      string `json:"policy_id"`
	AssetName     string `json:"asset_name"`
	Fingerprint   string `json:"fingerprint"`   // CIP-14 id, "asset1…"
	CreationTime  int64  `json:"creation_time"` // first mint, unix seconds
	TotalSupply   string `json:"total_supply"`  // raw units, before decimals
	TokenRegistry *struct {
		Ticker      string `json:"ticker"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Logo        string `json:"logo"` // base64 PNG
		Decimals    int    `json:"decimals"`
	} `json:"token_registry_metadata"`
}

// AssetInfos returns supply and registry details for several tokens in one request.
func (c *Client) AssetInfos(ctx context.Context, assets []Asset) ([]AssetInfo, error) {
	list := make([][2]string, len(assets))
	for i, a := range assets {
		list[i] = [2]string{a.PolicyID, a.AssetName}
	}
	body := map[string]any{"_asset_list": list}
	var out []AssetInfo
	err := c.do(ctx, c.http, http.MethodPost, "/asset_info?select=policy_id,asset_name,fingerprint,creation_time,total_supply,token_registry_metadata", body, &out)
	return out, err
}

// AssetSummary counts a token's holders. Koios takes a minute or more for big tokens.
type AssetSummary struct {
	StakedWallets     int `json:"staked_wallets"`
	UnstakedAddresses int `json:"unstaked_addresses"`
}

func (c *Client) AssetSummary(ctx context.Context, asset Asset) (AssetSummary, error) {
	q := url.Values{}
	q.Set("_asset_policy", asset.PolicyID)
	q.Set("_asset_name", asset.AssetName)
	var out []AssetSummary
	if err := c.do(ctx, c.slow, http.MethodGet, "/asset_summary?"+q.Encode(), nil, &out); err != nil {
		return AssetSummary{}, err
	}
	if len(out) == 0 {
		return AssetSummary{}, errors.New("Koios returned no summary")
	}
	return out[0], nil
}

// TxInfo returns transactions with their outputs, assets and inline datums.
func (c *Client) TxInfo(ctx context.Context, hashes []string) ([]Tx, error) {
	body := map[string]any{
		"_tx_hashes": hashes,
		"_inputs":    false, "_metadata": false, "_assets": true, "_withdrawals": false,
		"_certs": false, "_scripts": true, "_bytecode": false, "_governance": false,
	}
	var out []Tx
	err := c.do(ctx, c.http, http.MethodPost, "/tx_info?select=tx_hash,block_height,tx_timestamp,tx_block_index,outputs", body, &out)
	return out, err
}

// Requests is the number of requests sent to Koios since start, retries included.
func (c *Client) Requests() int64 { return c.count.Load() }

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
			continue
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
		resp.Body.Close()

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

type retryAfter struct {
	wait time.Duration
	msg  string
}

func (r retryAfter) Error() string { return r.msg }

func parseRetryAfter(h string) time.Duration {
	if s, err := strconv.Atoi(h); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	return 60 * time.Second // Koios blocks an IP for about 60s after a burst
}

func snippet(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "…"
	}
	return string(b)
}

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
