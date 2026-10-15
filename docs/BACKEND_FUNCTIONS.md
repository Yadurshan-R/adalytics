# Backend functions

Every function in the Go backend except the market package (see [MARKET_FUNCTIONS.md](MARKET_FUNCTIONS.md)), file by file. Each one also has the same comment above it in the code. See [CODE_GUIDE.md](CODE_GUIDE.md) for how the parts fit together.

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
| `PoolUTxOs` | Returns the unspent outputs at address that hold any of the given assets. Each Minswap pool holds its own LP token, so asking for the LP tokens at the pool address returns exactly those pools, traded recently or not. Requests are split so each body stays under Koios's size limit, and run three at a time. |
| `AddressTxsPage` | Returns one page of the transactions that touched the addresses from a block height on, newest first. |
| `AddressTxs` | Lists every transaction that touched the addresses from a block height on (newest first). Koios returns at most 1000 rows per page, so it pages through; progress, if set, is called after each page. |
| `AssetInfos` | Returns supply and registry details for several tokens in one request. |
| `AssetSummary` | Asks Koios how many wallets hold a token. It uses the slow HTTP client, because Koios can take minutes to count big tokens. |
| `TxInfo` | Returns transactions with their inputs and outputs, assets and inline datums. The inputs show which wallets' orders a batch swapped. |

## `internal/koios/parallel.go`

| Function | What it does |
|---|---|
| `Parallel` | Runs a list of jobs a few at a time (Workers = 3) and returns the first error; jobs not started yet are skipped after one fails. Same number of Koios requests, about three times as fast. |

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

## `internal/api/server.go`

| Function | What it does |
|---|---|
| `New` | Builds the HTTP handler with every API route (tokens, chart, trades, wallets, activity, logo, health). Each route only reads from the Tracker's memory, so it answers instantly and never calls Koios. |
| `candleChart` | Answers a candlestick chart at one candle size (`?type=candles&size=1h`), with all the history loaded so far. |
| `validWallet` | Checks a wallet looks like a mainnet stake address (stake1…) or payment address (addr1…): lowercase letters and digits only. |
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

## `internal/api/gzip.go`

| Function | What it does |
|---|---|
| `withGzip` | Compresses JSON answers for browsers that accept gzip, so pages load faster. |
| `WriteHeader` | Decides whether to compress: only successful JSON answers; logos and errors are sent as they are. |
| `Write` / `close` | Send the bytes (compressed when chosen) and finish the answer. |


# Backend tests

Run them with `go test ./...` from `backend/`. They use a fake Koios server, so they need no internet and no API token.

## `internal/minswap/pool_test.go`

| Function | What it does |
|---|---|
| `loadFixture` | testdata/snek_pools.json is real Koios output (SNEK/MIN pool and SNEK/ADA pool). |
| `TestParseSNEKADAPool` | Reads a real SNEK/ADA pool and checks its assets, reserves and price. |
| `TestPriceUsesTokenDecimals` | Checks that a token's decimals are applied when working out the price. |

## `internal/api/logging_test.go`

| Function | What it does |
|---|---|
| `TestLogsEveryRequest` | Checks the log line, the visitor behind Nginx and the request ID. |
| `TestLogsErrorReason` | Checks that a 404 is a warning with its reason. |
| `TestCrashIsCaught` | Checks that a crashing request gets a 500 and is logged, and the server keeps running. |
| `TestVisitorIsOnlyTrustedLocally` | Checks that X-Forwarded-For is ignored unless the request is local. |
| `TestSummary` | Checks the minute summary line. |

## `internal/api/gzip_test.go`

| Function | What it does |
|---|---|
| `TestGzipJSON` | Checks that JSON is compressed for browsers that accept gzip, unpacks to the same answer, and is sent plain to others. |
| `TestGzipSkipsLogosAndErrors` | Checks that logos and error answers are not compressed. |

## `internal/api/wallet_test.go`

| Function | What it does |
|---|---|
| `TestValidWallet` | Checks only real-looking stake and payment addresses pass the wallet route's check. |
