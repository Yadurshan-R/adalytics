# Chart functions

Every function behind the price chart on a token page: the bar on top, the chart itself, its legend, Price / FDV, the loading box and indicators. The drawing tools are in [DRAW_FUNCTIONS.md](DRAW_FUNCTIONS.md). Each one also has the same comment above it in the code. The rest of the website is in [FRONTEND_FUNCTIONS.md](FRONTEND_FUNCTIONS.md). How to use the chart, for people using the site, is in [CHART_GUIDE.md](CHART_GUIDE.md).

## `src/components/PriceChart.vue`

| Function | What it does |
|---|---|
| `applyTheme` | Repaints the chart, candles, line and volume in the current theme's colours when the light / dark switch is used; the zoom stays. |
| `isCandles` | Reports whether the chart shows candles instead of a line. |
| `showBar` | Puts a bar in the legend (the latest one when the mouse is off the chart), with the close before it for the line's change. |
| `load` | Puts the candles into the chart, as candles or as a line through each close. The view is set only when the token, size or unit changes, so a zoom or scroll survives the minute refresh and a switch between candles and line. Candles open on the latest ones, about 14px wide each; the line shows about four times as many, so it reads as a trend. The latest price has a label on the axis and a dotted line across the chart, like Birdeye and CoinMarketCap. It also sets the price step (priceStep) and puts the drawings on the series on screen. 5m to 12h are moved to the visitor's clock first (localTime.js), so the new day starts at local midnight. |
| `timeAxis` | Sets the time axis and the date under the mouse for the candle size (`size` prop): clock times for 5m to 12h, dates only for 1D, 1W and 1M. |
| `hide` | The toolbar's hide menu: hides or shows the drawings or the indicators; "Hide all" hides both, or shows both when both are hidden. |
| `remove` | The toolbar's remove menu: removes the drawings, the indicators (unpicks them in the Indicators menu) or both. |
| `fmt` (prop) | How the axis and the legend write values: a price, or a compact FDV. |
| (slot) | Whatever ChartPanel puts over the chart area, here the loading box (ChartWaiting). |

## `src/chart/chartOptions.js`

| Function | What it does |
|---|---|
| `chartColors` | Reads the chart's colours from the theme (style.css): background, grid, text, TradingView's green and red candles, the blue line. |
| `rgba` | Turns a colour like #089981 into rgba() with transparency a (0 to 1). |
| `themeOptions` | The chart options that change with light / dark: background, text, grid, crosshair labels. |
| `chartOptions` | The options the chart is created with: theme colours, axes, dashed crosshair, the time axis and the date under the mouse (localTime.js), drag and zoom (view.js). `maxBarSpacing` 40 keeps charts with few candles (1W, 1M) at a normal candle width, like Birdeye, instead of a few huge candles. |

## `src/chart/localTime.js`

| Function | What it does |
|---|---|
| `isDaySize` / `chartShift` | 1D, 1W and 1M are whole UTC days, like Birdeye's and Binance's: they are not moved and show only their date. 5m to 12h are moved by the visitor's offset (19,800 seconds in Sri Lanka), so the new day starts at local midnight, as the library's docs advise. |
| `localShift` | The seconds between UTC and the visitor's clock. One offset for the whole chart, so times stay in order across a daylight saving change. |
| `toLocal` | Moves the candles by that shift before they are drawn. |
| `tickMarkFor` | The time axis for a size: hours like "16:00"; a new day as "5 Oct"; a month or year as its name. 1M writes "Nov"; no day candle gets a clock time. Written as UTC, because the times are already moved. |
| `crosshairFor` | The date under the mouse: "Tue 13 Oct, 10:00"; "Mon 9 Nov" (1D); "Week of 2 Nov" (1W); "November 2026" (1M). |

## `src/chart/view.js`

| Function | What it does |
|---|---|
| `SCROLL` / `SCALE` | The chart can be dragged sideways and zoomed with the wheel, a pinch or by dragging the axes; the wheel does not scroll it (it stays for scrolling the page). |
| `zoom` | Shows fewer or more bars (the zoom buttons), keeping the newest bar on the right. |
| `resetView` | Shows every bar loaded (the reset button). |

## `src/components/ChartLegend.vue`

| Function | What it does |
|---|---|
| `change` | The move and percent: open to close for a candle, the close before to this close for the line. Values are written with the `fmt` the chart passes in (price or FDV). A candle with no move reads "0 (0.00%)", like TradingView. |
| (indicators) | Each picked indicator's value, or "needs 50 candles" in grey when there are too few candles for it. |
| `volDir` | Colours the Vol number like its volume bar: green when the candle closed at or above its open, red when below, like TradingView. |

## `src/components/IndicatorMenu.vue`

| Function | What it does |
|---|---|
| `toggle` | Switches one indicator on or off. |
| `show` / `close` | Opens and closes the list; a click anywhere else closes it. |

## `src/chart/indicators.js`

| Function | What it does |
|---|---|
| `INDICATORS` | The Indicators menu: MA 20, MA 50, RSI 14 and ATR 14, the ones most traders use, with their colour, length, panel and whether their value is a price. |
| `minBars` | How many candles an indicator needs for its first value: n, one more for RSI. |
| `sma` | Simple moving average: the average close of the last n bars. |
| `wilder` | Wilder's smoothing, used by RSI and ATR: the average of the first n values, then each new value counts 1/n. |
| `rsi` | Relative strength index, 0 to 100: above 70 overbought, below 30 oversold. |
| `trueRange` | How far the price moved in one bar, gaps included: the largest of high − low, \|high − previous close\| and \|low − previous close\|. |
| `atr` | Average true range: how much the price usually moves in one bar, in price; size only, no direction. |

