<!-- The candle size picker, like TradingView's: four quick buttons and a list
     with every size. A size whose history is still loading shows how far it
     has got, so the visitor sees more candles are on the way. -->
<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'

const props = defineProps({
  modelValue: { type: String, required: true }, // the size picked now
  history: { type: Object, required: true }, // { days, loaded_hours, done } from the backend
})
const emit = defineEmits(['update:modelValue'])

// SIZES is every candle size and how many days back it goes (as the backend).
const SIZES = [
  { key: '5m', label: '5 minutes', days: 2 },
  { key: '15m', label: '15 minutes', days: 7 },
  { key: '30m', label: '30 minutes', days: 7 },
  { key: '1h', label: '1 hour', days: 30 },
  { key: '2h', label: '2 hours', days: 30 },
  { key: '4h', label: '4 hours', days: 30 },
  { key: '12h', label: '12 hours', days: 30 },
  { key: '1D', label: '1 day', days: 30 },
]
const QUICK = ['15m', '1h', '4h', '1D']

// buttons are the quick sizes, plus the picked size when it came from the list.
const buttons = computed(() => (QUICK.includes(props.modelValue) ? QUICK : [...QUICK, props.modelValue]))

// progress is how much of a size's history is loaded, 0 to 100.
function progress(key) {
  const h = props.history
  if (h.done) return 100
  const days = Math.min(SIZES.find((s) => s.key === key).days, h.days || 30)
  return Math.min(100, Math.floor(((h.loaded_hours || 0) / (days * 24)) * 100))
}

const open = ref(false)
const root = ref(null)
// outside closes the list when the visitor clicks anywhere else.
const outside = (e) => !root.value?.contains(e.target) && close()
function close() {
  open.value = false
  document.removeEventListener('pointerdown', outside)
}
function toggle() {
  if (open.value) return close()
  open.value = true
  setTimeout(() => document.addEventListener('pointerdown', outside), 0)
}
// pick chooses a size and closes the list.
function pick(key) {
  emit('update:modelValue', key)
  close()
}
onBeforeUnmount(close)
</script>

<template>
  <div ref="root" class="picker">
    <div class="pick" role="group" aria-label="Candle size">
      <button v-for="k in buttons" :key="k" type="button" :aria-pressed="k === modelValue" @click="pick(k)">
        {{ k }}<span v-if="progress(k) < 100" class="pct num">{{ progress(k) }}%</span>
      </button>
      <button type="button" class="more" :aria-expanded="open" aria-label="All candle sizes" title="All candle sizes" @click="toggle">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M5.5 8l4.5 4.5L14.5 8" /></svg>
      </button>
    </div>
    <div v-if="open" class="menu" role="menu">
      <button v-for="s in SIZES" :key="s.key" type="button" role="menuitemradio" class="item"
              :aria-checked="s.key === modelValue" @click="pick(s.key)">
        <span class="name">{{ s.label }}</span>
        <span v-if="progress(s.key) < 100" class="pct num">{{ progress(s.key) }}%</span>
        <svg v-else-if="s.key === modelValue" class="tick" viewBox="0 0 20 20" aria-hidden="true"><path d="M4 10.5l4 4 8-9" /></svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.picker { position: relative; }
.pick { display: inline-flex; background: var(--chip); border-radius: 10px; padding: 4px; gap: 2px; }
.pick button { display: inline-flex; align-items: baseline; gap: 4px; font: inherit; font-size: 14px; font-weight: 600; color: var(--muted); background: none; border: 0; padding: 6px 12px; border-radius: 8px; cursor: pointer; }
.pick button[aria-pressed="true"] { background: var(--bg); color: var(--fg); box-shadow: 0 1px 3px rgba(13, 20, 33, 0.12); }
.pick button:focus-visible { outline: 2px solid var(--fg); outline-offset: 1px; }
.pick .more { padding: 6px 6px; align-items: center; }
.more:hover, .more[aria-expanded="true"] { color: var(--fg); }
.more svg { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.pct { font-size: 11px; font-weight: 600; color: var(--bull); }
.menu { position: absolute; z-index: 20; top: calc(100% + 6px); right: 0; min-width: 190px; padding: 6px 0; background: #fff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(19, 23, 34, 0.06), 0 8px 24px rgba(0, 0, 0, 0.12); }
.item { display: flex; align-items: center; gap: 10px; width: 100%; padding: 8px 14px; border: 0; background: none; color: #131722; font: 14px Inter, system-ui, sans-serif; text-align: left; cursor: pointer; }
.item:hover { background: #f0f3fa; }
.item[aria-checked="true"] { color: #2962ff; font-weight: 600; }
.name { flex: 1; }
.tick { width: 18px; height: 18px; fill: none; stroke: #2962ff; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
@media (max-width: 700px) { .menu { right: auto; left: 0; } }
</style>
