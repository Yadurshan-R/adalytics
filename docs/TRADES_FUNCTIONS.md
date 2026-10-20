# Large Trades functions

Every function behind the Large Trades page (`/trades`), backend and frontend. Each one also has the same comment above it in the code. See [CODE_GUIDE.md](CODE_GUIDE.md) for how the parts fit together.

## What the page is

One table of the big trades across every token, up to 30 days back, laid out like Birdeye's Find Trades. Who uses it: traders who want to see where big money is going without opening each token. A trade is on the page when it is at least 1% of its pool's ADA (it moves the price about 2%) or ₳10,000 and more. The filters are Token (a search with suggestions), Time (last 1 hour to 30 days), Size (all, fish, dolphin, whale) and Wallet. Everything is read from the trades the backend already holds for the chart history, so the page costs no Koios requests.

| Size | Meaning |
|---|---|
| All sizes | Every trade on the page (the default). |
| 🐟 Fish | Under ₳10K, but at least 1% of the pool. |
| 🐬 Dolphin | ₳10K to ₳50K. |
| 🐋 Whale | ₳50K or more. |

Picking a size shows only that size.

## `internal/market/whales.go`

| Function | What it does |
|---|---|
| `WhaleTrade` | One trade on the page: the trade, its token (id, ticker, logo or not) and its share of the pool in %. |
| `WhaleFilter` | What the page asks for: hours, the size as an ADA range (smallest, and below what; 0 means no top), a token, a wallet, how many trades to skip (Show more) and how many to send. |
| `WhaleHours` | The time choices in hours: 1, 4, 24, 168 (7 days), 720 (30 days). Anything else is 24. |
| `PageWhales` | 200: how many trades one page, and each Show more, gets. |
| `MaxWhales` | 1000: the most one answer sends, when a page refreshes all the rows it has opened with Show more. |
| `BigTradeADA` | ₳10,000: a trade this big is on the page even when it is under 1% of a deep pool. |
| `WhaleTrades` | One page of the big trades that match the filter, newest first, and whether there are more. A wallet matches whatever its letter case. |
| `WhaleTokens` | Every listed token with its ticker, name and whether it has a logo, the biggest pool first, for the token search. |

## `internal/market/whales_test.go`

| Function | What it does |
|---|---|
| `TestWhaleTrades` | Small trades are left out; ₳10K+ trades under 1% of the pool are in; each filter (24 hours, 7 days, ₳25K and up, fish, ₳20K to ₳50K, wallet in capitals, another token, offset, limit) gives the right trades in the right order; the pool shares are right. |

## `internal/api/server.go`

| Route | What it returns |
|---|---|
| `GET /api/whales?hours=24&min=10000&max=50000&token=snek&wallet=stake1…&offset=200&limit=200` | One page of trades (`trades`), whether there are more (`more`), the token list (`tokens`) and how much of 7D and 30D is loaded (`ranges`). |

## `src/pages/LargeTrades.vue`

| Function | What it does |
|---|---|
| `ask` | Builds the backend address for one page of trades with the filters picked. |
| `load` | Asks for the newest trades again (as many as are shown, so the pages opened with Show more stay; one page after a filter change), or the next 200 for Show more. Runs again every minute. Only the answer to the latest question is used, so a slow answer for old filters never shows under new ones. |
| `addNew` | Adds the next page without repeating a trade: trades arriving between two pages move the rest down, so the next page can start with a few already shown. |
| (watch) | A filter change loads again and goes into the page address, so a filtered view can be shared; typing a wallet waits for a short pause. |
| `changed` / `reset` | Whether any filter is off its default, and Reset all. |
| `pickWallet` | A click on a trader shows only that wallet's trades; the same wallet again shows all. |
| `span` | "24 hours", "7 days"… for the empty message. |

## `src/components/TradeFilters.vue`

| Function | What it does |
|---|---|
| (template) | The filter bar: the token search, the Time and Size lists and the wallet box side by side (two by two on a phone), and Reset all beside the Wallet title. Each size in the list shows its range, e.g. "🐬 Dolphin · ₳10K–₳50K". |
| `TIMES` | The Time list: last 1 hour, 4 hours, 24 hours, 7 days, 30 days. |
| `loading` | How much of 7 or 30 days is loaded; until it is all in, that choice is grey with "Loading 35%", like the chart's list. |

## `src/components/TokenPicker.vue`

| Function | What it does |
|---|---|
| (template) | The Token box, like Birdeye's "Select a token": closed, it shows the token picked (logo and ticker) or "All tokens"; open, a search box with suggestions under it. |
| `picked` | The token picked, for the closed box. |
| `matches` | The suggestions: "All tokens", then up to 8 tokens. Before anything is typed, the biggest pools; after, tickers starting with the text first, then tickers or names containing it. |
| `show` | Opens the search, empty and focused. |
| `choose` | Picks a token (or All tokens) and closes. |
| `onKey` | Up and down arrows move through the suggestions, Enter picks, Esc closes. A click anywhere else closes too. |

## `src/components/TradesTable.vue`

| Function | What it does |
|---|---|
| (template) | Time, Buy or Sell, the value with its size icon, both sides of the swap (the token on top, the ADA under it), % of pool, the trader and a Cardanoscan link. A phone hides the value, % of pool and link, and shows the size icon beside Buy or Sell. |
| `rowKey` | A unique name for each row: transaction, order number (`fill`) and token. Without the order number, two orders of one wallet in one batch looked like one row, and rows were left behind on screen when the filters changed. |
| `showSize` (prop) | The fish, dolphin or whale icon on each row, shown only with All sizes: with one size picked, every row is that size. |
| `open` | Goes to the token's page. |
| `ada` | ₳62,977: the value in whole ADA. |
| `sign` | + or −, from the trader's side: a buy gets tokens (+) for ADA (−), a sell the other way round. |

## `src/tradeSizes.js`

| Function | What it does |
|---|---|
| `SIZES` | Fish, dolphin and whale: their icon, name, range text, ADA range (min, and max or 0 for no top) and the hover text for the table. |
| `sizeOf` | The size of one trade, for its icon in the table. |
