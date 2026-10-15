# Chart functions

Every function behind the price chart on a token page: the chart itself, its legend, units, indicators and drawing tools. Each one also has the same comment above it in the code. The rest of the website is in [FRONTEND_FUNCTIONS.md](FRONTEND_FUNCTIONS.md).

## `src/components/PriceChart.vue`

| Function | What it does |
|---|---|
| `rgba` | Turns a colour like #16c784 into rgba() with transparency a (0 to 1). |
| `isCandles` | Reports whether the chart shows candles instead of a line. |
| `paint` | Line mode: colour and bold/faded split for the hovered point (-1 = none). |
| `load` | Puts the current points or candles into the chart. The view is set only when the token, range, size or type changes, so a zoom or scroll survives the minute refresh. Candles open on the latest ones, about 14px wide each. |
| `fmt` / `volFmt` (props) | How the axis and the legend write values and volume: ₳ or $ price, or a compact FDV. |
| `onKey` | Esc stops the drawing being placed and unselects; Delete or Backspace removes the selected drawing. |

## `src/components/ChartLegend.vue`

| Function | What it does |
|---|---|
| `change` | The move in the chart's unit and percent: open to close for a candle, first price to this price for the line. Values and volume are written with the `fmt` and `volFmt` the chart passes in (₳ or $, price or FDV). |

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
| `buttons` | The quick sizes (15m, 1h, 4h, 1D), plus the picked size when it came from the list. |
| `progress` | How much of a size's history is loaded, 0 to 100; shown in green until it is all there. |
| `pick` | Chooses a size and closes the list. |

## `src/components/ChartPanel.vue`

| Function | What it does |
|---|---|
| `load` | Fetches the chart: the line for the picked range, or the candles at the picked size. Runs again every minute. |
| `setType` / `setRange` | Switch between line and candles, or the line's range. |
| `setUnit` / `setCurrency` | Switch between Price and FDV, and between ADA and USD (the sums are in useUnits.js). FDV shows only when the total supply is known, USD only when the USDM pool is tracked. |
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
| `useUnits` | Turns the chart's ADA prices into what the visitor picked: Price or FDV (price × total supply), in ADA or USD, like TapTools' chart switches. |
| `loadRates` | When USD is picked, fetches the USDM pool's chart for the same view. USDM's price in ADA is the number of ADA in one US dollar at that time. It comes from the backend's memory, so no extra Koios request. |
| `convert` | Scales each bar's prices and volume. In USD each bar uses the rate at its own time, so an old candle shows what the token was worth in dollars then, not at today's ADA price. |
| `fmt` / `volFmt` | How values and volume are written for the picked units. |
| `key` | Changes when the units change, so the chart fits again and old drawings are cleared (a line drawn at ₳0.18 means nothing on a $ chart). |
