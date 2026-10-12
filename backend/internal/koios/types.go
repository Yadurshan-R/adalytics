package koios

import (
	"encoding/json"
	"fmt"
)

// Asset identifies a native token. An empty PolicyID means ADA.
type Asset struct {
	PolicyID  string `json:"policy_id"`
	AssetName string `json:"asset_name"`
}

// IsADA reports whether the asset is ADA itself rather than a native token.
func (a Asset) IsADA() bool { return a.PolicyID == "" && a.AssetName == "" }

// AssetAmount is how much of one native token a UTxO holds.
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

// InlineDatum is the data attached to a UTxO: raw bytes and decoded value.
type InlineDatum struct {
	Bytes string     `json:"bytes"`
	Value PlutusData `json:"value"`
}

// UTxO is one unspent transaction output: an address, the ADA and tokens it
// holds, its datum, and the block it was created in.
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

// TxOutput is one output of a transaction, as tx_info returns it.
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

// Tip is the newest block Koios has indexed.
type Tip struct {
	BlockHeight int   `json:"block_height"`
	BlockTime   int64 `json:"block_time"`
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

// AssetSummary counts a token's holders. Koios takes a minute or more for big tokens.
type AssetSummary struct {
	StakedWallets     int `json:"staked_wallets"`
	UnstakedAddresses int `json:"unstaked_addresses"`
}
