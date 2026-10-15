package koios

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// Tip returns the newest block Koios knows about: its height and time.
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
// Requests are split so each body stays under Koios's size limit, and run
// three at a time (Workers).
func (c *Client) PoolUTxOs(ctx context.Context, address string, assets []Asset, progress func(done, total int)) ([]UTxO, error) {
	const perRequest = 30 // 30 LP tokens make a body of about 3.9 KB
	q := url.Values{}
	q.Set("select", "tx_hash,tx_index,address,value,block_height,block_time,asset_list,inline_datum")
	q.Set("address", "eq."+address)
	batches := (len(assets) + perRequest - 1) / perRequest
	found := make([][]UTxO, batches)
	var mu sync.Mutex
	done := 0
	err := Parallel(ctx, batches, Workers, func(i int) error {
		start := i * perRequest
		end := min(start+perRequest, len(assets))
		list := make([][2]string, 0, end-start)
		for _, a := range assets[start:end] {
			list = append(list, [2]string{a.PolicyID, a.AssetName})
		}
		body := map[string]any{"_asset_list": list, "_extended": true}
		if err := c.do(ctx, c.http, http.MethodPost, "/asset_utxos?"+q.Encode(), body, &found[i]); err != nil {
			return err
		}
		mu.Lock()
		done += end - start
		if progress != nil {
			progress(done, len(assets))
		}
		mu.Unlock()
		return nil
	})
	if err != nil {
		return nil, err
	}
	var all []UTxO
	for _, rows := range found {
		all = append(all, rows...)
	}
	return all, nil
}

// AddressTxsPageSize is the most rows Koios returns in one page.
const AddressTxsPageSize = 1000

// AddressTxsPage returns one page of the transactions that touched the
// addresses from a block height on, newest first.
func (c *Client) AddressTxsPage(ctx context.Context, addresses []string, afterBlockHeight, offset int) ([]AddressTx, error) {
	body := map[string]any{"_addresses": addresses, "_after_block_height": afterBlockHeight}
	q := url.Values{}
	q.Set("order", "block_height.desc,tx_hash.asc")
	q.Set("limit", strconv.Itoa(AddressTxsPageSize))
	q.Set("offset", strconv.Itoa(offset))
	var rows []AddressTx
	err := c.do(ctx, c.http, http.MethodPost, "/address_txs?"+q.Encode(), body, &rows)
	return rows, err
}

// AddressTxs lists every transaction that touched the addresses from a block
// height on (newest first). Koios returns at most 1000 rows per page, so it
// pages through; progress, if set, is called after each page.
func (c *Client) AddressTxs(ctx context.Context, addresses []string, afterBlockHeight int, progress func(rows int)) ([]AddressTx, error) {
	const page = AddressTxsPageSize
	var all []AddressTx
	seen := map[string]bool{}
	for offset := 0; ; offset += page {
		rows, err := c.AddressTxsPage(ctx, addresses, afterBlockHeight, offset)
		if err != nil {
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

// ScriptInfos reads the scripts with the given hashes (up to 25 per call).
// Hashes Koios does not know are left out of the answer.
func (c *Client) ScriptInfos(ctx context.Context, hashes []string) ([]ScriptInfo, error) {
	body := map[string]any{"_script_hashes": hashes}
	var out []ScriptInfo
	err := c.do(ctx, c.http, http.MethodPost, "/script_info?select=script_hash,type,value", body, &out)
	return out, err
}

// AssetSummary asks Koios how many wallets hold a token. It uses the slow
// HTTP client, because Koios can take minutes to count big tokens.
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

// TxInfo returns transactions with their inputs and outputs, assets and inline
// datums. The inputs show which wallets' orders a batch swapped.
func (c *Client) TxInfo(ctx context.Context, hashes []string) ([]Tx, error) {
	body := map[string]any{
		"_tx_hashes": hashes,
		"_inputs":    true, "_metadata": false, "_assets": true, "_withdrawals": false,
		"_certs": false, "_scripts": true, "_bytecode": false, "_governance": false,
	}
	var out []Tx
	err := c.do(ctx, c.http, http.MethodPost, "/tx_info?select=tx_hash,block_height,tx_timestamp,tx_block_index,inputs,outputs", body, &out)
	return out, err
}
