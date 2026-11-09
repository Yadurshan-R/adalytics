# Market tests

Every test in `internal/market`, file by file. Run them with `go test ./internal/market/` from `backend/`. They use a fake Koios server, so they need no internet and no API token. The functions they check are in [MARKET_FUNCTIONS.md](MARKET_FUNCTIONS.md).

## `internal/market/holders_test.go`

| Function | What it does |
|---|---|
| `holdersFake` | A pretend Koios that only answers asset_summary, slowly, noting how many requests run at once, and timing out (504) the first few counts of the chosen tokens. |
| `TestCountHolders` | A full count runs HoldersWorkers at a time; the failed tokens are handed back, and the retry counts them one at a time (HoldersRetryWorkers) and fills them in. |
| `TestHoldersUnknownAfterFailures` | A token Koios keeps timing out on is asked once per try (no retries at once), shows "…" until it failed HoldersGiveUp times in a row, then shows as unknown; one count that works brings its number back, with the time it was counted. |

## `internal/market/calendar_test.go`

| Function | What it does |
|---|---|
| `TestCandleStartsFixed` | Checks that 1h and 1D also leave out a first candle that began before the history (from 10:30: 11:00, and the next day), and keep it when the history starts on a whole step. |
| `TestWeekAndMonthStart` / `TestCandleStartsCalendar` | Checks that weeks start on Monday and months on the 1st (UTC), and that a first week or month missing days is left out. |
| `TestWeekCandleCountsItsWholeWeek` | Checks that a week candle holds every swap of its week, with the right open, high, low, close and volume. |

## `internal/market/candles_test.go`

| Function | What it does |
|---|---|
| `TestCandleSizes` | Checks that each candle size covers all the history loaded so far: 24 hours at first, then 7 days once the background load is done, without a first candle that began before the history (4h, 12h, 1D), and that a wrong size or token is an error. |

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

## `internal/market/backfill_test.go`

| Function | What it does |
|---|---|
| `TestBackfillSeveralPages` | Loads 2,500 older transactions (three pages of the list): every page is read, every transaction ends up in the history, and the next page's list is asked for while the current page is still being read. |

## `internal/market/rebuild_test.go`

| Function | What it does |
|---|---|
| `minToken` | MIN from tokens.json. |
| `TestFirstStateTrades` | A real routed trade (NIGHT → MIN → ADA, ₳677.89, testdata/min_routed.json): as the only pool change loaded it still counts as a trade in the 1h numbers, and once the state before it is loaded too it is still one trade, not two. |

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

## `internal/market/batches_test.go`

| Function | What it does |
|---|---|
| `multiOrderBatches` | testdata/multi_order_batches.json: 5 real batches (4 Nov 2026) in which one wallet placed several orders, one of them with an order the batcher killed. |
| `poolAsset` | Finds the token and LP asset of the pool a transaction swapped in. |
| `TestFillsSeveralOrdersOneWallet` | Every trade of the 5 batches: the wallet, the ADA and the tokens, in the order they were swapped; the killed order gives no trade. |
| `TestFillsSeveralOrdersWithoutList` | Without the batcher's list such a batch is not split, never guessed. |
| `TestBatchOrder` | Reads the batcher's list from a real batch. |
| `TestAddBatchOrders` | The lists are fetched only for the transactions that need them, in one request, and attached to the right ones. |

## `internal/market/pairs_test.go`

| Function | What it does |
|---|---|
| `nightBatch` | testdata/night_batch.json is a real mainnet batch (fa72d267…): 7 NIGHT sell orders from 6 wallets, with outputs that Koios lists out of order, and the batcher's list. |
| `TestFillsRealBatch` | Every order of the real batch goes to the wallet its datum names, with the exact amounts Cardanoscan shows, and together they add up to the pool's change. |
| `TestFillsRealBatchUnsure` | The real batch changed so the pairing is uncertain (an order for another pool, a payout no order names, an order without a datum) is not split. |
| `TestReadOrder` | Reads the receiver and the pool from a real order's datum. |

## `internal/market/fills_test.go`

| Function | What it does |
|---|---|
| `TestFillsRealSwaps` | Three real SNEK swaps from mainnet: a ₳487 buy, then one wallet selling twice. Each gives the right wallet, ADA and tokens. |
| `TestFillsOrderWithoutOwnerKey` | Three more real SNEK swaps whose orders carry no stake key, or the pool's: the trader is the wallet the payout goes to. |
| `TestFillsRouted` | Two real NIGHT swaps routed through a stablecoin pool (a buy and a sell): NIGHT's own pool change is the wallet's trade. |
| `TestFillsBatch` | Three wallets in one batch (two buys and a sell) become three trades that add up to the pool's change. |
| `TestFillsNotSplit` | Transactions that cannot be split for certain give no fills: no inputs, tokens missing, a fee too big, an order without a payout, two orders routed through two pools, payouts paired the wrong way round. |
| `TestTradesFromFills` | Fills become trades with wallets; a swap without fills is still one trade from the pool's change; two orders of one wallet in one batch get Fill 0 and 1. |

## `internal/market/wallets_test.go`

| Function | What it does |
|---|---|
| `TestWallet` | A wallet's trades are added up per token for the last 24 hours, other wallets are left out, and the whale level is 1% of the pool. |
