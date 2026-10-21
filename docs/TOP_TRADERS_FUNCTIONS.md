# Top Traders functions

Every function behind the Top Traders tab on a token page, backend and frontend. Each one also has the same comment above it in the code. See [CODE_GUIDE.md](CODE_GUIDE.md) for how the parts fit together.

## What the tab is

Under the chart, the token page has two tabs, like Birdeye's and DexScreener's: Trades (the latest trades) and Top Traders. Who uses it: someone looking at a token who wants to know who is moving it, for example whether one wallet is buying a lot or the big wallets are selling. Top Traders lists the 50 wallets that traded the token most in the last 24 hours, 7 days or 30 days, with what each bought and sold. Everything is read from the trades the backend already holds for the chart history, so the tab costs no Koios requests.

| Column | Meaning |
|---|---|
| Bought / Sold | ADA spent buying and ADA received selling, with the tokens under it. |
| Net ₳ | Bought minus sold, the same Net as the wallet card: above zero (green), the wallet bought more than it sold. It is not profit: it does not know what the tokens are worth now. |
| Trades | Buys / sells. |
| Last trade | Its latest trade in this token. |

Rank by Volume (bought plus sold) or Net (biggest net buyers first). Trades whose wallet is not known are left out.

## `internal/market/traders.go`

| Function | What it does |
|---|---|
| `TopTrader` | What one wallet did in one token over a time range: buys, sells, ADA and tokens each way, latest trade. |
| `add` | Counts one more trade of the wallet. |
| `TraderHours` | 24, 168 and 720 hours. Anything else becomes 24. |
| `MaxTopTraders` / `MaxWalletTrades` | 50 wallets in the tab; 200 of one wallet's trades in the Trades tab. |
| `inRange` | Goes through a token's trades in the time range, newest first. |
| `TopTraders` | The top 50 wallets of a token, by volume or net buying. Ties go by wallet, so the order is the same every time. |
| `TraderTrades` | One wallet's trades in a token over the time range (the latest 200) and its totals over all of them, so the wallet card matches its Top Traders row even when the list is cut. |

## `internal/market/traders_test.go`

| Function | What it does |
|---|---|
| `TestTopTraders` | Trades are added up per wallet (also two orders of one wallet in one batch), ranked by volume or net buying, cut to 24 hours or 7 days; unknown wallets and tokens, a wrong time range and the 50-wallet limit are handled. |
| `TestTraderTrades` | A busy token where the top wallet is not in the 50 latest trades: its list and totals still match its Top Traders row over 24 hours and 7 days, also when it has more than 200 trades. |

## `internal/api/server.go`

| Route | What it returns |
|---|---|
| `GET /api/tokens/{id}/traders?hours=168&sort=net` | The top wallets (`traders`) and how much of 7D and 30D is loaded (`ranges`). |
| `GET /api/tokens/{id}/trades?wallet=stake1…&hours=168` | One wallet's trades in the token (`trades`) and its totals (`trader`). A wallet that does not look like a Cardano address gets an error. |

## `src/components/TradesSection.vue`

| Function | What it does |
|---|---|
| (template) | The Trades and Top Traders tabs. Trades keeps its filters while Top Traders is open, and Top Traders keeps its time range and sort (KeepAlive) while Trades is open. A new token starts on Trades. |
| `showWallet` | Opens the Trades tab with one wallet's trades, over the time range picked in Top Traders. |

## `src/components/TopTraders.vue`

| Function | What it does |
|---|---|
| `load` | Asks the backend for the top traders with the time range and sort picked; an answer that comes after a newer question is dropped. Runs every minute, only while the tab is open. |
| `waiting` | How much of 7 or 30 days is loaded; until it is all in, that choice is grey with "Loading 35%", like the chart's list. |
| `net` / `signed` / `ada` | Net buying, written like +₳17,034 or −₳38,581; ₳ amounts in whole ADA, or a dash for none. |
| (template) | Rank, wallet, Bought, Sold, Net ₳, Trades and Last trade. Clicking a wallet shows its trades. 10 rows at first, Show more adds 10. The columns fit the space the table has: Trades and Last trade go first, then the token amounts. |
