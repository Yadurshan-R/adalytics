# Market functions

Every function in `internal/market`, the package that reads the Minswap pools and keeps prices, history and trades in memory. Each one also has the same comment above it in the code. The rest of the backend is in [BACKEND_FUNCTIONS.md](BACKEND_FUNCTIONS.md).

## `internal/market/activity.go`

| Function | What it does |
|---|---|
| `Note` | Writes a message to the log and, while the first load is running, also keeps it for the loading page. It is exported so main can report Koios retries ("Koios is busy, trying again in 15s") the same way. Do not call Note while holding t.mu: it takes the lock itself. |
| `recentActivity` | Returns the kept messages, newest first, as a copy so the caller can use it after the lock is released. Call with t.mu held. |
| `timeLeft` | Guesses the seconds left in the first load from how long it has taken so far and how far it has got. It returns 0 (unknown) until at least 5% is done, because earlier guesses jump around too much. |

## `internal/market/backfill.go`

| Function | What it does |
|---|---|
| `backfillLoop` | Waits for the first load, then loads the rest of the history. |
| `backfill` | Loads the history older than the first 24 hours, newest first, one page of 1,000 transactions at a time, so 7 days are ready long before 30. A page that fails is tried again a minute later. |
| `finishHistory` | Marks the whole history span as loaded, so every chart range is ready. |

## `internal/market/candles.go`

| Function | What it does |
|---|---|
| `CandleSizeByKey` | Finds a candle size by its key (5m, 15m, 30m, 1h, 2h, 4h, 12h, 1D), ignoring case. 5 minute candles go back 2 days, 15 and 30 minutes 7 days, the others 30 days. |
| `Candles` | Returns one token's candles at the given size, oldest first. While older history is still loading it returns what is loaded so far, so the chart grows instead of waiting for all 30 days. |

## `internal/market/details.go`

| Function | What it does |
|---|---|
| `loadInfo` | Reads total supply, ticker and logo for every tracked token with asset_info, 25 tokens per request. If it fails, prices still work and it is tried again next minute. |
| `holdersLoop` | Counts holders now, then again every HoldersEvery, until ctx ends. |
| `loadHolders` | Counts the holders of each token in the current top list: wallets with a stake key plus addresses without one. One slow asset_summary request per token. |
| `safeURL` | Keeps a registry website only if it is a plain http(s) link, so a token's metadata cannot put another kind of link on the page. |

## `internal/market/format.go`

| Function | What it does |
|---|---|
| `scaled` | Turns an amount in the smallest unit into whole units, e.g. 1,500,000 lovelace with 6 decimals is 1.5 ADA. |
| `plural` | Returns "s" unless n is 1, for log messages like "2 swaps". |
| `FormatADA` | Shows a price with 4 significant digits, e.g. ₳0.002657. |
| `change` | Describes a price move for the log, e.g. "▲ 2.10%". |
| `group` | Formats a number as 1,977,518. |

## `internal/market/history.go`

| Function | What it does |
|---|---|
| `RangeByKey` | Finds a range by its key ("24H", "7D", "30D"), ignoring case. |
| `priceAt` | Returns the price at time t: the last swap at or before t. If every swap is after t, the first known price is used. |
| `bucket` | Turns swaps into evenly spaced chart points between from and to. Each point carries the last price in its interval (or the previous price if nothing traded) and the ADA volume traded in it. |
| `changePct` | The percentage change from the price at time t to now. |
| `volumeSince` | Adds up the ADA traded from time t (unix seconds) until now. |
| `candles` | Groups swaps into candlesticks between from and to. A slot with no swaps is a flat candle at the previous close, and each candle opens at the previous close so the candles join up. |

## `internal/market/pools.go`

| Function | What it does |
|---|---|
| `readPools` | Reads transactions with tx_info (50 per request) and returns, for each token, the pool state after each of its transactions, oldest first. |
| `merge` | Adds pool states to a token's history (skipping ones it already has), keeps the history in chain order and rebuilds prices, volume and trades. It returns how many states were new. Call with t.mu held. |
| `rebuild` | Works out a token's price points, volume and trades from its pool history, drops history older than the kept span (keeping the last state before it, so the price at the start is known) and updates the quote from the newest pool state. Call with t.mu held. |
| `tradeFrom` | Turns the change in a pool's reserves into a trade. ADA going in while tokens come out is a buy; the other way round is a sell. |

## `internal/market/progress.go`

| Function | What it does |
|---|---|
| `History` | Returns how much history is loaded and which chart ranges are ready. |
| `setBoot` | Changes the first load's progress safely while other code reads it. |
| `bootLoading` | Turns the first load's progress into steps and one percentage: listing 0-10%, reading 10-80%, quiet tokens 80-95%, details 95-100%. Call with t.mu held. |
| `rangeReady` | Reports whether the whole range is loaded. Call with t.mu held. |

## `internal/market/quotes.go`

| Function | What it does |
|---|---|
| `Quotes` | Returns the top TopN tokens ranked by 24h trading volume. Only tokens with a Minswap pool holding at least MinPoolADA are ranked; tokens that did not trade in 24h follow, by pool size. |
| `sparkline` | The price at every step from start to end, for the list's small lines. |
| `usdPerADA` | The ADA/USD rate from the USDM pool (1 USDM ≈ 1 USD), or 0 if that pool is not tracked. Call with t.mu held. |
| `Chart` | Returns one token's chart for a range (24H, 7D, 30D): evenly spaced line points, or candles when candles is true. The candle chart on the site uses `Candles` instead. |
| `Trades` | Returns a token's most recent trades of the last 24 hours, newest first, at most limit (and never more than MaxTrades). |
| `Logo` | Returns a token's PNG logo from the token registry, if it has one. |

