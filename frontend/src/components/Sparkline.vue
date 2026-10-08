<!-- Small price line for the list: green if the last price is above the first,
     red if below, and grey through the middle when the price did not move. -->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  values: { type: Array, default: () => [] },
  width: { type: Number, default: 140 },
  height: { type: Number, default: 44 },
})

// No trades in the period: every value is the same price.
const flat = computed(() => {
  const v = props.values
  return v.length >= 2 && Math.max(...v) === Math.min(...v)
})

const color = computed(() => {
  const v = props.values
  if (v.length < 2 || flat.value) return '#a6b0c3'
  const first = v[0]
  const last = v[v.length - 1]
  return last > first ? 'var(--bull)' : last < first ? 'var(--bear)' : '#a6b0c3'
})

const points = computed(() => {
  const v = props.values
  if (v.length < 2) return ''
  if (flat.value) {
    const mid = (props.height / 2).toFixed(1)
    return `0,${mid} ${props.width},${mid}`
  }
  const min = Math.min(...v)
  const max = Math.max(...v)
  const span = max - min || 1
  const pad = 3
  return v
    .map((p, i) => {
      const x = (i / (v.length - 1)) * props.width
      const y = pad + (1 - (p - min) / span) * (props.height - pad * 2)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
})
</script>

<template>
  <svg :width="width" :height="height" :viewBox="`0 0 ${width} ${height}`" class="spark" aria-hidden="true">
    <polyline v-if="points" :points="points" fill="none" :stroke="color" stroke-width="1.6" stroke-linejoin="round" stroke-linecap="round" />
  </svg>
</template>

<style scoped>
.spark { display: block; margin-left: auto; }
</style>