## `src/chart/useIndicators.js`

| Function | What it does |
|---|---|
| `panelHeight` / `RSI_LEVELS` | An indicator panel's height (110px, 70px on a phone); RSI's lines as TradingView draws them: 70 and 30 dashed, 50 dotted, on an axis that always shows 0 to 100. |
| `create` | Adds an indicator's line: the moving averages on the price, RSI and ATR each in a new panel under the chart (with RSI's level lines). The chart removes an empty panel by itself. |
| `dataOf` | Works out an indicator's line from the bars. |
| `set` | Shows the picked indicators for the bars on the chart and removes the rest. One with too few candles for any value (MA 50 on 1D, all four on 1W and 1M) gets no line and no empty panel; the legend says how many candles it needs. |
| `at` | Each indicator's value at the bar under the mouse, or how many candles it needs, for the legend (nothing while the indicators are hidden). |
| `show` | Hides or shows every indicator, lines and legend (the toolbar's "Hide indicators"); indicators added while hidden stay hidden. |

## `src/components/CandleSizePicker.vue`

| Function | What it does |
|---|---|
| `buttons` | The quick sizes (5m, 15m, 1h, 4h, 1D, 1W, 1M), plus the picked size when it came from the list. Every button is a candle size, like Birdeye's; there are no range buttons. |
| `pick` | Chooses a size and closes the list. |

## `src/components/ChartPanel.vue`

| Function | What it does |
|---|---|
| `load` | Fetches the candles at the picked size; candles and line both draw them. Runs again every minute. Each size shows the history loaded so far. |
| `wait` | How far the history is from what the picked size needs (loadWait); while it is short the loading box covers the chart, and the candles are fetched again the moment it is there. |
| `empty` | A 1W or 1M chart with no candle at all once fetched (a 31-day month on its 31st day), so the box says "No full month yet". |
| `span` (event) | Tells the page the low and high of the last 24 hours, for the range bar. |

## `src/components/ChartHeader.vue`

| Function | What it does |
|---|---|
| (template) | The slim bar on top of the chart, like Birdeye's: candle size, candles or line, Indicators, and Price / FDV on the right (only when the total supply is known). One bar for both chart styles, so switching style keeps the size. |

## `src/components/ChartWaiting.vue`

| Function | What it does |
|---|---|
| (template) | The loading box over the chart, like TradingView's and DexScreener's: "Loading 1 day candles", a bar and the percentage, until the picked size has enough history. The drawing toolbar stays usable. |
| `name` | The size in words, for the title ("1 hour", "1 week"). |

## `src/chart/loadWait.js`

| Function | What it does |
|---|---|
| `MIN_CANDLES` / `SIZES` | A size is shown once it has 20 candles, as Birdeye and TradingView open with; each size's length, how far back it goes and its name. |
| `needHours` | The hours of history a size needs: 20 candles, never more than its span or the history kept. 1W and 1M need all 30 days, which hold only a few of them. |
| `loadWait` | `{ pct }` while the history is shorter than that (0 to 99), null once it is there or the whole history is in. |

## Tests (`npm test`, Vitest)

| Test | What it checks |
|---|---|
| `priceStep.test.js` | The step is one unit of the fifth significant digit (SNEK, HOSKY, LQ, an FDV); a real move counts 3,297 ticks, not 329 million; no price gives the fallback. |
| `toolList.test.js` | Every tool on the toolbar exists in the library and has a name and an icon; each tool is in one group; the 14 tools chosen. |
| `indicators.test.js` | The menu lists MA 20, MA 50, RSI 14, ATR 14; the moving average, RSI (100 all up, 0 all down, 50 for equal moves), the true range with gaps, and ATR give the textbook values on hand-checked bars; with too few candles they give no value (not a wrong one), and the first value comes exactly at `minBars`. |
| `localTime.test.js` | The offset of Colombo and New York; only 5m to 12h are moved (a 1D candle stays on 9 Nov in New York); intraday times written as the local clock, "9 Nov" at local midnight; 1D, 1W, 1M written as dates ("Nov", "Week of 2 Nov", "Mon 9 Nov"), never a clock time. |
| `loadWait.test.js` | 20 candles per size (1h after 20 hours, 4h after 80, 1D after 20 days, 1W after all 30 days), capped by the history kept; never 100% while something is missing. |
| `useDrawTools.test.js` | With a pretend manager: drawings move with a switch between candles and line (with the magnet and hide settings); the measurement is not counted and goes on the next click; lock all locks new drawings too and keeps locked ones from Delete; a button picking the tool on goes back to the cursor, a menu does not; typed text is written, and an empty new text removed; the picked tool stays after a drawing by default; each timeframe keeps its own drawings (hidden on 4h, back on 1h), a deleted drawing or the measurement is not kept, and drawings come back when the chart is built again. |

## `src/chart/useUnits.js`

| Function | What it does |
|---|---|
| `useUnits` | Price / FDV, like Birdeye's Price / MC and TapTools' Price / MCap: the candles as they are, or each price multiplied by the total supply. |
| `fmt` | How values are written: ₳0.1903 for a price, ₳4.57B for an FDV. |
| `key` | Changes with the unit, so the chart fits again and old drawings are cleared (a line drawn at ₳0.19 means nothing on an FDV chart). |
