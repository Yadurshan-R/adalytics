# Market functions

Every function in `internal/market`, the package that reads the Minswap pools and keeps prices, history and trades in memory. Each one also has the same comment above it in the code. The rest of the backend is in [BACKEND_FUNCTIONS.md](BACKEND_FUNCTIONS.md). The tests are in [MARKET_TESTS.md](MARKET_TESTS.md).

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
| `backfill` | Loads the history older than the first 24 hours, newest first, one page of 1,000 transactions at a time, so 7 days are ready long before 30. The next page's list is fetched while this page's transactions are read, so no time is lost between pages. A page that fails is tried again a minute later. |
| `fetchPage` / `historyPage` | Asks Koios for one page of the list of Minswap transactions in the background and hands it over when it arrives. |
| `finishHistory` | Marks the whole history span as loaded, so every chart range is ready. |

## `internal/market/candles.go`

| Function | What it does |
|---|---|
| `CandleSizeByKey` | Finds a candle size by its key (5m, 15m, 30m, 1h, 2h, 4h, 12h, 1D, 1W, 1M), ignoring case. 5 minute candles go back 2 days, 15 and 30 minutes 7 days, the others 30 days. |
| `Candles` | Returns one token's candles at the given size, oldest first. While older history is still loading it returns what is loaded so far, so the chart grows instead of waiting for all 30 days. |

## `internal/market/calendar.go`

| Function | What it does |
|---|---|
| `weekStart` / `monthStart` | The Monday 00:00 UTC of t's week, and the 1st 00:00 UTC of its month. |
| `candleStarts` | Where each candle starts between from and to: the first whole step at or after from (a multiple of the size, or the first Monday or 1st for 1W and 1M), so a first candle that began before the history, and would miss part of its time, is left out. The last candle is the running one. |

## `internal/market/details.go`

| Function | What it does |
|---|---|
| `loadInfo` | Reads total supply, ticker and logo for every tracked token with asset_info, and each minting policy with script_info, 25 tokens per request. If it fails, prices still work and it is tried again next minute. |
| `safeURL` | Keeps a registry website only if it is a plain http(s) link, so a token's metadata cannot put another kind of link on the page. |

## `internal/market/holders.go`

