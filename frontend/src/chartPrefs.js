// The visitor's chart choices, kept while they move between token pages.
import { ref } from 'vue'

export const chartType = ref('candles') // 'candles' or 'line'
export const candleSize = ref('1h') // '5m', '15m', '30m', '1h', '2h', '4h', '12h' or '1D'
export const indicators = ref([]) // picked indicators, e.g. ['ma', 'rsi']
export const chartUnit = ref('price') // 'price', or 'fdv': the chart in fully diluted value
