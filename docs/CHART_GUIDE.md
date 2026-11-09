# Chart guide

How to read the price chart on a token page and what every button does. It works like TradingView's and Birdeye's charts, so what you learn here carries over. For the code behind it, see [CHART_FUNCTIONS.md](CHART_FUNCTIONS.md).

```
+-----------------------------------------------------------------------+
| 5m 15m 1h 4h 1D 1W 1M v | candles line | Indicators      Price / FDV   |  <- top bar
+------+----------------------------------------------------------------+
| tools| O 0.1868  H 0.1905  L 0.1862  C 0.1887  +1.05%  Vol 37,877    |  <- legend
|      |                                                   [0.1914]     |  <- latest price
|  +   |        candles                     | |                         |
|  /   |               | |    | |                                       |
|  =   |                                                                |
|  #   |   volume bars                                                  |
|  T   +----------------------------------------------------------------+
| ...  |   RSI panel (only when RSI is on)                              |
+------+----------------------------------------------------------------+
```

## Reading the chart

**A candle** is the trading in one period of time (the candle size, like 15 minutes):

* **O**pen: the price at the start of the period; **C**lose: the price at the end
* **H**igh and **L**ow: the highest and lowest price in the period: the thin line (the wick) above and below the body
* **Green** when it closed at or above where it opened (the price went up), **red** when it closed lower

**The legend** at the top left shows the candle under your mouse, or the latest one when the mouse is off the chart: O, H, L, C, how much the candle moved (in ADA and %), and **Vol**, the ADA traded in it. Vol is green or red like its candle.

**Volume bars** along the bottom show how much ADA was traded in each candle, green or red like the candle above. Tall bars mean many people traded: a move on high volume is more meaningful than one on low volume.

**The latest price** is the label on the right axis, with a dotted line across the chart.

**The crosshair** follows your mouse: dashed lines with the price (right axis) and the date and time (bottom axis) where you point.

## The top bar

| Button | What it does |
|---|---|
| **5m 15m 1h 4h 1D 1W 1M** | The candle size: one candle per 5 minutes, 15 minutes, 1 hour, 4 hours, 1 day, 1 week or 1 month. The **⌄** list has every size (also 30m, 2h and 12h). Small candles show detail, big ones show the trend. 5m goes 2 days back, 15m and 30m 7 days, the rest 30 days. |
| **1W / 1M** | Calendar candles, as on Birdeye, Binance and TradingView: a week candle starts on Monday 00:00 UTC, a month candle on the 1st 00:00 UTC. Adalytics keeps 30 days of history, so 1W shows about 5 candles and 1M one or two. A first week or month that starts before the history does is left out, because it would be missing days. The last candle is the running one and grows until its week or month ends. |
| **Candles / line** | Candles show open, high, low and close; the line joins the closes only, which is calmer to read for a trend. |
| **Indicators** | Adds lines worked out from the candles (see below). The number shows how many are on. |
| **History 64%** | Only while older history is still loading after the backend starts. The chart grows as it loads. |
| **Price / FDV** | Shown when the token's total supply is known. Price is ADA per token. FDV (fully diluted value) is the price × the total supply: what all the tokens would be worth. Use FDV to compare the size of two projects; price alone says nothing about that. |

Your choices (size, style, indicators, Price / FDV) stay the same when you open another token, until you reload the page.

## Indicators

They are worked out in your browser from the candles on screen, so they follow the candle size you pick.

| Indicator | What it is | How traders use it |
|---|---|---|
| **MA 20** (orange) | The average close of the last 20 candles. | Smooths out the noise. Price above it: short-term trend up; below: down. |
| **MA 50** (pink) | The average close of the last 50 candles. | The slower trend. When MA 20 crosses above MA 50, the trend may be turning up; below, down. |
| **RSI 14** (violet, own panel) | Relative strength index, 0 to 100, from the last 14 candles' ups and downs. Dashed lines at 70 and 30, a dotted one at 50, like TradingView. | Above 70: overbought (rose fast, may pull back). Below 30: oversold (fell fast, may bounce). |
| **ATR 14** (cyan, own panel) | Average true range: how much the price usually moves in one candle, in ADA, from the last 14 candles (gaps included), worked out the way TradingView's ATR is. | Size, not direction: a high ATR means wild moves. Traders set a stop-loss beyond the normal swing, for example 1.5 × ATR from the entry. A token nobody trades shows a low ATR, because its price did not move. |

Indicators describe the past; none of them predicts the price.

## Drawing tools

The toolbar on the left works like Birdeye's (TradingView's). Each button holds a group of tools: click it to use the tool it shows, or hover it and click the small arrow tab on its right to open the group's menu. The button then keeps the last tool you picked. The tool you are using has a grey square behind it. After one drawing the tool goes back to the cross (✛), like TradingView. The drawings are TradingView's own tools, from the open-source library lightweight-charts-drawing.

