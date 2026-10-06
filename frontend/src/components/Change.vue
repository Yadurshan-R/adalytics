<!-- A percentage change: green ▲ when up, red ▼ when down, grey when it
     rounds to 0.0%. -->
<script setup>
import { computed } from 'vue'
import { formatPct } from '../format.js'

const props = defineProps({ value: { type: Number, default: null } })

const dir = computed(() => {
  if (props.value == null || !isFinite(props.value)) return 'none'
  // Shown with one decimal, so anything under 0.05% would read as 0.0%.
  if (props.value >= 0.05) return 'up'
  if (props.value <= -0.05) return 'down'
  return 'flat'
})
</script>

<template>
  <span class="change num" :class="dir">
    <template v-if="dir === 'none'">—</template>
    <template v-else>
      <svg v-if="dir !== 'flat'" class="caret" viewBox="0 0 10 6" aria-hidden="true">
        <path v-if="dir === 'up'" d="M5 0 10 6H0z" />
        <path v-else d="M0 0h10L5 6z" />
      </svg>
      {{ formatPct(value) }}
    </template>
  </span>
</template>

<style scoped>
.change { display: inline-flex; align-items: center; gap: 4px; font-weight: 500; white-space: nowrap; }
.change.none, .change.flat { color: var(--muted); }
.caret { width: 9px; height: 6px; fill: currentColor; }
</style>
