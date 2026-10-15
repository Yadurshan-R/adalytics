// useUnits turns the chart's prices into what the visitor picked, like
// TapTools' chart switches:
//   - Price or FDV (fully diluted value: price × total supply);
//   - ADA or USD.
// USD uses the ADA/USD rate at each bar's own time, read from the USDM pool's
// chart (1 USDM ≈ 1 USD). So a candle from last week shows what the token was
// worth in dollars last week, not at today's ADA price. The rate comes from
// the backend's memory: no extra Koios request.
import { ref, computed } from 'vue'
import { formatPrice, formatUSD, formatCompact, formatADA } from '../format.js'
import { chartUnit, chartCurrency } from '../chartPrefs.js'

// USDM is the US dollar stablecoin whose ADA pool gives the ADA/USD rate.
const USDM = { policy: 'c48cbb3d5e57ed56e276bc45f99ab39abe94e6cd7ac39fb402da47ad', name: '0014df105553444d' }

// props has supply (total supply); tokens is the token list; points and
// candles are the chart in ADA, as the backend sends it.
export function useUnits(props, tokens, points, candles) {
  const usdmId = computed(() => tokens.value.find((t) => t.policy_id === USDM.policy && t.asset_name === USDM.name)?.id ?? '')
  const rates = ref([]) // [{ time, ada }]: ADA for one US dollar, oldest first

  const canUsd = computed(() => usdmId.value !== '')
  const wantUsd = computed(() => chartCurrency.value === 'usd' && canUsd.value)
  const usd = computed(() => wantUsd.value && rates.value.length > 0)
  const fdv = computed(() => chartUnit.value === 'fdv' && props.supply > 0)

  // loadRates fetches the USDM chart for the same view (q), when USD is picked.
  async function loadRates(q) {
    if (!wantUsd.value) return
    const res = await fetch(`/api/tokens/${encodeURIComponent(usdmId.value)}/chart?${q}`, { cache: 'no-store' })
    if (!res.ok) return
    const data = await res.json()
    rates.value = (data.candles ?? data.points ?? []).map((b) => ({ time: b.time, ada: b.close ?? b.price })).filter((r) => r.ada > 0)
  }

  // convert scales each bar's prices (keys) and volume. Each bar uses the
  // latest rate at or before its time.
  function convert(list, keys) {
    if (!fdv.value && !usd.value) return list
    const r = rates.value
    let j = 0
    return list.map((b) => {
      let rate = 1
      if (usd.value) {
        while (j + 1 < r.length && r[j + 1].time <= b.time) j++
        rate = r[j].ada
      }
      const k = (fdv.value ? props.supply : 1) / rate
      const out = { ...b, volume: b.volume / rate }
      for (const key of keys) out[key] = b[key] * k
      return out
    })
  }

  const shownPoints = computed(() => convert(points.value, ['price']))
  const shownCandles = computed(() => convert(candles.value, ['open', 'high', 'low', 'close']))

  // fmt writes a value on the axis and in the legend; volFmt writes volume.
  const fmt = computed(() => {
    const symbol = usd.value ? '$' : '₳'
    if (fdv.value) return (v) => formatCompact(v, symbol)
    return usd.value ? formatUSD : formatPrice
  })
  const volFmt = computed(() => (usd.value ? (v) => '$' + Math.round(v).toLocaleString('en-US') : formatADA))
  // key changes when the units change, so the chart fits again and old
  // drawings (placed at old values) are cleared.
  const key = computed(() => `${fdv.value ? 'fdv' : 'price'}-${usd.value ? 'usd' : 'ada'}`)

  return { fdv, usd, canUsd, key, shownPoints, shownCandles, fmt, volFmt, loadRates }
}
