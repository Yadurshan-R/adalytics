package market

import (
	"context"
	"errors"
	"math/big"
	"time"

	"cardano-analytics/internal/koios"
)

// protocolReserve is a protocol that minted its tokens' whole supply up front
// and keeps the part not given out yet in its own contract, next to an NFT.
// The on-chain total supply then counts tokens nobody holds, so the supply
// shown (and the FDV, price × supply) leaves out what sits in the reserve.
type protocolReserve struct {
	name string
	nft  koios.Asset // the one-of-a-kind token kept at the reserve
}

// protocolReserves lists such protocols. Djed pre-minted 1,000,000,000,000
// DJED and SHEN each; the Djed contract hands them out and takes them back
// instead of minting and burning (forum.cardano.org/t/113798). Checked on
// 10 Nov 2026: 1,410,269 DJED and 31,686,386 SHEN were outside the reserve,
// the same as Minswap shows. The reserve is found through its NFT
// (DjedStableCoinNFT), so it is still found if Djed moves its contract.
var protocolReserves = []protocolReserve{{
	name: "Djed",
	nft: koios.Asset{PolicyID: "8db269c3ec630e06ae29f74bc39edd1f87c819f1056206e879a1cd61",
		AssetName: "446a6564537461626c65436f696e4e4654"}, // DjedStableCoinNFT
}}

// ReservesEvery is how often the reserves are read again: the amount given
// out changes whenever someone mints or redeems, and each read is only two
// Koios requests per protocol.
const ReservesEvery = 10 * time.Minute

// reserveOf returns the protocol whose reserve holds most of a token's
// supply (any token under its policy except the NFT itself), or nil.
func reserveOf(a koios.Asset) *protocolReserve {
	for i, r := range protocolReserves {
		if a.PolicyID == r.nft.PolicyID && a.AssetName != r.nft.AssetName {
			return &protocolReserves[i]
		}
	}
	return nil
}

// issued returns the supply given out, in the smallest unit: the on-chain
// total supply, minus the protocol reserve for the tokens that have one. It
// is nil while that is not known (supply not loaded, or the reserve not read
// yet), so no supply or FDV is shown rather than a wrong one.
func (s *tokenState) issued() *big.Int {
	if s.supply == nil || reserveOf(s.token.Asset) == nil {
		return s.supply
	}
	if s.reserved == nil || s.reserved.Cmp(s.supply) > 0 {
		return nil
	}
	return new(big.Int).Sub(s.supply, s.reserved)
}

// loadDetails runs before the list opens and every minute after: token
// details until they have loaded once (loadInfo), and the reserves
// (loadReserves).
func (t *Tracker) loadDetails(ctx context.Context) {
	if !t.infoReady {
		t.loadInfo(ctx)
	}
	t.loadReserves(ctx)
}

// loadReserves reads, at most every ReservesEvery, how much of each tracked
// token sits in its protocol's reserve: the address holding the protocol's
// NFT (Koios asset_nft_address), then the tokens at that address
// (address_assets). If it fails, the old amounts stay and it is tried again
// next minute. A token with nothing at the address keeps no amount, so it
// shows no supply instead of the full pre-minted one.
func (t *Tracker) loadReserves(ctx context.Context) {
	if !t.reservesAt.IsZero() && t.now().Sub(t.reservesAt) < ReservesEvery {
		return
	}
	need := map[string]*protocolReserve{}
	t.mu.RLock()
	for _, s := range t.tracked() {
		if r := reserveOf(s.token.Asset); r != nil {
			need[r.name] = r
		}
	}
	t.mu.RUnlock()

	held := map[koios.Asset]*big.Int{}
	for _, r := range need {
		if err := t.readReserve(ctx, r, held); err != nil {
			if !errors.Is(err, context.Canceled) {
				t.log.Printf("Could not read the %s reserve: %v (trying again in %s)", r.name, err, t.interval)
			}
			return
		}
	}
	t.mu.Lock()
	for _, s := range t.states {
		if reserveOf(s.token.Asset) != nil {
			s.reserved = held[s.token.Asset] // nil if none there
		}
	}
	t.mu.Unlock()
	t.reservesAt = t.now()
}

// readReserve adds what protocol r's reserve holds of each of its tokens to
// held, in the smallest unit.
func (t *Tracker) readReserve(ctx context.Context, r *protocolReserve, held map[koios.Asset]*big.Int) error {
	addr, err := t.client.NFTAddress(ctx, r.nft)
	if err != nil {
		return err
	}
	assets, err := t.client.AddressAssets(ctx, addr)
	if err != nil {
		return err
	}
	for _, a := range assets {
		n, ok := new(big.Int).SetString(a.Quantity, 10)
		key := koios.Asset{PolicyID: a.PolicyID, AssetName: a.AssetName}
		if !ok || reserveOf(key) != r {
			continue
		}
		if held[key] == nil {
			held[key] = new(big.Int)
		}
		held[key].Add(held[key], n)
	}
	return nil
}
