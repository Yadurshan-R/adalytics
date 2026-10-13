# Backend functions

Every function in the Go backend, file by file. Each one also has the same comment above it in the code. See [CODE_GUIDE.md](CODE_GUIDE.md) for how the parts fit together.

## `main.go`

| Function | What it does |
|---|---|
| `main` | Loads the settings, starts the Tracker that reads Cardano data from Koios in the background, and serves the API until Ctrl+C is pressed. Every API call, every Koios answer (with LOG_LEVEL=debug) and a summary each minute are written to the log. |
| `logKoios` | Logs each Koios answer: endpoint, status, time, and how much was sent and received. Failures and very slow answers always show. |

## `internal/logx/handler.go`

| Function | What it does |
|---|---|
| `New` | Returns a logger that prints lines like `13:20:04 INFO  api  message  key=value`. |
| `ParseLevel` | Turns LOG_LEVEL (debug, info, warn, error) into a level. |
| `StdLogger` | Gives the Tracker, which uses the older log.Logger, a logger that writes through the same format. |
| `Handle` | Writes one log record as one line: time, level, part of the backend, message, details. |
| `Size` / `Count` / `Duration` | Show bytes as 18KB, numbers as 50,000, and times as 4ms or 1.2s. |

## `internal/config/config.go`

| Function | What it does |
|---|---|
| `Load` | Reads .env (from the current folder, or the parent folder) and then the environment. Values already set in the environment win over the file. |
| `getenv` | Returns the environment variable key, or fallback when it is empty or not set. |
| `loadDotEnv` | Reads the first file that exists. Lines look like KEY=VALUE; blank lines and lines starting with # are skipped. |

## `internal/koios/client.go`

| Function | What it does |
|---|---|
| `New` | Creates a Client for the Koios API at baseURL, using the API token. |
| `Requests` | The number of requests sent to Koios since start, retries included. |
| `BytesReceived` | How much data Koios has sent back since start. |
| `report` | Tells OnResponse about one request and its answer, for the log. |
| `do` | Sends a request and decodes the JSON answer into out. It waits for the rate limiter and retries when Koios is busy (429), slow or briefly down (5xx). |

## `internal/koios/endpoints.go`

| Function | What it does |
|---|---|
| `Tip` | Returns the newest block Koios knows about: its height and time. |
| `UTxOInfo` | Looks up UTxOs by reference ("tx_hash#index") and says if each is spent. |
| `PoolUTxOs` | Returns the unspent outputs at address that hold any of the given assets. Each Minswap pool holds its own LP token, so asking for the LP tokens at the pool address returns exactly those pools, traded recently or not. Requests are split so each body stays under Koios's size limit. |
| `AddressTxsPage` | Returns one page of the transactions that touched the addresses from a block height on, newest first. |
| `AddressTxs` | Lists every transaction that touched the addresses from a block height on (newest first). Koios returns at most 1000 rows per page, so it pages through; progress, if set, is called after each page. |
| `AssetInfos` | Returns supply and registry details for several tokens in one request. |
| `AssetSummary` | Asks Koios how many wallets hold a token. It uses the slow HTTP client, because Koios can take minutes to count big tokens. |
| `TxInfo` | Returns transactions with their outputs, assets and inline datums. |

## `internal/koios/retry.go`

| Function | What it does |
|---|---|
| `Error` | Returns the message, so retryAfter can be used as an error. |
| `parseRetryAfter` | Reads the Retry-After header (in seconds). Without one it waits 60 seconds, about how long Koios blocks after too many requests. |
| `snippet` | Shortens an error response to 200 characters for the log. |
| `sleep` | Waits for d, or returns early with an error when ctx is cancelled. |
| `wait` | Blocks until another request is allowed: it forgets requests older than the window, and if max are still in it, sleeps until the oldest leaves. |

## `internal/koios/types.go`

| Function | What it does |
|---|---|
| `IsADA` | Reports whether the asset is ADA itself rather than a native token. |
| `Ref` | The UTxO reference "tx_hash#index". |
| `UTxO` | Converts a transaction output into the same shape as a UTxO lookup. |

