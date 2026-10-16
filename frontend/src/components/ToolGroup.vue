<!-- A toolbar button that holds several tools, like TradingView's line tools.
     The button shows the last tool used from the group; the small arrow opens
     the list of the others. -->
<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'
import ToolIcon from './ToolIcon.vue'

const props = defineProps({
  tools: { type: Array, required: true }, // [{ key, label, icon, dots }]
  tool: { type: String, required: true }, // the tool picked now
})
const emit = defineEmits(['pick'])

const last = ref(props.tools[0].key) // the tool the button shows
const open = ref(false)
const at = ref({ left: 0, top: 0 }) // where the list opens, on screen
const btn = ref(null)

const shown = computed(() => props.tools.find((t) => t.key === (props.tools.some((t) => t.key === props.tool) ? props.tool : last.value)))
const active = computed(() => props.tools.some((t) => t.key === props.tool))

// toggle opens the list beside the button (or under it on a phone).
function toggle() {
  if (open.value) return close()
  const r = btn.value.getBoundingClientRect()
  const phone = window.innerWidth <= 700
  at.value = phone ? { left: r.left, top: r.bottom + 4 } : { left: r.right + 6, top: r.top }
  open.value = true
  setTimeout(() => document.addEventListener('pointerdown', outside), 0)
}

function close() {
  open.value = false
  document.removeEventListener('pointerdown', outside)
}

// outside closes the list when the visitor clicks anywhere else.
const outside = (e) => !e.target.closest('.menu') && close()

function choose(key) {
  last.value = key
  emit('pick', key)
  close()
}

onBeforeUnmount(close)
</script>

<template>
  <div ref="btn" class="group">
    <button type="button" class="btn" :aria-pressed="active" :aria-label="shown.label" :title="shown.label" @click="emit('pick', shown.key)">
      <ToolIcon :tool="shown" />
    </button>
    <button type="button" class="more" :aria-expanded="open" aria-label="More line tools" title="More line tools" @click="toggle">
      <svg viewBox="0 0 6 10" aria-hidden="true"><path d="M1.5 2 4.5 5l-3 3" /></svg>
    </button>
    <Teleport to="body">
      <div v-if="open" class="menu" role="menu" :style="{ left: at.left + 'px', top: at.top + 'px' }">
        <button v-for="t in tools" :key="t.key" type="button" role="menuitem" class="item" :class="{ on: t.key === tool }" @click="choose(t.key)">
          <ToolIcon :tool="t" />
          <span>{{ t.label }}</span>
        </button>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.group { position: relative; display: flex; flex: none; }
.btn { display: inline-flex; align-items: center; justify-content: center; width: 46px; height: 38px; border: 0; border-radius: 6px; background: none; color: #131722; cursor: pointer; }
.btn:hover { background: #f0f3fa; color: #131722; }
.btn[aria-pressed="true"] { color: #2962ff; }
.more { position: absolute; right: -6px; top: 0; width: 10px; height: 38px; padding: 0; border: 0; border-radius: 3px; background: none; color: #b2b5be; cursor: pointer; }
.more:hover, .more[aria-expanded="true"] { color: #131722; background: #f0f3fa; }
.more svg { width: 6px; height: 10px; fill: none; stroke: currentColor; stroke-width: 1.4; stroke-linecap: round; stroke-linejoin: round; }
.btn:focus-visible, .more:focus-visible { outline: 2px solid #2962ff; outline-offset: 1px; }

.menu { position: fixed; z-index: 50; min-width: 200px; padding: 6px 0; background: #fff; border-radius: 6px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(19, 23, 34, 0.06), 0 8px 24px rgba(0, 0, 0, 0.12); }
.item { display: flex; align-items: center; gap: 10px; width: 100%; padding: 6px 14px; border: 0; background: none; color: #131722; font: 14px Inter, system-ui, sans-serif; text-align: left; cursor: pointer; }
.item:hover { background: #f0f3fa; }
.item.on { color: #2962ff; }

@media (max-width: 700px) {
  .btn { width: 40px; height: 40px; }
  .group { margin-right: 8px; }
  .more { height: 40px; right: -9px; }
}
</style>
