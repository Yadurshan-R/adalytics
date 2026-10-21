<!-- A small dropdown for the Time and Size filters on Large Trades. The
     browser's own list draws big system-styled rows; this one matches the
     token search: compact rows, the choice's note in grey on the right
     (e.g. "Big  ₳10K–₳50K"), a tick on the one picked, and grey rows that
     cannot be picked yet (e.g. "Last 30 days  Loading 35%"). Arrow keys and
     Enter pick, Esc or a click elsewhere closes. -->
<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

const model = defineModel({ type: [String, Number], required: true })
const props = defineProps({
  options: { type: Array, required: true }, // [{ value, label, note?, disabled? }]
  label: { type: String, required: true }, // for screen readers, e.g. "Time"
})

const open = ref(false)
const active = ref(0) // the row the arrow keys are on
const root = ref(null)
const picked = computed(() => props.options.find((o) => o.value === model.value) ?? props.options[0])

// toggle opens the list on the row picked, or closes it.
function toggle() {
  open.value = !open.value
  active.value = Math.max(0, props.options.indexOf(picked.value))
}
// choose picks a row that can be picked, and closes.
function choose(o) {
  if (o.disabled) return
  model.value = o.value
  open.value = false
}
// onKey moves through the rows (skipping grey ones) and picks with Enter.
function onKey(e) {
  if (!open.value && ['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(e.key)) { toggle(); e.preventDefault(); return }
  const step = e.key === 'ArrowDown' ? 1 : e.key === 'ArrowUp' ? -1 : 0
  if (step) {
    let i = active.value
    do i = (i + step + props.options.length) % props.options.length
    while (props.options[i].disabled && i !== active.value)
    active.value = i
  } else if (e.key === 'Enter') choose(props.options[active.value])
  else if (e.key === 'Escape') open.value = false
  else return
  e.preventDefault()
}
const outside = (e) => { if (open.value && !root.value?.contains(e.target)) open.value = false }
onMounted(() => document.addEventListener('pointerdown', outside))
onUnmounted(() => document.removeEventListener('pointerdown', outside))
</script>

<template>
  <div ref="root" class="pick">
    <button type="button" class="field" aria-haspopup="listbox" :aria-expanded="open" :aria-label="label" @click="toggle" @keydown="onKey">
      <span>{{ picked.label }}</span>
      <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M4 6l4 4 4-4" /></svg>
    </button>
    <ul v-if="open" class="list" role="listbox" :aria-label="label">
      <li v-for="(o, i) in options" :key="o.value" role="option" :aria-selected="o.value === model" :aria-disabled="!!o.disabled"
          :class="{ on: i === active && !o.disabled, off: o.disabled }" @pointerdown.prevent="choose(o)" @mouseenter="active = i">
        <span class="text">{{ o.label }}</span>
        <span v-if="o.note" class="note">{{ o.note }}</span>
        <svg v-if="o.value === model" class="tick" viewBox="0 0 16 16" aria-hidden="true"><path d="M3.5 8.5l3 3 6-7" /></svg>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.pick { position: relative; }
.field { display: flex; align-items: center; width: 100%; height: 34px; padding: 0 10px; border: 1px solid var(--line); border-radius: 6px; background: var(--bg, #fff); color: var(--fg); font: inherit; font-size: 14px; text-align: left; cursor: pointer; }
.field:focus-visible, .field[aria-expanded="true"] { outline: 2px solid #2962ff33; border-color: #2962ff; }
.field svg { width: 14px; height: 14px; margin-left: auto; fill: none; stroke: currentColor; stroke-width: 1.6; stroke-linecap: round; stroke-linejoin: round; }
.list { position: absolute; top: calc(100% + 4px); left: 0; right: 0; z-index: 20; margin: 0; padding: 4px; list-style: none; border: 1px solid var(--line); border-radius: 8px; background: var(--bg, #fff); box-shadow: 0 8px 24px rgba(13, 20, 33, 0.12); }
li { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; font-size: 13px; cursor: pointer; white-space: nowrap; }
li.on { background: var(--hover); }
li[aria-selected="true"] .text { color: #2962ff; font-weight: 600; }
li.off { color: var(--muted); opacity: 0.6; cursor: default; }
.note { margin-left: auto; color: var(--muted); font-size: 12px; }
.tick { width: 14px; height: 14px; flex: none; fill: none; stroke: #2962ff; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.note + .tick { margin-left: 0; }
.text + .tick { margin-left: auto; }
</style>
