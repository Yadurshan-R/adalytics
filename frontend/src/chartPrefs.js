// The visitor's chart choices, kept while they move between token pages.
import { ref } from 'vue'

export const chartType = ref('line') // 'line' or 'candles'
export const chartRange = ref('24H') // line: '24H', '7D' or '30D'
export const candleSize = ref('1h') // candles: '15m', '1h', '4h' or '1D'
export const indicators = ref([]) // picked indicators, e.g. ['ma', 'rsi']
