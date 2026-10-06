<!-- Small price line for the list: green if the last price is above the first, red if below. -->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  values: { type: Array, default: () => [] },
  width: { type: Number, default: 140 },
  height: { type: Number, default: 44 },
})

const color = computed(() => {
  const v = props.values
  if (v.length < 2) return 'var(--muted)'
  return v[v.length - 1] >= v[0] ? 'var(--bull)' : 'var(--bear)'
})

const points = computed(() => {
  const v = props.values
  if (v.length < 2) return ''
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
