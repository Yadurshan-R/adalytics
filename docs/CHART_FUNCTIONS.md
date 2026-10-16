# Chart functions

Every function behind the price chart on a token page: the bar on top, the chart itself, its legend, Price / FDV, indicators and drawing tools. Each one also has the same comment above it in the code. The rest of the website is in [FRONTEND_FUNCTIONS.md](FRONTEND_FUNCTIONS.md).

## `src/components/PriceChart.vue`

| Function | What it does |
|---|---|
| `rgba` | Turns a colour like #089981 into rgba() with transparency a (0 to 1). |
| `isCandles` | Reports whether the chart shows candles instead of a line. |
| `showBar` | Puts a bar in the legend (the latest one when the mouse is off the chart), with the close before it for the line's change. |
| `load` | Puts the candles into the chart, as candles or as a line through each close. The view is set only when the token, size or unit changes, so a zoom or scroll survives the minute refresh and a switch between candles and line. Candles open on the latest ones, about 14px wide each; the line shows about four times as many, so it reads as a trend. The latest price has a label on the axis and a dotted line across the chart, like Birdeye and CoinMarketCap. |
| `fmt` (prop) | How the axis and the legend write values: a price, or a compact FDV. |
| `onKey` | Esc stops the drawing being placed and unselects; Delete or Backspace removes the selected drawing. |

## `src/components/ChartLegend.vue`

| Function | What it does |
|---|---|
| `change` | The move and percent: open to close for a candle, the close before to this close for the line. Values are written with the `fmt` the chart passes in (price or FDV). |

## `src/components/IndicatorMenu.vue`

| Function | What it does |
|---|---|
| `toggle` | Switches one indicator on or off. |
| `show` / `close` | Opens and closes the list; a click anywhere else closes it. |

## `src/chart/indicators.js`

| Function | What it does |
|---|---|
| `sma` | Simple moving average: the average close of the last n bars. |
| `ema` | Exponential moving average: recent bars count more, so it turns sooner. |
| `bollinger` | The n-bar average with a band two standard deviations above and below it. |
| `vwap` | Volume weighted average price of each day so far; starts again every day (UTC). |
| `rsi` | Relative strength index, 0 to 100 (Wilder's smoothing): above 70 overbought, below 30 oversold. |

## `src/chart/useIndicators.js`

| Function | What it does |
|---|---|
| `set` | Shows the picked indicators for the bars on the chart and removes the rest. RSI goes in its own panel under the chart. |
| `at` | Each indicator's value at the bar under the mouse, for the legend. |

## `src/components/CandleSizePicker.vue`

| Function | What it does |
|---|---|
| `buttons` | The quick sizes (5m, 15m, 1h, 4h, 1D), plus the picked size when it came from the list. |
| `progress` | How much of a size's history is loaded, 0 to 100; shown in the list of sizes until it is all there. |
| `pick` | Chooses a size and closes the list. |

## `src/components/ChartPanel.vue`

| Function | What it does |
|---|---|
| `load` | Fetches the candles at the picked size; candles and line both draw them. Runs again every minute. Each size shows the history loaded so far. |
| `span` (event) | Tells the page the low and high of the last 24 hours, for the range bar. |

## `src/components/ChartHeader.vue`

| Function | What it does |
|---|---|
| (template) | The slim bar on top of the chart, like Birdeye's: candle size, candles or line, Indicators, and Price / FDV on the right (only when the total supply is known). One bar for both chart styles, so switching style keeps the size. |
| `loaded` | How much of the whole history is in; a quiet "History 16%" note shows until it is all loaded. |

## `src/components/ChartToolbar.vue`

| Function | What it does |
|---|---|
| `pick` | Chooses a tool; picking the one already on goes back to the cursor. The line tools are grouped under one button (ToolGroup). |
| (layout) | In the order of Birdeye's and CoinMarketCap's bars: cross, lines, Fib, channel, rectangle, text; then ruler and zoom; then magnet, stay, lock, hide; then delete. Big icons; the column scrolls when it does not fit. |

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
| `finish` | Adds a drawing and goes back to the cursor, like TradingView, unless "stay in drawing mode" is on. |
| `submitText` | Adds the typed text where it was clicked. |
| `cancel` | Stops whatever is being drawn. |
| `recount` | Counts the drawings and the ruler, so the bin is grey only when the chart is empty. |
| `clearAll` | Removes every drawing and the ruler. |
| `zoom` | Shows fewer or more bars, keeping the newest bar on the right. |
| `priceOn` | The price on a line at a given time, for the channel's width. |
| `install` | Listens to clicks on the chart element directly, because the chart's own click event drops a second click made within half a second of the first. A press that moved more than a few pixels is a drag, not a click. |

## `src/chart/useUnits.js`

| Function | What it does |
|---|---|
| `useUnits` | Price / FDV, like Birdeye's Price / MC and TapTools' Price / MCap: the candles as they are, or each price multiplied by the total supply. |
| `fmt` | How values are written: ₳0.1903 for a price, ₳4.57B for an FDV. |
| `key` | Changes with the unit, so the chart fits again and old drawings are cleared (a line drawn at ₳0.19 means nothing on an FDV chart). |
