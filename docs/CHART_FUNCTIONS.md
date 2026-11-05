# Chart functions

Every function behind the price chart on a token page: the bar on top, the chart itself, its legend, Price / FDV, indicators and drawing tools. Each one also has the same comment above it in the code. The rest of the website is in [FRONTEND_FUNCTIONS.md](FRONTEND_FUNCTIONS.md). How to use the chart, for people using the site, is in [CHART_GUIDE.md](CHART_GUIDE.md).

## `src/components/PriceChart.vue`

| Function | What it does |
|---|---|
| `applyTheme` | Repaints the chart, candles, line and volume in the current theme's colours when the light / dark switch is used; the zoom stays. |
| `isCandles` | Reports whether the chart shows candles instead of a line. |
| `showBar` | Puts a bar in the legend (the latest one when the mouse is off the chart), with the close before it for the line's change. |
| `load` | Puts the candles into the chart, as candles or as a line through each close. The view is set only when the token, size or unit changes, so a zoom or scroll survives the minute refresh and a switch between candles and line. Candles open on the latest ones, about 14px wide each; the line shows about four times as many, so it reads as a trend. The latest price has a label on the axis and a dotted line across the chart, like Birdeye and CoinMarketCap. It also sets the price step (priceStep) and puts the drawings on the series on screen. |
| `hide` | The toolbar's hide menu: hides or shows the drawings or the indicators; "Hide all" hides both, or shows both when both are hidden. |
| `remove` | The toolbar's remove menu: removes the drawings, the indicators (unpicks them in the Indicators menu) or both. |
| `fmt` (prop) | How the axis and the legend write values: a price, or a compact FDV. |

## `src/chart/chartOptions.js`

| Function | What it does |
|---|---|
| `chartColors` | Reads the chart's colours from the theme (style.css): background, grid, text, TradingView's green and red candles, the blue line. |
| `rgba` | Turns a colour like #089981 into rgba() with transparency a (0 to 1). |
| `themeOptions` | The chart options that change with light / dark: background, text, grid, crosshair labels. |
| `tickMark` | Writes the time axis: hours like "16:00"; a new day as "5 Oct", like CoinGecko; a new month or year as its name. |
| `chartOptions` | The options the chart is created with: theme colours, axes, dashed crosshair, the date under the mouse ("Tue 13 Oct, 10:00"), drag and zoom (view.js). |

## `src/chart/view.js`

| Function | What it does |
|---|---|
| `SCROLL` / `SCALE` | The chart can be dragged sideways and zoomed with the wheel, a pinch or by dragging the axes; the wheel does not scroll it (it stays for scrolling the page). |
| `zoom` | Shows fewer or more bars (the zoom buttons), keeping the newest bar on the right. |
| `resetView` | Shows every bar loaded (the reset button). |

## `src/components/ChartLegend.vue`

