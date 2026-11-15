package market

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"cardano-analytics/internal/koios"
	"cardano-analytics/internal/minswap"
)

// Token is a Cardano native token the platform can track.
type Token struct {
	Ticker   string
	Name     string
	Asset    koios.Asset
	Decimals int
	// PoolLP is the LP token name of the token's Minswap V2 ADA pool. Minswap
	// derives it from the two assets: sha3_256(sha3_256(ADA) + sha3_256(policy+name)).
	PoolLP string
}

// tokens.json lists the tokens marked verified in Minswap's open token list
// (github.com/minswap/minswap-tokens), with each pool's LP token name
// precomputed. Which of them are shown, and in what order, is decided live
// from Koios data: see TopN and MinPoolADA.
//
//go:embed tokens.json
var tokensJSON []byte

// Tokens is every candidate token, loaded from tokens.json.
var Tokens = mustLoadTokens()

// mustLoadTokens reads the embedded tokens.json. It stops the program if the
// file is broken, because nothing can work without the token list.
func mustLoadTokens() []Token {
	var raw []struct {
		PolicyID  string `json:"policy_id"`
		AssetName string `json:"asset_name"`
		Name      string `json:"name"`
		Ticker    string `json:"ticker"`
		Decimals  int    `json:"decimals"`
		PoolLP    string `json:"pool_lp"`
	}
	if err := json.Unmarshal(tokensJSON, &raw); err != nil {
		panic(fmt.Sprintf("tokens.json: %v", err))
	}
	out := make([]Token, len(raw))
	for i, r := range raw {
		out[i] = Token{
			Ticker: r.Ticker, Name: r.Name, Decimals: r.Decimals, PoolLP: r.PoolLP,
			Asset: koios.Asset{PolicyID: r.PolicyID, AssetName: r.AssetName},
		}
	}
	return out
}

// withDecimals fills in the token's decimals on a pool read from the chain
// when Koios gave none. Koios knows decimals from the Cardano Token Registry
// and from CIP-68 metadata it can read; for a few CIP-68 tokens (SUPERIOR,
// BOS, FNX) it says 0, and a price and supply off by 10^decimals would show
// (SUPERIOR: 8 decimals, so a price 100,000,000 times too small). The token
// list (tokens.json, from Minswap's verified token list) has their decimals,
// the same Minswap shows. A token Koios gives decimals for keeps Koios's.
func (t Token) withDecimals(p minswap.Pool) minswap.Pool {
	if p.DecimalsB == 0 && t.Decimals > 0 {
		p.DecimalsB = t.Decimals
	}
	return p
}
