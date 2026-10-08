// The visitor's chart choices, kept while they move between token pages.
import { ref } from 'vue'

export const chartType = ref('line') // 'line' or 'candles'
export const chartRange = ref('24H') // '24H', '7D' or '30D'
