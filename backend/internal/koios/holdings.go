package koios

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// AddressAsset is how much of one native token an address holds.
type AddressAsset struct {
	PolicyID  string         `json:"policy_id"`
	AssetName string         `json:"asset_name"`
	Quantity  string         `json:"quantity"`   // raw units, before decimals
	AssetList []AddressAsset `json:"asset_list"` // older Koios versions nest the rows here
}

// NFTAddress returns the address holding an NFT (a token with a supply of
// one), from asset_nft_address. Protocols keep their state next to an NFT
// like this, so it finds their contract even after it has moved.
func (c *Client) NFTAddress(ctx context.Context, nft Asset) (string, error) {
	q := url.Values{}
	q.Set("_asset_policy", nft.PolicyID)
	q.Set("_asset_name", nft.AssetName)
	var out []struct {
		PaymentAddress string `json:"payment_address"`
	}
	if err := c.do(ctx, c.http, http.MethodGet, "/asset_nft_address?"+q.Encode(), nil, &out); err != nil {
		return "", err
	}
	if len(out) == 0 || out[0].PaymentAddress == "" {
		return "", errors.New("Koios found no holder for the NFT")
	}
	return out[0].PaymentAddress, nil
}

// AddressAssets returns every native token held at one address, from
// address_assets, as one flat list.
func (c *Client) AddressAssets(ctx context.Context, address string) ([]AddressAsset, error) {
	body := map[string]any{"_addresses": []string{address}}
	var rows []AddressAsset
	if err := c.do(ctx, c.http, http.MethodPost, "/address_assets", body, &rows); err != nil {
		return nil, err
	}
	var out []AddressAsset
	for _, r := range rows {
		if r.PolicyID != "" {
			out = append(out, r)
		}
		out = append(out, r.AssetList...)
	}
	return out, nil
}