## `internal/market/startup.go`

| Function | What it does |
|---|---|
| `loadHistory` | The first load, kept short so the site is usable within a few minutes: 1. every Minswap V2 transaction of the last 24 hours (address_txs, then tx_info 50 at a time): prices, volume and trades, and the pool of every token that traded; 2. the current pool of each token that did not trade in 24 hours (asset_utxos, 30 tokens per request). Older history is loaded afterwards in the background (backfill). |
| `findPools` | Asks Koios which UTxOs at the Minswap pool address hold each token's LP token: that UTxO is the token's pool, as it is now. |

## `internal/market/tokens.go`

| Function | What it does |
|---|---|
| `mustLoadTokens` | Reads the embedded tokens.json. It stops the program if the file is broken, because nothing can work without the token list. |

## `internal/market/tracker.go`

| Function | What it does |
|---|---|
| `chainOrder` | Gives one number for sorting transactions in the order they happened: block height first, then position inside the block. |
| `NewTracker` | Prepares a Tracker for the given tokens. Nothing is loaded until Run is called. |
| `SetHistoryDays` | Sets how many days of history are loaded and kept (1 to 30). Call it before Run. |
| `assignSlugs` | Gives each token a short unique id, normally its ticker in lower case; tokens sharing a ticker get part of their policy ID added. |
| `Run` | Loads history now, then checks for new swaps every interval, until ctx ends. Holder counts are loaded separately in the background. |
| `refresh` | Runs once a minute: the first time it does the first load, and after that it loads token details if they are still missing and checks for new swaps. |
| `tracked` | Returns the tokens whose pool was found and holds at least MinPoolADA: the ones listed and checked every minute. Call with t.mu held. |
| `sleepCtx` | Waits for d, or returns early with an error when ctx is cancelled. |
| `blocksIn` | About how many blocks Cardano makes in d. |
| `failAll` | Records an error that affects every token and logs it once. |

## `internal/market/update.go`

| Function | What it does |
|---|---|
| `update` | Checks the chain tip and, with utxo_info (50 pools per request), which pools were spent (a swap happened) since the previous check. Only then it reads Minswap's transactions since that check and takes the new pool states. |


# Market tests

Run them with `go test ./internal/market/` from `backend/`. They use a fake Koios server, so they need no internet and no API token.

## `internal/market/candles_test.go`

| Function | What it does |
|---|---|
| `TestCandleSizes` | Checks that each candle size covers all the history loaded so far: 24 hours at first, then 7 days once the background load is done, and that a wrong size or token is an error. |

## `internal/market/charts_test.go`

| Function | What it does |
|---|---|
| `TestBucketCarriesPriceForward` | Checks that chart steps with no trade keep the last price, and that volume is added up per step. |
| `TestFormatADA` | Checks prices are shown with 4 significant digits. |
| `TestCandles` | Checks open, high, low and close for each candle. |

## `internal/market/helpers_test.go`

| Function | What it does |
|---|---|
| `loadPools` | Reads real Minswap pools (SNEK/MIN and SNEK/ADA) saved from Koios. |
| `poolOutput` | The real SNEK/ADA pool from testdata with a chosen ADA reserve. |
| `swapOutput` | The SNEK/ADA pool with chosen LP supply and reserves (SNEK has 0 decimals). |
| `otherOutput` | An output of some other Minswap pool (not one of ours). |
| `serve` | Answers each Koios endpoint the Tracker uses from the fake's data, and records every request. |
| `snekToken` | SNEK from tokens.json. |
| `newTracker` | Starts a fake Koios server and a Tracker for SNEK that uses it, with the clock fixed at testNow. |

## `internal/market/history_test.go`

| Function | What it does |
|---|---|
| `TestBackfillLoadsOlderHistory` | Checks that 7 days of history are loaded in the background, and that the 7D chart and change appear only once they are. |
| `TestBackfillWhileUpdating` | The background history load runs while the minute updates and the API read the same tokens; run them together so the race detector can check it. |
| `TestLoadingProgress` | Checks the loading screen's steps and percentage. |

## `internal/market/quotes_test.go`

| Function | What it does |
|---|---|
| `TestTradesFromSwaps` | Checks that buys and sells are worked out from pool changes, and that dust and liquidity deposits are not counted as trades. |
| `TestUSDPriceFromUSDMPool` | Checks USD prices come from the USDM pool, and are left empty when that pool is missing. |
| `TestSafeURL` | Checks that only plain http(s) website links are kept. |
| `TestRankingAndIDs` | Checks the ranking by volume then pool size, that small pools are left out, and that tokens sharing a ticker get unique ids. |

## `internal/market/tracker_test.go`

| Function | What it does |
|---|---|
| `TestLoadsHistoryThenFollowsSwaps` | Checks the first 24 hour load (price, volume, change, chart), then that a new swap is picked up a minute later. |
| `TestQuietTokenIsListed` | Checks that a token with a pool but no trades today is still listed, with a flat price line. |
| `TestTokenWithoutPoolIsLeftOut` | Checks that a token with no Minswap pool is not listed. |
| `TestFindsAllPoolsInSmallRequests` | Checks that the pool search is split into requests small enough for the Koios free tier. |
