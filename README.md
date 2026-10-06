# Adalytics

**A Cardano token analytics platform.** Adalytics shows live prices, price charts and trading activity for Cardano native tokens, in the style of CoinGecko. Every number comes straight from the Cardano blockchain through the [Koios API](https://www.koios.rest).

![Token list](docs/screenshots/list.png)

## What it does

Adalytics tracks the 360 tokens that Minswap marks as verified and lists every one whose Minswap pool holds at least ₳10,000 (about 75 tokens on a normal day). Prices are in ADA, with a US dollar estimate next to them.

**Token list**

* Price, 1 hour and 24 hour change, 24 hour volume, fully diluted valuation (FDV) and a small 24 hour price line for every token
* Top Volume, Top Gainers and Top Losers views
* Search by name, ticker or policy ID (press `/` to jump to the search box)
* Refreshes by itself every minute

**Token page**

* Live 24 hour price chart (TradingView Lightweight Charts): green when the price is up, red when it is down, with a crosshair and tooltip on hover
* Price in ADA and USD, 24 hour range, volume, number of trades, buys against sells, holder count, pool liquidity and the time of the last trade
* Info box with the project website, policy ID and fingerprint (with copy buttons), creation date, a Cardanoscan link and the project description
* The 5 most recent trades, each with a link to the transaction on Cardanoscan

It also works on phones.

![Token page](docs/screenshots/token.png)

## How it works

```mermaid
flowchart LR
    K[Koios API<br/>Cardano mainnet] -->|HTTPS, read only| B[Go backend<br/>keeps everything in memory]
    B -->|JSON API| F[Vue 3 frontend]
    F --> U[Browser]
```

Adalytics is **read only**. It has no database and writes nothing to disk. When the backend starts, it loads what it needs from Koios into memory; when it stops, that memory is simply gone and is rebuilt on the next start.

**Where the prices come from.** Every token on Minswap V2 has a liquidity pool: one transaction output on the blockchain that holds ADA on one side and the token on the other. The pool's datum records both amounts. The price is simply

```
price in ADA = ADA in the pool / tokens in the pool
```

Whenever someone trades, the pool output is spent and a new one is created with the new amounts. Comparing one pool state with the next shows what happened:

* ADA went in and tokens came out: a **buy**
* tokens went in and ADA came out: a **sell**
* the pool's LP supply changed: someone added or removed liquidity, which is not counted as a trade
* changes under ₳1 are rounding dust and are ignored

**US dollar price.** Koios has no USD prices, so Adalytics uses the USDM stablecoin pool on Minswap: if 1 USDM costs ₳3.70, then ₳1 is worth about $0.27.

**What happens at startup** (about 3 to 5 minutes)

1. Find the current Minswap pool of each of the 360 tokens (`asset_utxos`, 30 tokens per request)
2. List every Minswap transaction of the last 24 hours (`address_txs`) and read them in batches of 50 (`tx_info`) to build the price history, volume and trade list
3. Load total supply, logos and project details (`asset_info`)
4. Count holders in the background (`asset_summary`, one token at a time)

**What happens every minute**

1. Ask Koios whether any tracked pool output has been spent (`utxo_info`, 50 pools per request)
2. Only if one was spent, read the new transactions and update that token's price, chart and trades

## Koios usage and laptop load

Adalytics uses the free Koios tier (50,000 requests a day, 100 requests per 10 seconds) and stays well below it.

| | Measured |
|---|---|
| Requests at startup | about 62 (13 to find the pools, about 50 for 24 hours of history) |
| Requests per minute after that | about 3 |
| Memory | about 25 MB |
| CPU | close to 0% between the minute checks |

The backend also keeps itself inside Koios's limits: at most 90 requests per 10 seconds, request bodies under 5,120 bytes, and automatic retries when Koios is busy.

## Running it yourself

You need Go 1.22 or newer, Node.js 20 or newer, and a free Koios API token from [koios.rest](https://www.koios.rest).

**1. Settings.** Copy the example file and put your Koios token in it:

```bash
cp .env.example .env
```

**2. Backend** (starts on http://localhost:8080):

```bash
cd backend
go test ./...
go build -o adalytics . && ./adalytics
```

**3. Frontend** (in a second terminal, opens on http://localhost:5173):

```bash
cd frontend
npm install
npm run dev
```

The list fills in once the backend has finished loading.

## Project structure

```
backend/
  main.go                     starts the tracker and the web server
  internal/config/            reads settings from .env
  internal/koios/             small Koios client: rate limit, retries, request size check
  internal/minswap/           reads a Minswap V2 pool from its datum and works out the price
  internal/market/
    tokens.json               the 360 verified tokens from Minswap's token list
    tracker.go                loads history, follows new swaps, ranks tokens, builds trades
    history.go                turns swaps into chart points, changes and volume
  internal/api/server.go      the JSON API used by the frontend

frontend/src/
  pages/TokenList.vue         the token list
  pages/TokenPage.vue         the token page
  components/PriceChart.vue   the price chart
  components/RecentTrades.vue the recent trades table
  components/Sparkline.vue    the small price lines in the list
  format.js                   number formats (₳, $, %, K/M/B)
```

**API**

| Endpoint | Returns |
|---|---|
| `GET /api/tokens` | every listed token with price, changes, volume and details |
| `GET /api/tokens/{id}/chart?range=24H` | chart points for one token |
| `GET /api/tokens/{id}/trades?limit=5` | the most recent trades for one token |
| `GET /api/tokens/{id}/logo` | the token's logo from the Cardano token registry |

## Current limits

* Prices come from Minswap V2 only. Tokens that mostly trade on other exchanges may show lower volume than sites that combine every exchange, such as DexHunter.
* Charts cover the last 24 hours. 7 day and 30 day charts are planned.
* Ranking is by 24 hour volume, because circulating supply (needed for market cap) is not known for every token.
* Minswap fills orders in batches, so two people's orders in the same transaction show up as one trade.

## Built with

Go, Vue 3 (Composition API), Vite, Vue Router, TradingView Lightweight Charts, Koios API, and the Minswap verified token list.

![Phone view](docs/screenshots/phone.png)
