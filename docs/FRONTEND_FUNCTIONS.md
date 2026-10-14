# Frontend functions

Every function in the Vue website, file by file. Each one also has the same comment above it in the code. See [CODE_GUIDE.md](CODE_GUIDE.md) for how the parts fit together.

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

## `src/components/PriceChart.vue`

| Function | What it does |
|---|---|
| `rgba` | Turns a colour like #16c784 into rgba() with transparency a (0 to 1). |
| `isCandles` | Reports whether the chart shows candles instead of a line. |
| `paint` | Line mode: colour and bold/faded split for the hovered point (-1 = none). |
| `load` | Puts the current points or candles into the chart. The view is set only when the token, range, size or type changes, so a zoom or scroll survives the minute refresh. Candles open on the latest ones, about 14px wide each. |
| `onKey` | Esc stops the drawing being placed and unselects; Delete or Backspace removes the selected drawing. |

## `src/components/ChartLegend.vue`

| Function | What it does |
|---|---|
| `change` | The move in ADA and percent: open to close for a candle, first price to this price for the line. |

## `src/components/ChartPanel.vue`

| Function | What it does |
|---|---|
| `load` | Fetches the chart: the line for the picked range, or the candles at the picked size. Runs again every minute. |
| `setType` / `setRange` / `setSize` | Switch between line and candles, the line's range, or the candle size. |
| `waiting` | True while the picked line range is still loading; the loading percentage shows instead. Candles never wait: they show the history loaded so far. |
| `span` (event) | Tells the page the low and high for the range bar: the line's range, or the last 24 hours of candles. |

## `src/components/ChartToolbar.vue`

| Function | What it does |
|---|---|
| `pick` | Chooses a tool; picking the one already on goes back to the cursor. The line tools are grouped under one button (ToolGroup). |

## `src/components/ToolGroup.vue`

| Function | What it does |
|---|---|
| `toggle` | Opens the list of line tools beside the button (under it on a phone). |
| `outside` | Closes the list when the visitor clicks anywhere else. |
| `choose` | Picks a tool from the list; the button then shows it. |

## `src/chart/drawingLayer.js`

| Function | What it does |
|---|---|
| `DrawingLayer` | A Lightweight Charts plugin that keeps the drawings, the drawing being placed and the ruler, and draws them on top of the chart (shapes.js draws each one). |
| `setBars` | Remembers the first bar's time and the bar spacing, so a time can be turned into a position even when it is scrolled out of view. |
| `x` / `y` / `xy` | The position of a time, a price, or a point on the chart, in pixels. |
| `hitTest` | Tells the chart which cursor to show over a drawing: a hand over the drawing, a move cursor over one of its points. |

## `src/chart/hitTest.js`

| Function | What it does |
|---|---|
| `hitDrawing` | Finds the topmost drawing under the mouse, and which part: one of its points, a channel's width handle, or the drawing itself. |
| `toSegment` | The distance from the mouse to a line, in pixels. |
| `inQuad` | Whether the mouse is inside a channel. |

## `src/chart/useSelect.js`

| Function | What it does |
|---|---|
| `select` | Marks a drawing as selected, or none. |
| `pointAt` | Turns a position on the chart into a time and price. |
| `moved` | The drawing shifted by the drag: the whole drawing, one of its points, or a channel's width. |
| `remove` | Deletes the selected drawing. |
| `install` | Listens before the chart does, so pressing on a drawing drags the drawing instead of scrolling the chart. |

## `src/chart/shapes.js`

| Function | What it does |
|---|---|
| `drawShape` | Draws one drawing: trend line, ray, extended line, horizontal or vertical line, channel, rectangle, Fibonacci or text. |
| `channel` | Draws two parallel lines, the space between them and a dashed middle line. |
| `fib` | Draws the Fibonacci levels (0, 0.236, 0.382, 0.5, 0.618, 0.786, 1) between two points, with their prices. |
| `drawRuler` | Draws the measurement: teal if up, red if down, with the % change, the price change, the number of bars and the time. |
| `extend` | Carries a line on to the edge of the chart, for rays and extended lines. |
| `duration` | Turns seconds into short text: 5400 → "1h 30m". |

## `src/chart/useDrawings.js`

| Function | What it does |
|---|---|
| `useDrawings` | Connects the toolbar to the chart and returns the tool, magnet and hide settings and the actions below. |
| `attach` | Puts the drawing layer on the series on screen (line or candles). |
| `pointFrom` | Turns a mouse position into a time and price. With the magnet on, the price snaps to the nearest value of the bar under the mouse. |
| `onClick` | Places drawings: one click for a horizontal or vertical line or text, three for a channel, two for the others. In cursor mode a click clears the ruler. |
| `onMove` | Moves the second point of the drawing being placed, or for a channel's last step, its width. |
| `finish` | Adds a drawing and goes back to the cursor, like TradingView. |
| `submitText` | Adds the typed text where it was clicked. |
| `cancel` | Stops whatever is being drawn. |
| `recount` | Counts the drawings and the ruler, so the bin is grey only when the chart is empty. |
| `clearAll` | Removes every drawing and the ruler. |
| `zoom` | Shows fewer or more bars, keeping the newest bar on the right. |
| `priceOn` | The price on a line at a given time, for the channel's width. |
| `install` | Listens to clicks on the chart element directly, because the chart's own click event drops a second click made within half a second of the first. A press that moved more than a few pixels is a drag, not a click. |

## `src/components/RecentTrades.vue`

| Function | What it does |
|---|---|
| `load` | Asks the backend for this token's latest trades and keeps any error to show. |
| `ada` | ₳1,250 for normal amounts, ₳4.35 for small ones. |

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
