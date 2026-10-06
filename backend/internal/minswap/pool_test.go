package minswap

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"cardano-analytics/internal/koios"
)

var snek = koios.Asset{PolicyID: "279c909f348e533da5808898f87f9a14bb2c3dfbbacccd631d927a3f", AssetName: "534e454b"}

// testdata/snek_pools.json is real Koios output (SNEK/MIN pool and SNEK/ADA pool).
func loadFixture(t *testing.T) []koios.UTxO {
	t.Helper()
	data, err := os.ReadFile("testdata/snek_pools.json")
	if err != nil {
		t.Fatal(err)
	}
	var utxos []koios.UTxO
	if err := json.Unmarshal(data, &utxos); err != nil {
		t.Fatal(err)
	}
	return utxos
}

func TestParseSNEKADAPool(t *testing.T) {
	utxos := loadFixture(t)

	minPool, err := Parse(utxos[0])
	if err != nil {
		t.Fatal(err)
	}
	if minPool.IsADAPair(snek) {
		t.Error("SNEK/MIN pool must not count as an ADA pair")
	}

	p, err := Parse(utxos[1])
	if err != nil {
		t.Fatal(err)
	}
	if !p.IsADAPair(snek) {
		t.Fatal("SNEK/ADA pool not recognised")
	}
	if got := p.ReserveA.String(); got != "1977518866875" {
		t.Errorf("ADA reserve = %s", got)
	}
	if got := p.ReserveB.String(); got != "744260765" {
		t.Errorf("SNEK reserve = %s", got)
	}
	if p.LPAsset.AssetName != "2ffadbb87144e875749122e0bbb9f535eeaa7f5660c6c4a91bcc4121e477f08d" {
		t.Errorf("LP asset = %s", p.LPAsset.AssetName)
	}
	if got, want := p.PriceADA(), 0.002657024204245134; math.Abs(got-want) > 1e-15 {
		t.Errorf("price = %v, want %v", got, want)
	}
}

func TestPriceUsesTokenDecimals(t *testing.T) {
	// 100 ADA against 50 tokens with 6 decimals = 2 ADA per token.
	utxos := loadFixture(t)
	p, _ := Parse(utxos[1])
	p.ReserveA.SetInt64(100_000_000)
	p.ReserveB.SetInt64(50_000_000)
	p.DecimalsB = 6
	if got := p.PriceADA(); got != 2 {
		t.Errorf("price = %v, want 2", got)
	}
}
