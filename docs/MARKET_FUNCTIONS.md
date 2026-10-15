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
| `backfillLoop` | Waits for the first 24 hours, then loads the rest of the history. |
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

## `internal/market/periods.go`

| Function | What it does |
|---|---|
| `Periods` | One token's trading over the last 5 minutes, hour, 4 hours and 24 hours: price change, volume, buys and sells and the ADA on each side, from the trades already in memory. |
| `periodsOf` | Works the periods out for one token; the token list (Quotes) and the activity box both use it. |

## `internal/market/pools.go`

| Function | What it does |
|---|---|
| `readPools` | Reads transactions with tx_info (50 per request, three requests at a time) and returns, for each token, the pool state after each of its transactions, oldest first. |
| `merge` | Adds pool states to a token's history (skipping ones it already has), keeps the history in chain order and rebuilds prices, volume and trades. It returns how many states were new. Call with t.mu held. |
| `rebuild` | Works out a token's price points, volume and trades from its pool history, drops history older than the kept span (keeping the last state before it, so the price at the start is known) and updates the quote from the newest pool state. Call with t.mu held. |
| `replaceCurrent` | Swaps a pool state found at startup (as the pool is now) for the same pool output read from its transaction, which has the wallets and its place in the block. |
| `tradesFrom` | Turns a swap into trades: one per wallet when the transaction was split into fills, otherwise one trade without a wallet from the change in the pool's reserves. |
| `tradeFrom` | Makes one trade from ADA into the pool and tokens to the trader: both positive is a buy, both negative a sell. Dust below MinTradeADA is left out. |

## `internal/market/fills.go`

Minswap does not swap each order on its own: a batcher swaps several people's orders against the pool in one transaction. These functions split such a transaction into one trade per order. The trader is the wallet the payout goes to: an order's own address often has no stake key, or Minswap's. A routed swap (e.g. USDA → ADA → NIGHT) passes through more than one pool; with one order, the token's own pool change is that wallet's trade.

| Function | What it does |
|---|---|
| `fillsFrom` | Finds the token's pool by its LP asset, then sorts the transaction into pools, the batcher (its coins go in and come out at the same address), the orders and one payout per order, paired in order. Each pair's payout minus order, less its share of the batcher fee, is what went into or out of the pool. Returns nil, so the swap shows as one trade without a wallet, when anything does not add up: as many payouts as orders, one order only when the swap is routed through other pools, every token accounted for, a fee of ₳0 to ₳10 per order, and with several orders each one near the pool's price. |
| `findPool` | The index of the pool holding a given LP asset, or -1. |
| `isPool` | Whether an input or output is a Minswap pool. |
| `walletOf` | The wallet a payout goes to: its stake address, or its payment address when it has none. |
| `nearPrice` | Whether a pair's ADA per token is within 20% of the pool's price before the batch; shows orders and payouts were paired right. |
| `lovelace` | The ADA in an output, in lovelace. |
| `units` | How much of one token an output holds, in its smallest unit. |
| `sub` | a − b as a new number. |

## `internal/market/wallets.go`

| Function | What it does |
|---|---|
| `WhaleADA` | The smallest whale trade for a token: 1% of its pool's ADA (WhaleShare). A share of the pool fits big and small tokens alike: either way such a trade moves the price about 2%. |
| `Wallet` | Every token one wallet traded in the last 24 hours (buys, sells, ADA each way, last trade), latest first, from the trades in memory. |

## `internal/market/progress.go`

| Function | What it does |
|---|---|
| `History` | Returns how much history is loaded and which chart ranges are ready. |
| `setBoot` | Changes the first load's progress safely while other code reads it. |
| `bootLoading` | Turns the first load's progress into steps and one percentage: listing 0-20%, the newest hour 20-35%, other tokens' pools 35-90%, details 90-100%. Call with t.mu held. |
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
| `loadHistory` | The first load, kept short so the site opens within about half a minute: 1. list every Minswap V2 transaction of the last 24 hours (address_txs, newest first); 2. read the newest hour of them (tx_info): prices, 5m and 1h numbers; 3. the current pool of every other token (asset_utxos, 30 per request); 4. supply, logos and links. The list opens here; the rest of the 24 hours is read in the background (readDay), then the older history (backfill). |
| `findPools` | Asks Koios which UTxOs at the Minswap pool address hold each token's LP token: that UTxO is the token's pool, as it is now. |

## `internal/market/day.go`

| Function | What it does |
|---|---|
| `readDay` | Reads the rest of the first 24 hours after the list has opened, newest first, 300 transactions at a time (DayChunk). After each chunk the site has more of the day: 4h numbers after a few chunks, 24h at the end. A chunk that fails is tried again a minute later. Then the older history starts. |

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

## `internal/market/periods_test.go`

| Function | What it does |
|---|---|
| `TestPeriods` | Checks buys, sells, ADA and the price change for each period, and that an unknown token is an error. |

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

## `internal/market/day_test.go`

| Function | What it does |
|---|---|
| `TestListOpensAfterTheNewestHour` | The list opens with the newest hour read: the price is right at once and the 24H range waits. Once the rest of the day is read, the 24h volume is complete and the pool found at startup is not counted twice. |
| `TestQuietTokenIsListed` | Checks that a token with a pool but no trades today is still listed, with a flat price line. |
| `TestTokenWithoutPoolIsLeftOut` | Checks that a token with no Minswap pool is not listed. |
| `TestFindsAllPoolsInSmallRequests` | Checks that the pool search is split into requests small enough for the Koios free tier. |

## `internal/market/fills_test.go`

| Function | What it does |
|---|---|
| `TestFillsRealSwaps` | Three real SNEK swaps from mainnet: a ₳487 buy, then one wallet selling twice. Each gives the right wallet, ADA and tokens. |
| `TestFillsOrderWithoutOwnerKey` | Three more real SNEK swaps whose orders carry no stake key, or the pool's: the trader is the wallet the payout goes to. |
| `TestFillsRouted` | Two real NIGHT swaps routed through a stablecoin pool (a buy and a sell): NIGHT's own pool change is the wallet's trade. |
| `TestFillsBatch` | Three wallets in one batch (two buys and a sell) become three trades that add up to the pool's change. |
| `TestFillsNotSplit` | Transactions that cannot be split for certain give no fills: no inputs, tokens missing, a fee too big, an order without a payout, two orders routed through two pools, payouts paired the wrong way round. |
| `TestTradesFromFills` | Fills become trades with wallets; a swap without fills is still one trade from the pool's change. |

## `internal/market/wallets_test.go`

| Function | What it does |
|---|---|
| `TestWallet` | A wallet's trades are added up per token for the last 24 hours, other wallets are left out, and the whale level is 1% of the pool. |
