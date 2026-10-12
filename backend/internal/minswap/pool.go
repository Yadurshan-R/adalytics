// Package minswap reads Minswap V2 liquidity pools from Koios UTxOs.
//
// Each pool is one UTxO at the pool script. Its inline datum holds the pool
// state: field 1 = asset A, field 2 = asset B, field 4 = reserve A,
// field 5 = reserve B. ADA always sorts first, so in an ADA pair asset A is ADA.
package minswap

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"cardano-analytics/internal/koios"
)

const (
	// PoolCredential is the payment credential (script hash) of every Minswap V2 pool.
	PoolCredential = "ea07b733d932129c378af627436e7cbc2ef0bf96e0036bb51b3bde6b"
	// LPPolicy is the policy of the pool's LP tokens and of its "MSP" authentication token.
	LPPolicy      = "f5808c2c990d86da54bfc97d89cee6efa20cd8461616359478d96b4c"
	authAssetName = "4d5350" // "MSP"
)

// Pool is one Minswap V2 liquidity pool at one moment: which two assets it
// holds and how much of each. Prices come from the ratio of the reserves.
type Pool struct {
	UTxO      string      // tx_hash#index of the current pool output
	LPAsset   koios.Asset // unique per pool; used to fetch this pool again
	AssetA    koios.Asset
	AssetB    koios.Asset
	Liquidity *big.Int // total LP supply; changes on deposits/withdrawals, not on swaps
	ReserveA  *big.Int
	ReserveB  *big.Int
	DecimalsB int
	LastSwap  time.Time
}

// Parse turns one Koios UTxO into a Pool.
func Parse(u koios.UTxO) (Pool, error) {
	p := Pool{UTxO: fmt.Sprintf("%s#%d", u.TxHash, u.TxIndex), LastSwap: time.Unix(u.BlockTime, 0)}
	if u.InlineDatum == nil {
		return p, errors.New("pool has no inline datum")
	}
	f := u.InlineDatum.Value.Fields
	if len(f) < 6 {
		return p, fmt.Errorf("pool datum has %d fields, expected at least 6", len(f))
	}

	var err error
	if p.AssetA, err = asset(f[1]); err != nil {
		return p, fmt.Errorf("asset A: %w", err)
	}
	if p.AssetB, err = asset(f[2]); err != nil {
		return p, fmt.Errorf("asset B: %w", err)
	}
	if p.Liquidity, err = integer(f[3]); err != nil {
		return p, fmt.Errorf("total liquidity: %w", err)
	}
	if p.ReserveA, err = integer(f[4]); err != nil {
		return p, fmt.Errorf("reserve A: %w", err)
	}
	if p.ReserveB, err = integer(f[5]); err != nil {
		return p, fmt.Errorf("reserve B: %w", err)
	}

	for _, a := range u.AssetList {
		if a.PolicyID == LPPolicy && a.AssetName != authAssetName {
			p.LPAsset = koios.Asset{PolicyID: a.PolicyID, AssetName: a.AssetName}
		}
		if a.PolicyID == p.AssetB.PolicyID && a.AssetName == p.AssetB.AssetName {
			p.DecimalsB = a.Decimals
		}
	}
	if p.LPAsset.PolicyID == "" {
		return p, errors.New("pool has no LP token")
	}
	return p, nil
}

// IsADAPair reports whether this is the ADA / token pool.
func (p Pool) IsADAPair(token koios.Asset) bool {
	return p.AssetA.IsADA() && p.AssetB == token
}

// PriceADA is the price of one whole token in ADA:
// (reserve A / 10^6) / (reserve B / 10^decimals).
func (p Pool) PriceADA() float64 {
	if p.ReserveB.Sign() == 0 {
		return 0
	}
	num := new(big.Int).Mul(p.ReserveA, pow10(p.DecimalsB))
	den := new(big.Int).Mul(p.ReserveB, pow10(6))
	f, _ := new(big.Rat).SetFrac(num, den).Float64()
	return f
}

// ReserveADA returns reserve A in ADA.
func (p Pool) ReserveADA() float64 { return scaled(p.ReserveA, 6) }

// ReserveToken returns reserve B in whole tokens (after the token's decimals).
func (p Pool) ReserveToken() float64 { return scaled(p.ReserveB, p.DecimalsB) }

// asset reads an asset (policy ID and name) from a pool datum field.
func asset(d koios.PlutusData) (koios.Asset, error) {
	if len(d.Fields) != 2 || d.Fields[0].Bytes == nil || d.Fields[1].Bytes == nil {
		return koios.Asset{}, errors.New("unexpected asset shape")
	}
	return koios.Asset{PolicyID: *d.Fields[0].Bytes, AssetName: *d.Fields[1].Bytes}, nil
}

// integer reads a whole number from a pool datum field. Reserves can be
// larger than 64 bits, so it returns a big.Int.
func integer(d koios.PlutusData) (*big.Int, error) {
	if d.Int == nil {
		return nil, errors.New("not an integer")
	}
	n, ok := new(big.Int).SetString(d.Int.String(), 10)
	if !ok {
		return nil, fmt.Errorf("bad integer %q", d.Int.String())
	}
	return n, nil
}

// pow10 returns 10 to the power n, used to apply a token's decimals.
func pow10(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }

// scaled turns an amount in the smallest unit into whole units, e.g.
// 1,500,000 lovelace with 6 decimals is 1.5 ADA.
func scaled(n *big.Int, decimals int) float64 {
	f, _ := new(big.Rat).SetFrac(n, pow10(decimals)).Float64()
	return f
}