## `internal/minswap/pool.go`

| Function | What it does |
|---|---|
| `Parse` | Turns one Koios UTxO into a Pool. |
| `IsADAPair` | Reports whether this is the ADA / token pool. |
| `PriceADA` | The price of one whole token in ADA: (reserve A / 10^6) / (reserve B / 10^decimals). |
| `ReserveADA` | Returns reserve A in ADA. |
| `ReserveToken` | Returns reserve B in whole tokens (after the token's decimals). |
| `asset` | Reads an asset (policy ID and name) from a pool datum field. |
| `integer` | Reads a whole number from a pool datum field. Reserves can be larger than 64 bits, so it returns a big.Int. |
| `pow10` | Returns 10 to the power n, used to apply a token's decimals. |
| `scaled` | Turns an amount in the smallest unit into whole units, e.g. 1,500,000 lovelace with 6 decimals is 1.5 ADA. |

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
| `Chart` | Returns one token's chart for a range: evenly spaced line points, or candles when candles is true. |
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

## `internal/api/server.go`

| Function | What it does |
|---|---|
| `New` | Builds the HTTP handler with every API route. Each route only reads from the Tracker's memory, so it answers instantly and never calls Koios. |
| `writeJSON` | Sends v as JSON with the given HTTP status, and tells the browser not to cache it, so prices are always fresh. |
| `withCORS` | Lets the Vue dev server (another port) read the API. Read-only: GET only. |

## `internal/api/logging.go`

| Function | What it does |
|---|---|
| `withLogging` | Logs one line per API call (status, time, size, visitor, request ID), sends the ID back in X-Request-ID, and catches a crashing request so the server keeps running. |
| `levelOf` | Errors and slow answers are louder; logos and health checks only show with LOG_LEVEL=debug. |
| `visitorOf` | The visitor's IP; behind Nginx and ngrok it is read from X-Forwarded-For, which is only trusted for local requests. |
| `errorOf` | The reason in an error answer, for the log. |

## `internal/api/stats.go`

| Function | What it does |
|---|---|
| `add` / `take` | Count each answered request, and hand over the counts each minute. |
| `ReportEveryMinute` | Logs one summary line a minute: website traffic and Koios usage today against the free tier's 50,000 a day. |


# Backend tests

Run them with `go test ./...` from `backend/`. They use a fake Koios server, so they need no internet and no API token.

## `internal/minswap/pool_test.go`

| Function | What it does |
|---|---|
| `loadFixture` | testdata/snek_pools.json is real Koios output (SNEK/MIN pool and SNEK/ADA pool). |
| `TestParseSNEKADAPool` | Reads a real SNEK/ADA pool and checks its assets, reserves and price. |
| `TestPriceUsesTokenDecimals` | Checks that a token's decimals are applied when working out the price. |

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

## `internal/api/logging_test.go`

| Function | What it does |
|---|---|
| `TestLogsEveryRequest` | Checks the log line, the visitor behind Nginx and the request ID. |
| `TestLogsErrorReason` | Checks that a 404 is a warning with its reason. |
| `TestCrashIsCaught` | Checks that a crashing request gets a 500 and is logged, and the server keeps running. |
| `TestVisitorIsOnlyTrustedLocally` | Checks that X-Forwarded-For is ignored unless the request is local. |
| `TestSummary` | Checks the minute summary line. |

## `internal/market/tracker_test.go`

| Function | What it does |
|---|---|
| `TestLoadsHistoryThenFollowsSwaps` | Checks the first 24 hour load (price, volume, change, chart), then that a new swap is picked up a minute later. |
| `TestQuietTokenIsListed` | Checks that a token with a pool but no trades today is still listed, with a flat price line. |
| `TestTokenWithoutPoolIsLeftOut` | Checks that a token with no Minswap pool is not listed. |
| `TestFindsAllPoolsInSmallRequests` | Checks that the pool search is split into requests small enough for the Koios free tier. |
