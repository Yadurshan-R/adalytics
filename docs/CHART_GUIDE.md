# Chart guide

How to read the price chart on a token page and what every button does. It works like TradingView's and Birdeye's charts, so what you learn here carries over. For the code behind it, see [CHART_FUNCTIONS.md](CHART_FUNCTIONS.md).

```
+-----------------------------------------------------------------------+
| 5m 15m 1h 4h 1D v | candles  line | Indicators          Price / FDV   |  <- top bar
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
| **5m 15m 1h 4h 1D** | The candle size: one candle per 5 minutes, 15 minutes, 1 hour, 4 hours or 1 day. The **⌄** list has every size (also 30m, 2h and 12h). Small candles show detail, big ones show the trend. 5m goes 2 days back, 15m and 30m 7 days, the rest 30 days. |
| **1 week / 30 days** (in the ⌄ list) | Shows that whole span at a fitting candle size (1h for a week, 4h for 30 days), once that much history has loaded. |
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
| **EMA 50** (purple) | An average of the last 50 candles where recent ones count more. | A slower trend line. When MA 20 crosses above EMA 50, the trend may be turning up; below, down. |
| **Bollinger Bands 20 2** (blue) | The MA 20 with a band 2 standard deviations above and below. | Narrow bands: a quiet market, often before a big move. Price at the top band: stretched up; at the bottom band: stretched down. |
| **VWAP** (pink) | The average price of the day, weighted by how much was traded at each price. Starts again every day (UTC). | Shows the "fair" price of the day. Buying below it is buying cheaper than most did today. |
| **RSI 14** (violet, own panel) | Relative strength index, 0 to 100, from the last 14 candles' ups and downs. Dashed lines at 70 and 30. | Above 70: overbought (rose fast, may pull back). Below 30: oversold (fell fast, may bounce). |

Indicators describe the past; none of them predicts the price.

## Drawing tools

Pick a tool on the left, then click on the chart. After one drawing the tool goes back to the cross (✛), like TradingView.

| Tool | Clicks | What it is for |
|---|---|---|
| **Cross** ✛ | | The normal mouse: point, drag the chart, pick drawings. |
| **Trend line** | 2 | Joins two points, like two low points of a rising price: while the price stays above the line, the uptrend holds. |
| **Ray** | 2 | A trend line that carries on to the right, to see where the trend would be later. |
| **Extended line** | 2 | Carries on both ways. |
| **Horizontal line** | 1 | A price level across the chart, with its price on the right: support (where the price keeps stopping on the way down) or resistance (on the way up). |
| **Vertical line** | 1 | Marks a moment, like a big announcement. |
| **Fib retracement** | 2 | Click a low and a high (or the other way). Draws the 0.236, 0.382, 0.5, 0.618 and 0.786 levels in between with their prices: after a big move, prices often pull back to one of these before going on. |
| **Parallel channel** | 3 | Click two points for the first line, then a third for the width. The price often moves between the two lines. |
| **Rectangle** | 2 | Marks an area: a range where the price went sideways, or a zone you watch. |
| **Text** | 1 | Click, type a note (up to 80 characters), press Enter. |

The **line tools** (trend, ray, extended, horizontal, vertical) share one button: the small › beside it opens the list, and the button keeps the last one you used.

## Measure, zoom and the other buttons

| Button | What it does |
|---|---|
| **Measure** (ruler) | Click a start and an end point. Shows the change in % and ADA, how many candles and how much time lies between them; green when up, red when down. Click the chart again to remove it. |
| **Zoom in / out** | Fewer or more candles on screen, keeping the newest on the right. |
| **Reset view** | Fits all the history loaded so far on screen. |
| **Magnet** | While drawing, points snap to the nearest open, high, low or close of the candle under the mouse, so a line starts exactly on a low or a high. |
| **Stay in drawing mode** | Keeps the tool on after a drawing, to draw several in a row. Press Esc to stop. |
| **Lock drawings** | Drawings can no longer be picked or moved by accident. |
| **Hide drawings** | Hides them all to see the chart clean; click again to show them. |
| **Delete all** (bin) | Removes every drawing and the ruler. |

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
7. Pick **Fib retracement**, click the lowest point of a big rise, then its top. Did the price pull back to one of the levels?
8. Pick **Measure** and click two candles to see how much the price moved between them.
9. Click your horizontal line to select it, drag it, then press Delete.
10. Switch to **FDV** and compare with another token's FDV: which project is bigger?

This chart is for looking at the market, not advice on what to buy or sell.
