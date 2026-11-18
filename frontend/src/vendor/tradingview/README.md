# TradingView's own code (do not edit)

These files are copied unchanged from TradingView's open-source
Lightweight Charts repository, so the chart's indicators are TradingView's
code, not ours.

- Source: https://github.com/tradingview/lightweight-charts
  (`indicator-examples/src`), commit `c29c51e` (16 Nov 2026), the code of
  version 5.3.0.
- Licence: Apache 2.0 (`LICENSE`), with TradingView's notice (`NOTICE`).
  Both must stay in this folder.
- They are TypeScript; Vite and Vitest read `.ts` files as they are, so
  nothing extra is needed to build.
- They are longer than our 200-line rule (the moving average is 238 lines);
  the rule is for our own files, and these must stay as TradingView wrote them.

| File | What it works out |
| --- | --- |
| `indicators/moving-average/` | Average close of the last n candles |
| `indicators/momentum/` | Close now minus the close n candles ago |
| `indicators/percent-change/` | % change of the close from the candle before |
| `indicators/average-price/` | (open + high + low + close) / 4 |
| `indicators/median-price/` | (high + low) / 2 |
| `indicators/weighted-close/` | (high + low + 2 × close) / 4 |
| `helpers/timestamp-data.ts` | Checks every candle has a numeric time |

## Our only changes (marked `// Adalytics:` in the code)

TradingView's examples have a bug: four of them leave out the newest candle,
so the line stopped one candle short of the price. We changed one line in each
and nothing else:

| File | Change |
| --- | --- |
| `average-price-calculation.ts` | `endIndex` is `data.length` (was `data.length - 1`); removed a `console.log(result)` that printed every refresh |
| `median-price-calculation.ts` | `endIndex` is `data.length` (was `data.length - 1`) |
| `weighted-close-calculation.ts` | `endIndex` is `data.length` (was `data.length - 1`) |
| `percent-change-calculation.ts` | `endIndex` is `data.length` (was `data.length - 1`) |

`indicators.test.js` checks the newest candle has a value, so the bug cannot
come back unnoticed.

To update them: copy the same files again from a newer TradingView release,
make the changes above again if TradingView has not fixed them, change the
commit above, and run `npm test`.

The pair indicators in the same folder (correlation, ratio, spread, product,
sum) compare two tokens, so they are not copied.
