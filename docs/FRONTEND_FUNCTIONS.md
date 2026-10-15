# Frontend functions

Every function in the Vue website except the chart (see [CHART_FUNCTIONS.md](CHART_FUNCTIONS.md)), file by file. Each one also has the same comment above it in the code. See [CODE_GUIDE.md](CODE_GUIDE.md) for how the parts fit together.

## `src/useTokens.js`

| Function | What it does |
|---|---|
| `load` | Fetches /api/tokens once and stores the tokens, history and any error. |
| `schedule` | Plans the next refresh: fast during the first load, slower after it. |
| `useTokens` | Loads the list when the first page using it opens and refreshes it while any page using it is open: every 3 seconds during the first load, every 15 seconds while 7D / 30D load, and every minute after that. |

## `src/format.js`

| Function | What it does |
|---|---|
| `formatPrice` | ₳0.1699, ₳0.002681, and for tiny prices CoinGecko-style ₳0.0₇3391 (the small 7 means seven zeros after the decimal point). |
| `formatUSD` | $0.0503, $1.01, $0.0₇8970: a price in US dollars, same style as formatPrice. |
| `priceWith` | Formats a price with ₳ or $, using 4 significant digits and the small-number style for tiny prices. formatPrice and formatUSD both use it. |
| `fullPrice` | Full price for tooltips, e.g. 0.00000003391 |
| `formatADA` | ₳1,234,567 (whole ADA), like CoinGecko's volume and market cap columns. |
| `formatPct` | Shows a change as a percentage with one decimal, without its sign: the arrow and colour show the direction. |
| `formatTime` | Shows the clock time, e.g. 14:05:09. |
| `formatAgo` | "just now", "12 min ago", "3 hours ago", "4 days ago" |
| `formatCompact` | ₳3.71T, ₳3.35M, ₳309.9K, ₳812, like CoinGecko's stats column. |
| `formatDateTime` | "Oct 5, 2026, 13:20:00 GMT+5:30" for chart tooltips. |
| `formatAmount` | 76.72B, 21.0M, 940K: token amounts without the ₳ sign. |
| `short` | Shortens a long id for the info box: the first 8 and last 6 characters. |
| `host` | Shows only the website's name, e.g. https://www.snek.com/x → snek.com. |
| `createdDate` | Shows a date like Apr 24, 2023. |

## `src/components/Change.vue`

| Function | What it does |
|---|---|
| `dir` | Decides the colour and arrow: up, down, flat (rounds to 0.0%) or none (no value). |

## `src/components/LoadingProgress.vue`

| Function | What it does |
|---|---|
| `n` | Formats a whole number with thousands separators: 12345 → "12,345". |
| `count` | The number next to a step, e.g. "1,200 / 3,400". Steps not started yet, and the details step, show nothing. |

## `src/components/ActivityBox.vue`

| Function | What it does |
|---|---|
| `load` | Asks the backend for the token's activity, and again every minute. |
| `net` | ADA spent buying minus ADA received selling, for the picked period. |
| `buyShare` | The buyers' part of the green and red bar, in percent. |
| `pct` | Shows a change like +8.54% or −2.49%. |

## `src/components/RecentTrades.vue`

| Function | What it does |
|---|---|
| `load` | Asks the backend for this token's latest trades and the whale level, and keeps any error to show. |
| `isWhale` | True for a trade of at least 1% of the pool's ADA; the row is tinted and its amount bold. |
| `matching` | Every kept trade that passes the filter (All, Buys, Sells, Whales) and the wallet picked. 10 show at first; Show more adds 10. |
| `setFilter` / `setWallet` | Pick a filter or a wallet (the same wallet again shows everyone) and start again at 10 rows. |
| `emptyText` | Explains why no trade is shown, e.g. "No whale trades by this wallet in the last 24 hours." |
| `ada` | ₳1,250 for normal amounts, ₳4.35 for small ones. |

## `src/components/TraderCell.vue`

| Function | What it does |
|---|---|
| `short` | The wallet's short stake address, like stake1u82z…r2dn; a phone shows only the end. Clicking it, or its filter icon, shows only that wallet's trades. |

## `src/components/WalletCard.vue`

| Function | What it does |
|---|---|
| `load` | Asks the backend which other tokens the wallet traded in the last 24 hours. |
| `stats` | Adds up the wallet's buys and sells in this token, and its net (bought minus sold). |
| `signed` / `tone` | Write a net amount like +₳1.4K or −₳640, green or red. |
| `copy` | Puts the address on the clipboard and shows "Copied" for 1.5 seconds. |
| (Cardanoscan link) | Opens the stake key page for a stake1… wallet, or the address page for a wallet without a stake key (addr1…). |

## `src/components/Sparkline.vue`

| Function | What it does |
|---|---|
| `flat` | No trades in the period: every value is the same price. |
| `color` | Green if the price ended higher than it started, red if lower, grey if flat. |
| `points` | Turns the prices into SVG line coordinates that fill the width and height. |

## `src/pages/TokenList.vue`

| Function | What it does |
|---|---|
| `weekRange` | The backend's loading status of the 7 day history. |
| `weekReady` | True once 7 days of history are loaded. |
| `week` | True when the list shows 7 day numbers. |
| `setPeriod` | Picks 24h or 7d; 7d only works once it is loaded. |
| `changeOf` | A token's price change for the chosen period. |
| `periodText` | The period in words, for messages like "No token is up in the last 24 hours". |
| `moved` | A change that rounds to 0.0% is not a gain or a loss. |
| `active` | True if the token traded in the chosen period. |
| `place` | Place in the Top Volume list (the backend's order without the quiet tokens); quiet tokens found by search have no place. |
| `ranks` | Numbers the rows by their Top Volume place when searching or on Top Volume; gainers and losers are numbered 1, 2, 3 instead. |
| `shown` | The list on screen: the chosen view (volume, gainers or losers), filtered by the search box. |
| `emptyText` | Explains why the list is empty. |

## `src/components/ListToolbar.vue`

| Function | What it does |
|---|---|
| `onKey` | Makes "/" jump to the search box from anywhere on the page. |

## `src/components/TokenTable.vue`

| Function | What it does |
|---|---|
| `open` | Goes to a token's page. |
| `lineOf` | A token's small price line for the chosen period. |
| `lineLabel` | The title of the small price line column. |

## `src/pages/TokenPage.vue`

| Function | What it does |
|---|---|
| `token` | The token this page is about, found by the id in the address. |
| `back` | Goes to the previous page if it was the list, otherwise opens the list. |
| `copy` | Puts text (a policy ID or fingerprint) on the clipboard and shows "Copied" next to it for 1.5 seconds. |
| `span` | The low and high for the range bar, sent by the chart (ChartPanel). |
| `rangePos` | Places the current price on the low/high bar, from 0% to 100%. |