| Function | What it does |
|---|---|
| `holdersLoop` | Waits for the first load, counts every listed token's holders, then counts them all again every HoldersEvery (6 hours), until ctx ends. Tokens whose count failed are tried again in between, one at a time (HoldersRetryWorkers), after growing waits (HoldersRetryWaits: a minute twice, then 5 and 15 minutes, then every hour). Each wait starts when the previous round is done; the first starts when the full count of every token is done. How long Koios takes varies with how busy it is (HOSKY: 26 seconds once, timed out later the same day), so a heavy token gets a few quick chances, then the waits grow because each try costs Koios minutes of work. One at a time because heavy counts running together share Koios's effort and time out together. |
| `loadHolders` | Counts the holders of every token in the current list, busiest first, and returns the tokens whose count failed. |
| `countHolders` | Counts the holders of the given tokens, a given number at a time (HoldersWorkers, 5, in a full count; HoldersRetryWorkers, 1, for retries): wallets with a stake key plus addresses without one (contracts included, such as the token's own pool), from one slow asset_summary request per token, noting when (holders_at in the API, shown as "counted 2 hours ago"). Each token keeps how many counts in a row failed; from HoldersGiveUp (5) on, its holders show as unknown ("—") until a count works. Returns the tokens whose count failed. |
| `stateByID` | Finds a token by its page id and returns it with its ticker. |

## `internal/market/policy.go`

| Function | What it does |
|---|---|
| `canMint` | Says whether a native script could still allow a mint now or later: a key can always sign, "before" closes at its slot, "after" opens at its slot; all, any and atLeast combine their parts. |
| `state` | The policy as the page shows it: "locked" (no more can be minted), "open", "script" (Plutus) or "unknown". Worked out at each request, so a time lock that passes turns into "locked" without reading again. |
| `readPolicies` | Reads the minting script of each policy with script_info, 25 per request. |

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
| `tradeVolume` | The ADA a pool state traded: the sum of its trades, so every order of a Minswap batch counts on its own (gross volume, as Minswap and DexScreener count it). The pool's own change nets buys against sells: a ₳1,000 buy and a ₳800 sell in one batch move the pool by ₳200 but trade ₳1,800. |
| `volumeSince` | Adds up the ADA traded from time t (unix seconds) until now. |
| `candles` / `candlesAt` | Groups swaps into candlesticks; `candlesAt` takes the candle starts (from `candleStarts`), so weeks and months work the same as fixed sizes. A slot with no swaps is a flat candle at the previous close, and each candle opens at the previous close so the candles join up. |

## `internal/market/periods.go`

| Function | What it does |
|---|---|
| `Periods` | One token's trading over the last 5 minutes, hour, 4 hours and 24 hours: price change, volume, buys and sells and the ADA on each side, from the trades already in memory. |
| `periodsOf` | Works the periods out for one token; the token list (Quotes) and the activity box both use it. |

## `internal/market/pools.go`

| Function | What it does |
|---|---|
| `readPools` | Reads transactions with tx_info (50 per request, several requests at a time: KOIOS_PARALLEL, default 8) and returns, for each token, the pool state after each of its transactions, oldest first. |
| `merge` | Adds pool states to a token's history (skipping ones it already has), keeps the history in chain order and rebuilds prices, volume and trades. It returns how many states were new. Call with t.mu held. |
| `rebuild` | Works out a token's price points, volume and trades from its pool history, drops history older than the kept span (keeping the last state before it, so the price at the start is known) and updates the quote from the newest pool state. A trade is the change from one pool state to the next; the oldest state loaded (just after a start only the newest hour is read) has none before it, but its own transaction's fills still count as trades. Each state's volume is its trades added up (`tradeVolume`), so the list, the activity box and the candles show the same volume. Call with t.mu held. |
| `replaceCurrent` | Swaps a pool state found at startup (as the pool is now) for the same pool output read from its transaction, which has the wallets and its place in the block. |
| `tradesFrom` | Turns a swap into trades: one per wallet when the transaction was split into fills, otherwise one trade without a wallet from the change in the pool's reserves. Each trade keeps its order's number (`Fill`), because one wallet can have two orders in one batch. |
| `tradeFrom` | Makes one trade from ADA into the pool and tokens to the trader: both positive is a buy, both negative a sell. Dust below MinTradeADA is left out. |

## `internal/market/fills.go`

Minswap does not swap each order on its own: a batcher swaps several people's orders against the pool in one transaction. These functions split such a transaction into one trade per order. The trader is the wallet the payout goes to: an order's own address often has no stake key, or Minswap's. A routed swap (e.g. USDA → ADA → NIGHT) passes through more than one pool; with one order, the token's own pool change is that wallet's trade.

| Function | What it does |
|---|---|
| `fillsFrom` | Finds the token's pool by its LP asset, then sorts the transaction into pools, the batcher (its coins go in and come out at the same address), the orders and one payout per order, each order paired with its own payout (pairOrders). An order the batcher killed (refunded, no swap) is left out. Each pair's payout minus order, less its share of the batcher fee, is what went into or out of the pool. Returns nil, so the swap shows as one trade without a wallet, when anything does not add up: as many payouts as orders, one order only when the swap is routed through other pools, every token accounted for, a fee of ₳0 to ₳10 per order, and with several orders each one near the pool's price. |
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

## `internal/market/pairs.go`

| Function | What it does |
|---|---|
| `ordersAndPayouts` | Sorts a transaction's inputs and outputs like Minswap's pool validator: the batcher's coins (in and out at the same address), the pools, the orders (sorted by UTxO, the ledger's order) and one payout per order (by place in the transaction). |
| `pairOrders` | Puts each order next to its own payout. When every order names a different wallet in its datum, that pairs them (its own wallet first, then its refund wallet for an order the batcher killed). When one wallet placed several orders, the batcher's own list (minswap.BatchOrder) says which output pays which order, and each payout must still go to its order's wallet. Koios lists outputs in no fixed order, so positions alone cannot pair them. Returns nil, so the swap is not split, when the pairing is not certain. |
| `sortPairs` | Orders the pairs by the payout's place in the transaction: the order the batcher swapped them in. |
| `needsBatchOrder` | Reports whether a transaction needs the batcher's list: one wallet placed several orders in it. |

## `internal/market/batches.go`

| Function | What it does |
|---|---|
| `addBatchOrders` | Fetches the batcher's list (TxContracts) only for the transactions that need it, 50 per request: a few in each batch of 50, so one small extra request now and then instead of downloading every transaction's scripts. If it fails, those batches show as one trade without a wallet. |
