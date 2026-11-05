<!-- One button of the chart's toolbar, like Birdeye's (TradingView's):
     - a grey square behind it while its tool is picked (active);
     - drawn inverted while a setting it switches is on (on: magnet, lock…);
     - with a menu (sections), a small arrow tab on its right, shown on hover,
       opens the menu (ToolMenu). The button itself does the main action. -->
<script setup>
import { ref } from 'vue'
import ToolIcon from './ToolIcon.vue'
import ToolMenu from './ToolMenu.vue'

const props = defineProps({
  icon: { type: Object, required: true }, // { icon, dots } (toolIcons.js)
  label: { type: String, required: true }, // tooltip and screen-reader name
  active: { type: Boolean, default: false },
  on: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  sections: { type: Array, default: null }, // the menu, as ToolMenu takes it
  menuLabel: { type: String, default: '' }, // the arrow's tooltip, e.g. "Fibonacci tools"
})
const emit = defineEmits(['press', 'choose'])

const open = ref(false)
const at = ref({ left: 0, top: 0 })
const box = ref(null)

// toggle opens the menu beside the button (under it on a phone), or closes it.
function toggle() {
  if (open.value) return (open.value = false)
  const r = box.value.getBoundingClientRect()
  at.value = window.innerWidth <= 700 ? { left: r.left, top: r.bottom + 4 } : { left: r.right + 8, top: r.top }
  open.value = true
}

// choose passes a menu item on and closes the menu.
function choose(key) {
  open.value = false
  emit('choose', key)
}
</script>

<template>
  <div ref="box" class="cell" :class="{ open }">
    <button type="button" class="btn" :class="{ active, on }" :disabled="disabled" :aria-pressed="active || on"
            :aria-label="label" :title="label" @click="emit('press')">
      <ToolIcon :tool="icon" />
    </button>
    <button v-if="sections" type="button" class="tool-more" :aria-expanded="open" :aria-label="menuLabel || label"
            :title="menuLabel || label" @click="toggle">
      <svg viewBox="0 0 6 10" aria-hidden="true"><path :d="open ? 'M4.5 2 1.5 5l3 3' : 'M1.5 2 4.5 5l-3 3'" /></svg>
    </button>
    <ToolMenu v-if="open" :sections="sections" :at="at" :label="menuLabel || label" @choose="choose" @close="open = false" />
  </div>
</template>

<style scoped>
.cell { position: relative; display: flex; flex: none; }
.btn { display: inline-flex; align-items: center; justify-content: center; width: 44px; height: 40px; border: 0; border-radius: 8px; background: none; color: var(--text); cursor: pointer; }
.btn:hover:not(:disabled) { background: var(--btn-hover); }
.btn.active { background: var(--tool-active-bg); }
.btn.on { background: var(--tool-on-bg); color: var(--tool-on-fg); }
.btn:disabled { color: var(--disabled); cursor: default; }
.btn:focus-visible, .tool-more:focus-visible { outline: 2px solid var(--accent); outline-offset: 1px; }
.btn :deep(.icon) { width: 28px; height: 28px; }

/* The arrow tab: hidden until the button is hovered or its menu is open. */
.tool-more { position: absolute; right: -9px; top: 6px; width: 12px; height: 28px; padding: 0; border: 0; border-radius: 4px; background: var(--tool-active-bg); color: var(--text); cursor: pointer; opacity: 0; }
.cell:hover .tool-more, .cell.open .tool-more, .tool-more:focus-visible { opacity: 1; }
.tool-more:hover { background: var(--chip-hover); }
.tool-more svg { width: 6px; height: 10px; fill: none; stroke: currentColor; stroke-width: 1.4; stroke-linecap: round; stroke-linejoin: round; }

@media (max-width: 700px) {
  .btn { width: 40px; height: 40px; }
  .btn :deep(.icon) { width: 24px; height: 24px; }
  .tool-more { opacity: 1; right: -6px; width: 10px; background: none; }
}
</style>
