// useUnits turns the chart's prices into what the visitor picked with the
// Price / FDV switch: the price in ADA, or the fully diluted value (price ×
// total supply), like Birdeye's Price / MC and TapTools' Price / MCap.
import { computed } from 'vue'
import { formatPrice, formatCompact } from '../format.js'
import { chartUnit } from '../chartPrefs.js'

// props has supply (total supply); candles are the chart as the backend sends it.
export function useUnits(props, candles) {
  const fdv = computed(() => chartUnit.value === 'fdv' && props.supply > 0)

  // shown is the candles with each price multiplied by the supply for FDV.
  const shown = computed(() => {
    if (!fdv.value) return candles.value
    const k = props.supply
    return candles.value.map((c) => ({ ...c, open: c.open * k, high: c.high * k, low: c.low * k, close: c.close * k }))
  })

  // fmt writes a value on the axis and in the legend: ₳0.1903 or ₳4.57B.
  const fmt = computed(() => (fdv.value ? formatCompact : formatPrice))
  // key changes with the unit, so the chart fits again and drawings placed at
  // prices are cleared (a line at ₳0.19 means nothing on an FDV chart).
  const key = computed(() => (fdv.value ? 'fdv' : 'price'))

  return { fdv, shown, fmt, key }
}