| Button | Tool | Clicks | What it is for |
|---|---|---|---|
| Cross ✛ | **Cross** | | The normal mouse: point, drag the chart, pick drawings. |
| Lines | **Trend line** | 2 | Joins two points, like two low points of a rising price: while the price stays above the line, the uptrend holds. |
| | **Ray** | 2 | A trend line that carries on to the right, to see where the trend would be later. |
| | **Horizontal line** | 1 | A price level across the chart, with its price on the right: support (where the price keeps stopping on the way down) or resistance (on the way up). |
| | **Horizontal ray** | 1 | A price level from one candle on to the right: a level that started at a certain moment. |
| | **Vertical line** | 1 | Marks a moment, like a big announcement. |
| | **Parallel channel** | 3 | Click two points for the first line, then a third for the width. The price often moves between the two lines. |
| Projection | **Long position** | 1 | Plans a buy: click the entry price. Green above is the target, red below the stop-loss; drag them. When selected it shows the target, the stop, the risk/reward ratio and the profit or loss so far, for a pretend account of 1,000 ADA risking 25% (TradingView's defaults). |
| | **Short position** | 1 | The same for a sell: target below, stop above. |
| | **Date and price range** | 2 | Measures between two points and keeps the measurement on the chart. |
| Shapes | **Brush** | drag | Draws freehand while the button is held. |
| | **Arrow** | 2 | Points at something. |
| | **Rectangle** | 2 | Marks an area: a range where the price went sideways, or a zone you watch. |
| Text | **Text** | 1 | Click, type a note (up to 80 characters), press Enter. Double-click it later to change it. |
| | **Price label** | 1 | A label with the price at that point. |

Hold **Shift** while drawing for straight angles (45° steps) and to turn the magnet off for a moment.

## Measure, zoom and the other buttons

| Button | What it does |
|---|---|
| **Measure** (ruler) | Click a start and an end point. Shows the change in ADA, in % and in ticks (steps of the fifth digit of the price), how many candles and how much time lies between them, and the volume traded. Click the chart again to remove it. |
| **Zoom in / out** | Fewer or more candles on screen, keeping the newest on the right. |
| **Reset view** | Fits all the history loaded so far on screen. |
| **Magnet** | While drawing, points snap to the open, high, low or close of the candle under the mouse, so a line starts exactly on a low or a high. Its menu has **Weak magnet** (snaps when the mouse is near one) and **Strong magnet** (always snaps). |
| **Stay in drawing mode** | Keeps the tool on after a drawing, to draw several in a row. Press Esc to stop. |
| **Lock all drawings** | Drawings can still be selected, but not moved or deleted with the Delete key, so nothing changes by accident. New drawings are locked too while it is on. |
| **Hide** (eye) | Hides every drawing to see the chart clean; click again to show them. Its menu can also hide the indicators, or both. |
| **Remove** (bin) | Removes every drawing. Its menu can also remove the indicators, or both. |

A switch between candles and line keeps the drawings.

**Changing a drawing:** with the cross on, click a drawing to select it. Drag it to move it, drag one of its round points to reshape it, and press Delete or Backspace (or the Delete button above the chart) to remove it. Esc unselects it or stops a drawing halfway.

**Drawings are not saved.** They stay while you are on the page, and are cleared when you open another token, switch between Price and FDV, or reload.

## Mouse, keyboard and phone

| Do this | To |
|---|---|
| Drag the chart | Scroll back in time |
| Mouse wheel on the chart | Zoom in and out (the page scrolls when the mouse is off the chart) |
| Drag the price axis up or down | Stretch or squash the price scale |
| Double-click an axis | Reset that axis |
| Esc | Stop drawing, unselect |
| Delete / Backspace | Remove the selected drawing |
| Pinch / drag (phone) | Zoom / scroll |

## Try it yourself (5 minutes)

1. Open a busy token from the list, like NIGHT or SNEK.
2. Move the mouse over a few candles and watch the legend: find a big green candle and check its Vol.
3. Click **1h**, then **1D**: the same history in bigger candles. Go back to **15m**.
4. Turn on **MA 20** under Indicators. Is the price above or below the orange line?
5. Turn on **RSI 14**. Find a moment it went above 70: what did the price do next?
6. Pick **Horizontal line** and click on a price where the candles stopped falling several times: that is support.
7. Pick **Measure** and click two candles to see how much the price moved between them.
8. Open the Projection menu, pick **Long position** and click a price: drag the green and red parts to a target and a stop-loss, and read the risk/reward ratio.
9. Click your horizontal line to select it, drag it, then press Delete.
10. Switch to **FDV** and compare with another token's FDV: which project is bigger?

This chart is for looking at the market, not advice on what to buy or sell.
