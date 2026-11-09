<!-- The candle size picker at the start of the chart bar, like Birdeye's:
     quick buttons and a list with every size. Every button is a candle size:
     1W and 1M are week and month candles (Monday and the 1st, UTC). -->
<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'

const props = defineProps({
  modelValue: { type: String, required: true }, // the size picked now
})
const emit = defineEmits(['update:modelValue'])

// SIZES is every candle size.
const SIZES = [
  { key: '5m', label: '5 minutes' },
  { key: '15m', label: '15 minutes' },
  { key: '30m', label: '30 minutes' },
  { key: '1h', label: '1 hour' },
  { key: '2h', label: '2 hours' },
  { key: '4h', label: '4 hours' },
  { key: '12h', label: '12 hours' },
  { key: '1D', label: '1 day' },
  { key: '1W', label: '1 week' },
  { key: '1M', label: '1 month' },
]
const QUICK = ['5m', '15m', '1h', '4h', '1D', '1W', '1M']

// buttons are the quick sizes, plus the picked size when it came from the list.
const buttons = computed(() => (QUICK.includes(props.modelValue) ? QUICK : [...QUICK, props.modelValue]))

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
        {{ k }}
      </button>
      <button type="button" class="more" :aria-expanded="open" aria-label="All candle sizes" title="All candle sizes" @click="toggle">
        <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M5.5 8l4.5 4.5L14.5 8" /></svg>
      </button>
    </div>
    <div v-if="open" class="menu" role="menu">
      <button v-for="s in SIZES" :key="s.key" type="button" role="menuitemradio" class="item"
              :aria-checked="s.key === modelValue" @click="pick(s.key)">
        <span class="name">{{ s.label }}</span>
        <svg v-if="s.key === modelValue" class="tick" viewBox="0 0 20 20" aria-hidden="true"><path d="M4 10.5l4 4 8-9" /></svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.picker { position: relative; }
.pick { display: inline-flex; align-items: center; gap: 2px; }
.pick button { display: inline-flex; align-items: baseline; gap: 3px; height: 30px; line-height: 30px; font: inherit; font-size: 13px; font-weight: 500; color: var(--subtle); background: none; border: 0; padding: 0 8px; border-radius: 4px; cursor: pointer; }
.pick button:hover { background: var(--btn-hover); color: var(--text); }
.pick button[aria-pressed="true"] { color: var(--accent); }
.pick button:focus-visible { outline: 2px solid var(--accent); outline-offset: -2px; }
.pick .more { padding: 0 4px; align-items: center; }
.more[aria-expanded="true"] { background: var(--btn-hover); color: var(--text); }
.more svg { width: 16px; height: 16px; fill: none; stroke: currentColor; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.menu { position: absolute; z-index: 20; top: calc(100% + 6px); left: 0; min-width: 190px; padding: 6px 0; background: var(--popup); border-radius: 8px; box-shadow: var(--popup-shadow); }
.item { display: flex; align-items: center; gap: 10px; width: 100%; padding: 8px 14px; border: 0; background: none; color: var(--text); font: 14px Inter, system-ui, sans-serif; text-align: left; cursor: pointer; }
.item:hover { background: var(--btn-hover); }
.item[aria-checked="true"] { color: var(--accent); font-weight: 600; }
.name { flex: 1; }
.tick { width: 18px; height: 18px; fill: none; stroke: var(--accent); stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
</style>
