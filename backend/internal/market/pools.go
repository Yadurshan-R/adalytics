package market

import (
	"context"
	"math"
	"math/big"
	"sort"
	"strings"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// readPools reads transactions with tx_info (50 per request) and returns, for
// each token, the pool state after each of its transactions, oldest first.
func (t *Tracker) readPools(ctx context.Context, txs []koios.AddressTx, progress func(done, total int)) (map[*tokenState][]poolState, error) {
	byLP := map[string]*tokenState{}
	for _, s := range t.states {
		byLP[s.lp.AssetName] = s
	}

	var all []koios.Tx
	for start := 0; start < len(txs); start += 50 {
		end := min(start+50, len(txs))
		hashes := make([]string, 0, end-start)
		for _, tx := range txs[start:end] {
			hashes = append(hashes, tx.TxHash)
		}
		batch, err := t.client.TxInfo(ctx, hashes)
		if err != nil {
			return nil, err
		}
		all = append(all, batch...)
		if progress != nil {
			progress(end, len(txs))
		}
	}

	found := map[*tokenState][]poolState{}
	for _, tx := range all {
		for _, o := range tx.Outputs {
			if o.PaymentAddr.Cred != minswap.PoolCredential {
				continue
			}
			p, err := minswap.Parse(o.UTxO(tx))
			if err != nil {
				continue
			}
			s, ok := byLP[p.LPAsset.AssetName]
			if !ok || !p.IsADAPair(s.token.Asset) {
				continue
			}
			found[s] = append(found[s], poolState{
				pool: p, height: tx.BlockHeight, order: chainOrder(tx.BlockHeight, tx.TxBlockIndex), address: o.PaymentAddr.Bech32,
			})
		}
	}
	for _, states := range found {
		sort.SliceStable(states, func(i, j int) bool { return states[i].order < states[j].order })
	}
	return found, nil
}

// merge adds pool states to a token's history (skipping ones it already has),
// keeps the history in chain order and rebuilds prices, volume and trades.
// It returns how many states were new. Call with t.mu held.
func (t *Tracker) merge(s *tokenState, states []poolState) int {
	if s.refs == nil {
		s.refs = map[string]bool{}
	}
	added := 0
	for _, st := range states {
		if s.refs[st.pool.UTxO] {
			continue
		}
		s.refs[st.pool.UTxO] = true
		s.history = append(s.history, st)
		if s.address == "" {
			s.address = st.address
		}
		added++
	}
	if added == 0 {
		return 0
	}
	sort.SliceStable(s.history, func(i, j int) bool { return s.history[i].order < s.history[j].order })
	t.rebuild(s)
	return added
}

// rebuild works out a token's price points, volume and trades from its pool
// history, drops history older than the kept span (keeping the last state
// before it, so the price at the start is known) and updates the quote from the
// newest pool state. Call with t.mu held.
func (t *Tracker) rebuild(s *tokenState) {
	now := t.now()
	cut := now.Add(-t.span).Unix()
	drop := 0
	for drop+1 < len(s.history) && s.history[drop+1].pool.LastSwap.Unix() <= cut {
		delete(s.refs, s.history[drop].pool.UTxO)
		drop++
	}
	if drop > 0 {
		s.history = append([]poolState(nil), s.history[drop:]...)
	}

	s.points, s.trades = s.points[:0], s.trades[:0]
	var prev *minswap.Pool
	for i := range s.history {
		p := &s.history[i].pool
		vol := 0.0
		// A swap moves ADA in or out without changing the LP supply; adding or
		// removing liquidity changes the LP supply and is not a trade.
		if prev != nil && p.Liquidity.Cmp(prev.Liquidity) == 0 {
			diffA := new(big.Int).Sub(p.ReserveA, prev.ReserveA)
			diffB := new(big.Int).Sub(p.ReserveB, prev.ReserveB)
			vol = math.Abs(scaled(diffA, 6))
			if tr, ok := tradeFrom(*p, diffA, diffB); ok {
				s.trades = append(s.trades, tr)
			}
		}
		s.points = append(s.points, Point{Time: p.LastSwap.Unix(), Price: p.PriceADA(), VolumeADA: vol})
		prev = p
	}

	last := s.history[len(s.history)-1]
	p := last.pool
	lastSwap := p.LastSwap
	s.lastHeight = last.height
	s.poolRef = p.UTxO
	s.quote.Status, s.quote.Error = "live", ""
	s.quote.PriceADA = p.PriceADA()
	s.quote.ReserveADA = p.ReserveADA()
	s.quote.ReserveToken = p.ReserveToken()
	s.quote.Decimals = p.DecimalsB
	s.quote.PoolUTxO = p.UTxO
	s.quote.LastSwap = &lastSwap
	s.quote.UpdatedAt = &now
}

// tradeFrom turns the change in a pool's reserves into a trade. ADA going in
// while tokens come out is a buy; the other way round is a sell.
func tradeFrom(p minswap.Pool, diffADA, diffToken *big.Int) (Trade, bool) {
	if diffADA.Sign() == 0 || diffToken.Sign() == 0 || diffADA.Sign() == diffToken.Sign() {
		return Trade{}, false
	}
	ada := math.Abs(scaled(diffADA, 6))
	tok := math.Abs(scaled(diffToken, p.DecimalsB))
	if ada < MinTradeADA || tok == 0 {
		return Trade{}, false
	}
	side := "sell"
	if diffADA.Sign() > 0 {
		side = "buy"
	}
	hash, _, _ := strings.Cut(p.UTxO, "#")
	return Trade{Time: p.LastSwap, Side: side, PriceADA: ada / tok, AmountADA: ada, AmountToken: tok, TxHash: hash}, true
}
