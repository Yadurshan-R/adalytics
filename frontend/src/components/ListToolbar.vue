<!-- The bar above the token list: Top Volume / Top Gainers / Top Losers, the
     24h / 7d switch and the search box. Press "/" anywhere to jump to the
     search box, Esc to clear it. -->
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

defineProps({
  views: { type: Array, required: true }, // [{ key, label }]
  view: { type: String, required: true },
  query: { type: String, required: true },
  week: { type: Boolean, default: false }, // the list shows 7 day numbers
  weekRange: { type: Object, default: null }, // the 7 day history's loading status
  weekReady: { type: Boolean, default: false },
})
const emit = defineEmits(['update:view', 'update:query', 'period'])

const search = ref(null)
// onKey makes "/" jump to the search box from anywhere on the page.
function onKey(e) {
  if (e.key === '/' && document.activeElement !== search.value) {
    e.preventDefault()
    search.value?.focus()
  }
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="toolbar">
    <div class="pickers">
      <div class="views" role="group" aria-label="Order the list">
        <button v-for="v in views" :key="v.key" type="button" :aria-pressed="view === v.key" @click="emit('update:view', v.key)">
          {{ v.label }}
        </button>
      </div>
      <div v-if="weekRange" class="period">
        <button type="button" class="side" :class="{ on: !week }" @click="emit('period', '24h')">24h</button>
        <button type="button" class="toggle" role="switch" aria-label="Show 7 days" :aria-checked="week"
                :disabled="!weekReady" @click="emit('period', week ? '24h' : '7d')"><span class="knob"></span></button>
        <button type="button" class="side" :class="{ on: week }" :disabled="!weekReady" @click="emit('period', '7d')">
          7d<span v-if="!weekReady" class="pct num">{{ weekRange.progress }}%</span>
        </button>
      </div>
    </div>
    <label class="search">
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M8.5 3a5.5 5.5 0 1 0 3.4 9.8l3.6 3.7 1.1-1.1-3.7-3.6A5.5 5.5 0 0 0 8.5 3Zm0 1.5a4 4 0 1 1 0 8 4 4 0 0 1 0-8Z" /></svg>
      <input ref="search" :value="query" type="search" placeholder="Search token" aria-label="Search token by name or ticker"
             autocomplete="off" spellcheck="false" @input="emit('update:query', $event.target.value)" @keydown.esc="emit('update:query', '')" />
      <kbd v-if="!query" aria-hidden="true">/</kbd>
    </label>
  </div>
</template>

<style scoped>
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; margin-bottom: 16px; }
.pickers { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.views { display: inline-flex; background: var(--chip); border-radius: 10px; padding: 4px; gap: 2px; }
.views button { font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 7px 14px; border-radius: 8px; cursor: pointer; white-space: nowrap; }
.views button:hover { color: var(--fg); }
.views button[aria-pressed="true"] { background: var(--bg); color: var(--fg); box-shadow: 0 1px 3px rgba(13, 20, 33, 0.12); }
.views button:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }

.search { position: relative; display: flex; align-items: center; width: 260px; }
.search svg { position: absolute; left: 12px; width: 16px; height: 16px; fill: var(--muted); pointer-events: none; }
.search input { width: 100%; font: inherit; font-size: 14px; color: var(--fg); background: var(--chip); border: 1px solid transparent; border-radius: 10px; padding: 9px 36px; outline: none; }
.search input::placeholder { color: var(--muted); }
.search input:focus { background: var(--bg); border-color: #cfd6e4; }
.search input::-webkit-search-cancel-button { display: none; }
.search kbd { position: absolute; right: 10px; font: inherit; font-size: 12px; color: var(--muted); border: 1px solid #d5dbe5; border-radius: 6px; padding: 0 6px; line-height: 18px; }

.period { display: inline-flex; align-items: center; gap: 8px; }
.side { font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 4px 2px; cursor: pointer; display: inline-flex; align-items: baseline; gap: 6px; }
.side.on { color: var(--fg); }
.side:disabled { cursor: default; color: #a6b0c3; }
.pct { font-size: 12px; font-weight: 600; color: var(--bull); }
.toggle { position: relative; width: 40px; height: 22px; border-radius: 11px; border: 0; padding: 0; background: #cfd6e4; cursor: pointer; transition: background 0.2s; }
.toggle[aria-checked="true"] { background: var(--fg); }
.toggle:disabled { cursor: default; background: var(--chip); }
.knob { position: absolute; top: 3px; left: 3px; width: 16px; height: 16px; border-radius: 50%; background: #fff; box-shadow: 0 1px 2px rgba(13, 20, 33, 0.2); transition: transform 0.2s; }
.toggle[aria-checked="true"] .knob { transform: translateX(18px); }
.toggle:focus-visible, .side:focus-visible { outline: 2px solid var(--fg); outline-offset: 2px; }
@media (prefers-reduced-motion: reduce) { .toggle, .knob { transition: none; } }

@media (max-width: 700px) {
  .toolbar { flex-direction: column; align-items: stretch; }
  .pickers { flex-direction: column; align-items: stretch; gap: 10px; }
  .period { align-self: flex-start; }
  .views { display: flex; }
  .views button { flex: 1; padding-inline: 6px; }
  .search { width: 100%; }
  .search kbd { display: none; }
}
</style>
