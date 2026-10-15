package market

import (
	"context"
	"encoding/base64"
	"errors"
	"math/big"
	"net/url"
	"strings"
	"time"

	"cardano-analytics/internal/koios"
)

// loadInfo reads total supply, ticker and logo for every tracked token with
// asset_info, and each token's minting policy with script_info (policy.go),
// 25 tokens per request. If it fails, prices still work and it is tried again
// next minute.
func (t *Tracker) loadInfo(ctx context.Context) {
	t.mu.RLock()
	act := t.tracked()
	t.mu.RUnlock()
	if len(act) == 0 {
		return
	}
	var infos []koios.AssetInfo
	for start := 0; start < len(act); start += 25 {
		end := min(start+25, len(act))
		assets := make([]koios.Asset, 0, end-start)
		for _, s := range act[start:end] {
			assets = append(assets, s.token.Asset)
		}
		batch, err := t.client.AssetInfos(ctx, assets)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.log.Printf("Could not load token supply and logos: %v (trying again in %s)", err, t.interval)
			}
			return
		}
		infos = append(infos, batch...)
	}
	seen := map[string]bool{}
	var policies []string
	for _, s := range act {
		if p := s.token.Asset.PolicyID; !seen[p] {
			seen[p] = true
			policies = append(policies, p)
		}
	}
	mints, err := t.readPolicies(ctx, policies)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.log.Printf("Could not read minting policies: %v (trying again in %s)", err, t.interval)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	logos := 0
	for _, s := range act {
		if m := mints[s.token.Asset.PolicyID]; m != nil {
			s.mint = m
		}
	}
	for _, info := range infos {
		for _, s := range act {
			if s.token.Asset.PolicyID != info.PolicyID || s.token.Asset.AssetName != info.AssetName {
				continue
			}
			if n, ok := new(big.Int).SetString(info.TotalSupply, 10); ok {
				s.supply = n
			}
			s.info.Fingerprint = info.Fingerprint
			if info.CreationTime > 0 {
				created := time.Unix(info.CreationTime, 0).UTC()
				s.info.Created = &created
			}
			if reg := info.TokenRegistry; reg != nil {
				if reg.Ticker != "" {
					s.quote.Ticker = reg.Ticker
				}
				s.info.Description = strings.TrimSpace(reg.Description)
				s.info.Website = safeURL(reg.URL)
				if png, err := base64.StdEncoding.DecodeString(reg.Logo); err == nil && len(png) > 0 {
					s.logo = png
					logos++
				}
			}
		}
	}
	t.infoReady = err == nil
	t.log.Printf("Token details loaded: total supply for %d tokens, %d logos", len(infos), logos)
}

// holdersLoop counts holders now, then again every HoldersEvery, until ctx ends.
func (t *Tracker) holdersLoop(ctx context.Context) {
	for {
		t.loadHolders(ctx)
		if err := sleepCtx(ctx, HoldersEvery); err != nil {
			return
		}
	}
}

// loadHolders counts the holders of each token in the current top list:
// wallets with a stake key plus addresses without one. One slow asset_summary
// request per token.
func (t *Tracker) loadHolders(ctx context.Context) {
	top := t.Quotes()
	t.log.Printf("Counting holders for %d tokens in the background (slow, can take an hour)…", len(top))
	for _, q := range top {
		if ctx.Err() != nil {
			return
		}
		var s *tokenState
		t.mu.RLock()
		for _, st := range t.states {
			if st.slug == q.ID {
				s = st
			}
		}
		t.mu.RUnlock()
		if s == nil {
			continue
		}
		start := time.Now()
		sum, err := t.client.AssetSummary(ctx, s.token.Asset)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.log.Printf("  %-8s could not count holders: %v (trying again in %s)", q.Ticker, err, HoldersEvery)
			}
			continue
		}
		n := sum.StakedWallets + sum.UnstakedAddresses
		t.mu.Lock()
		s.holders = n
		t.mu.Unlock()
		t.log.Printf("  %-8s %s holders (counted in %s)", q.Ticker, group(float64(n)), time.Since(start).Round(time.Second))
	}
}

// safeURL keeps a registry website only if it is a plain http(s) link, so a
// token's metadata cannot put another kind of link on the page.
func safeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return ""
	}
	return u.String()
}
