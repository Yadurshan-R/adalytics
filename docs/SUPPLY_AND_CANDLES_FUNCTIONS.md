# Supply and candle functions

The part of `internal/market` that works out a token's supply and decimals, and builds chart points and candles. Each function also has the same comment above it in the code. The rest of the package is in [MARKET_FUNCTIONS.md](MARKET_FUNCTIONS.md); the tests are in [MARKET_TESTS.md](MARKET_TESTS.md).

## `internal/market/details.go`

| Function | What it does |
|---|---|
| `loadInfo` | Reads total supply, ticker and logo for every tracked token with asset_info, and each minting policy with script_info, 25 tokens per request. If it fails, prices still work and it is tried again next minute. |
| `safeURL` | Keeps a registry website only if it is a plain http(s) link, so a token's metadata cannot put another kind of link on the page. |

## `internal/market/tokens.go`

| Function | What it does |
|---|---|
| `mustLoadTokens` | Reads the embedded tokens.json. It stops the program if the file is broken, because nothing can work without the token list. |
| `withDecimals` | Fills in the token's decimals on a pool read from the chain when Koios gave none. Koios knows decimals from the Cardano Token Registry and from the CIP-68 metadata it can read; for a few CIP-68 tokens (SUPERIOR, BOS, FNX) it says 0, which would show a price and supply off by 10^decimals (SUPERIOR has 8). The token list (`tokens.json`, from Minswap's verified list) has their decimals, the same Minswap shows. A token Koios gives decimals for keeps Koios's. |

## `internal/market/reserves.go`

Some protocols mint their tokens' whole supply up front and keep what is not given out in their own contract, next to an NFT. Djed did this: 1,000,000,000,000 DJED and SHEN each, while on 10 Nov 2026 only 1,410,269 DJED and 31,686,386 SHEN were outside the reserve (the same as Minswap shows). The supply shown, and the FDV, leave the reserve out.

| Function | What it does |
|---|---|
| `reserveOf` | Returns the protocol whose reserve holds most of a token's supply (any token under its policy except the NFT itself), or nil. The list (`protocolReserves`) has Djed, found through its DjedStableCoinNFT, so it is still found if Djed moves its contract. |
| `issued` | The supply given out: the on-chain total supply, minus the protocol reserve for the tokens that have one. Nil while that is not known (supply not loaded, reserve not read yet, or nothing of the token found there), so no supply or FDV is shown rather than a wrong one. |
| `loadDetails` | Runs before the list opens and every minute after: token details until they have loaded once (`loadInfo`), and the reserves (`loadReserves`). |
| `loadReserves` | Reads, at most every 10 minutes (`ReservesEvery`), how much of each tracked token sits in its protocol's reserve. If it fails, the old amounts stay and it is tried again next minute. |
| `readReserve` | Finds the address holding the protocol's NFT (Koios `asset_nft_address`) and adds up the protocol's tokens at that address (`address_assets`). |

## `internal/market/history.go`

| Function | What it does |
|---|---|
| `RangeByKey` | Finds a range by its key ("24H", "7D", "30D"), ignoring case. |
| `priceAt` | Returns the price at time t: the last swap at or before t. If every swap is after t, the first known price is used. |
| `bucket` | Turns swaps into evenly spaced chart points between from and to. Each point carries the last price in its interval (or the previous price if nothing traded) and the ADA volume traded in it. |
| `changePct` | The percentage change from the price at time t to now. |
| `volumeSince` | Adds up the ADA traded from time t (unix seconds) until now. |
| `candles` / `candlesAt` | Groups swaps into candlesticks; `candlesAt` takes the candle starts (from `candleStarts`), so weeks and months work the same as fixed sizes. A slot with no swaps is a flat candle at the previous close, and each candle opens at the previous close so the candles join up. Open and close are pool prices, like the Price card; high and low also take the prices orders were filled at (`Point.Low`, `Point.High`), so the wicks show what traders actually paid. |

## `internal/market/points.go`

| Function | What it does |
|---|---|
| `pointOf` | Makes the chart point for one pool state: the pool's price right after the transaction, the ADA it traded (`tradeVolume`), and the lowest and highest price its orders were filled at (`tradeRange`), but only when the batch was split into orders. |
| `tradeRange` | The lowest and highest price trades were filled at. These make a candle's wicks: a buyer pays a little above the pool price (the 0.3% pool fee plus slippage), a seller gets a little below, and orders in one batch can move the price up and back down, which the pool price after the batch does not show. Birdeye and DexScreener build highs and lows from trade prices the same way. Only batches read order by order are used: an unsplit batch is one trade of the pool's net change, which can be mostly fees (₳60 in for 1 token out would read as ₳60 a token). |
| `tradeVolume` | The ADA a pool state traded: the sum of its trades, so every order of a Minswap batch counts on its own (gross volume, as Minswap and DexScreener count it). The pool's own change nets buys against sells: a ₳1,000 buy and a ₳800 sell in one batch move the pool by ₳200 but trade ₳1,800. |
