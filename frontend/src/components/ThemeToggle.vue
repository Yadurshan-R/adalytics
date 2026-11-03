<!-- The light / dark switch at the right of the top bar: a sun in dark mode
     (click for light), a moon in light mode (click for dark). -->
<script setup>
import { computed } from 'vue'
import { theme, toggleTheme } from '../theme.js'

const dark = computed(() => theme.value === 'dark')
const label = computed(() => (dark.value ? 'Switch to light mode' : 'Switch to dark mode'))
</script>

<template>
  <button type="button" class="toggle" :aria-label="label" :title="label" :aria-pressed="dark" @click="toggleTheme">
    <svg v-if="dark" viewBox="0 0 20 20" aria-hidden="true">
      <circle cx="10" cy="10" r="3.6" />
      <path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.3 4.3l1.4 1.4M14.3 14.3l1.4 1.4M4.3 15.7l1.4-1.4M14.3 5.7l1.4-1.4" />
    </svg>
    <svg v-else viewBox="0 0 20 20" aria-hidden="true">
      <path d="M16.5 12.3A7 7 0 0 1 7.7 3.5a7 7 0 1 0 8.8 8.8Z" />
    </svg>
  </button>
</template>

<style scoped>
.toggle { display: inline-flex; align-items: center; justify-content: center; width: 40px; height: 40px; margin-left: auto; border: 1px solid var(--line); border-radius: 8px; background: none; color: var(--fg); cursor: pointer; }
.toggle:hover { background: var(--hover); }
.toggle:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.toggle svg { width: 19px; height: 19px; fill: none; stroke: currentColor; stroke-width: 1.6; stroke-linecap: round; stroke-linejoin: round; }
</style>
