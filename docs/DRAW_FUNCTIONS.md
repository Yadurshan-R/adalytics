# Drawing tool functions

Every function behind the drawing toolbar on the left of the price chart. The rest of the chart is in [CHART_FUNCTIONS.md](CHART_FUNCTIONS.md); how to use the tools is in [CHART_GUIDE.md](CHART_GUIDE.md).

The drawings themselves (placing, picking, moving, the handles, the labels, the magnet, the Esc and Delete keys) come from the library [lightweight-charts-drawing](https://github.com/deepentropy/lightweight-charts-drawing) (MIT licence, version pinned in package.json), TradingView's tools for Lightweight Charts. The files below are the toolbar and the few things the library leaves to the app.

## `src/components/DrawToolbar.vue`

| Function | What it does |
|---|---|
| (layout) | Like Birdeye's (TradingView's) toolbar: the tool groups of toolList.js; then Measure, zoom in, zoom out, reset; then magnet, stay in drawing mode, lock all, hide; then remove. A column beside the chart that scrolls when it does not fit; a row above the chart on a phone. |
| `last` / `shownOf` | Each group button shows, and picks, the last tool chosen from its menu (or the tool on now). |
| `item` | One menu row: key, name, icon, whether it is on or disabled. |
| `menuOf` | A group's menu, with its section headings (LINES, CHANNELS, PROJECTION…); none for a group of one tool. |
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
| `GROUPS` | The toolbar's group buttons and their menus: cursor; lines (trend line, ray, horizontal line and ray, vertical line, parallel channel); projection (long and short position, date and price range); brushes, arrows and shapes (brush, arrow, rectangle); text and notes (text, price label). 14 drawing tools, the ones that matter on a token chart. Adding one is adding its library kind here, with a name and an icon. |
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
| `show` | Switches to the drawings of one chart (token, Price/FDV and candle size): the ones on screen are put aside (drawBoards.js), and that timeframe's own come back. Called by PriceChart when the size, unit or token changes. |
| `removeAll` / `removeSelected` | Remove every drawing of this timeframe, or the selected ones that are not locked (the Delete button over the chart). |
| `install` / `onPress` | A click on the chart removes the measurement, like TradingView's ruler. |
| `uninstall` | Stops listening and removes the manager. |

## `src/chart/drawBoards.js`

| Function | What it does |
|---|---|
| `swap` | Puts the drawings on screen aside under their chart's key, clears the chart and brings back the drawings of the other key, if it has any. In memory only: everything goes when the page is reloaded. |
| `kept` | The drawings put aside for a key, used when the chart is built again (the page was left and opened again). |
| `forget` | Drops every chart's drawings (the tests). |

## `src/chart/textEditing.js`

| Function | What it does |
|---|---|
| `onTextEdit` / `submitText` / `cancelText` | The Text tool's box, which the library leaves to the app: open it where a text was placed or double-clicked; write the typed text on it; a new text left empty or cancelled is removed. |