| Function | What it does |
|---|---|
| `change` | The move and percent: open to close for a candle, the close before to this close for the line. Values are written with the `fmt` the chart passes in (price or FDV). |
| `volDir` | Colours the Vol number like its volume bar: green when the candle closed at or above its open, red when below, like TradingView. |

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
| `at` | Each indicator's value at the bar under the mouse, for the legend (nothing while the indicators are hidden). |
| `show` | Hides or shows every indicator, lines and legend (the toolbar's "Hide indicators"); indicators added while hidden stay hidden. |

## `src/components/CandleSizePicker.vue`

| Function | What it does |
|---|---|
| `buttons` | The quick sizes (5m, 15m, 1h, 4h, 1D), plus the picked size when it came from the list, or the range (7D, 30D) while one is on. |
| `pickRange` | 1 week or 30 days from the list: picked only once its history is fully loaded; until then it is grey with "Loading 35%". |
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
| `ready` | How much of a time range (7D, 30D) is loaded, as the backend reports it. |
| `pickRange` | Shows a time range with a fitting candle size (1 week: 1h, 30 days: 4h); picking it again goes back to the normal view. |
| `pickSize` | A candle size picked by hand; no range is on any more. |

## Drawing tools

The drawings themselves (placing, picking, moving, the handles, the labels, the magnet, the Esc and Delete keys) come from the library [lightweight-charts-drawing](https://github.com/deepentropy/lightweight-charts-drawing) (MIT licence, version pinned in package.json), TradingView's tools for Lightweight Charts. The files below are the toolbar and the few things the library leaves to the app.

## `src/components/DrawToolbar.vue`

| Function | What it does |
|---|---|
| (layout) | Like Birdeye's (TradingView's) toolbar: the tool groups of toolList.js; then Measure, zoom in, zoom out, reset; then magnet, stay in drawing mode, lock all, hide; then remove. A column beside the chart that scrolls when it does not fit; a row above the chart on a phone. |
| `last` / `shownOf` | Each group button shows, and picks, the last tool chosen from its menu (or the tool on now). |
| `item` | One menu row: key, name, icon, whether it is on or disabled. |
| `menuOf` | A group's menu, with its section headings (LINES, CHANNELS, FIBONACCI…); none for a group of one tool. |
| `arm` | A tool chosen from a menu: the group remembers it and it is picked. |
| `magnetShown` / `chooseMagnet` | The magnet button switches the magnet on in its last mode, or off; its menu picks weak or strong. |
| `plural` / `hideMenu` / `removeMenu` | The hide menu (drawings, indicators, all) and the remove menu ("Remove 3 drawings", "Remove 1 indicator", "Remove all"). |

## `src/components/ToolButton.vue`

| Function | What it does |
|---|---|
| (look) | A grey square behind the tool picked; a setting that is on (magnet, stay, lock, hide) drawn inverted, like Birdeye's. |
| `toggle` | The arrow tab, shown on hover, opens the button's menu beside it (under it on a phone), or closes it. |
| `choose` | Passes the menu item chosen on and closes the menu. |

## `src/components/ToolMenu.vue`

| Function | What it does |
|---|---|
| (layout) | Section headings in small capitals, then one row per item with its icon and name; the item that is on is drawn inverted. |
| `outside` / `onKey` | A click anywhere else or the Esc key closes the menu. |

## `src/components/ToolIcon.vue`

| Function | What it does |
|---|---|
| (icon) | Draws a toolbar or menu icon (toolIcons.js): one path on a 20 × 20 grid, with small circles for the points the visitor clicks. |

## `src/components/ChartTextBox.vue`

| Function | What it does |
|---|---|
| (box) | The box for typing a text where the Text tool was placed or a text was double-clicked. Enter or a click elsewhere keeps it; Esc drops it. |
| `vFocus` | Puts the cursor in the box a moment after it appears: the library places a text when the button goes down, and the browser's own mousedown would otherwise take the focus straight back. |

## `src/chart/toolList.js`

| Function | What it does |
|---|---|
| `GROUPS` | The toolbar's group buttons and their menus: cursor; lines (trend line, ray, horizontal line and ray, vertical line, parallel channel); Fibonacci (retracement, trend-based extension); projection (long and short position, date and price range); brushes, arrows and shapes (brush, arrow, rectangle); text and notes (text, price label). 16 drawing tools, the ones that matter on a token chart. Adding one is adding its library kind here, with a name and an icon. |
| `LABELS` | Each tool's name, as the menus and tooltips show it. |
| `kindsOf` | Every tool of a group, in menu order. |
| `groupOf` | The group a tool belongs to. |
| `DRAWING_KINDS` | Every tool placed with the library (all but the cursor). |

## `src/chart/toolIcons.js`

| Function | What it does |
|---|---|
| `ICONS` | The icon of every tool and toolbar action, drawn for this app in the thin-line style of TradingView's toolbar (not copied from it). |
| `ring` | A circle written as a path, so every icon is one path. |

## `src/chart/priceStep.js`

| Function | What it does |
|---|---|
| `priceStep` | The smallest price change the chart works with: one unit of the fifth significant digit (0.004650 → 0.0000001). The drawings count "ticks" in it (Measure, long / short position) and choose their decimals from it; one tiny step for every token made those counts absurd (329,707,313 ticks for a move of 0.00033 ADA). |
| `FALLBACK_STEP` | The step while there is no price yet. |

## `src/chart/useDrawTools.js`

| Function | What it does |
|---|---|
| `useDrawTools` | Connects the toolbar to the library and returns the tool, magnet, stay, lock, hide, count, selection and text box, and the actions below. Nothing is saved: drawings live in memory. |
| `attach` | Puts the drawings on the series on screen. After a switch between candles and line it builds a manager on the new series and moves every drawing over. |
| `detach` | Removes the manager. |
| `recount` | Counts the drawings, the measurement not counted, for the remove button. |
| `onAdd` | A measurement ends the Measure tool; while "lock all" is on, a new drawing is locked too. |
| `done` | The library put the tool down (a drawing was placed, or Esc): back to the cursor. |
| `pick` | Chooses a tool. Picking the tool that is on goes back to the cursor from a button, not from a menu. |
| `clearMeasure` | Removes the measurement. |
| `setMagnet` | Off, weak (snaps to the open, high, low or close when near one) or strong (always snaps). |
| `setStay` | Stay in drawing mode: keep the tool after a drawing, to draw several in a row. |
| `setLocked` | Locks or unlocks every drawing: a locked drawing can be selected but not moved or deleted with the Delete key. |
| `setHidden` | Hides or shows every drawing. |
| `removeAll` / `removeSelected` | Remove every drawing, or the selected ones that are not locked (the Delete button over the chart). |
| `onTextEdit` / `submitText` / `cancelText` | Open the text box where a text was placed or double-clicked; write the typed text on it; a new text left empty is removed. |
| `install` / `onPress` | A click on the chart removes the measurement, like TradingView's ruler. |
| `uninstall` | Stops listening and removes the manager. |

## Tests (`npm test`, Vitest)

| Test | What it checks |
|---|---|
| `priceStep.test.js` | The step is one unit of the fifth significant digit (SNEK, HOSKY, LQ, an FDV); a real move counts 3,297 ticks, not 329 million; no price gives the fallback. |
| `toolList.test.js` | Every tool on the toolbar exists in the library and has a name and an icon; each tool is in one group; the 16 tools chosen. |
| `useDrawTools.test.js` | With a pretend manager: drawings move with a switch between candles and line (with the magnet and hide settings); the measurement is not counted and goes on the next click; lock all locks new drawings too and keeps locked ones from Delete; a button picking the tool on goes back to the cursor, a menu does not; typed text is written, and an empty new text removed. |

## `src/chart/useUnits.js`

| Function | What it does |
|---|---|
| `useUnits` | Price / FDV, like Birdeye's Price / MC and TapTools' Price / MCap: the candles as they are, or each price multiplied by the total supply. |
| `fmt` | How values are written: ₳0.1903 for a price, ₳4.57B for an FDV. |
| `key` | Changes with the unit, so the chart fits again and old drawings are cleared (a line drawn at ₳0.19 means nothing on an FDV chart). |
