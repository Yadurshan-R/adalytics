<!-- The menu a toolbar button opens, laid out like Birdeye's (TradingView's):
     section headings in small capitals (FIBONACCI, PROJECTION…), then one row
     per item with its icon and name. The item that is on (the tool picked,
     the magnet in use) is drawn inverted. It opens beside the button (under
     it on a phone) and closes on a click anywhere else. -->
<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import ToolIcon from './ToolIcon.vue'

defineProps({
  // [{ title, items: [{ key, label, icon: { icon, dots }, on, disabled }] }]
  sections: { type: Array, required: true },
  at: { type: Object, required: true }, // { left, top } on screen
  label: { type: String, default: '' }, // what the menu is, for screen readers
})
const emit = defineEmits(['choose', 'close'])

// outside closes the menu on a click that is not in it; the button that
// opened it handles its own clicks (pressing it again closes it).
const outside = (e) => !e.target.closest('.tool-menu, .tool-more') && emit('close')
const onKey = (e) => e.key === 'Escape' && emit('close')

onMounted(() => {
  document.addEventListener('pointerdown', outside)
  document.addEventListener('keydown', onKey)
})
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', outside)
  document.removeEventListener('keydown', onKey)
})
</script>

<template>
  <Teleport to="body">
    <div class="tool-menu" role="menu" :aria-label="label" :style="{ left: at.left + 'px', top: at.top + 'px' }">
      <template v-for="(s, i) in sections" :key="i">
        <hr v-if="i > 0" class="rule" />
        <p v-if="s.title" class="title">{{ s.title }}</p>
        <button v-for="item in s.items" :key="item.key" type="button" role="menuitem" class="item"
                :class="{ on: item.on }" :disabled="item.disabled" @click="emit('choose', item.key)">
          <ToolIcon :tool="item.icon" />
          <span>{{ item.label }}</span>
        </button>
      </template>
    </div>
  </Teleport>
</template>

<style scoped>
.tool-menu { position: fixed; z-index: 50; min-width: 240px; max-height: calc(100vh - 16px); overflow-y: auto; padding: 6px 0; background: var(--popup); border-radius: 8px; box-shadow: var(--popup-shadow); }
.title { margin: 8px 18px 4px; color: var(--muted); font: 500 12px/1.4 Inter, system-ui, sans-serif; letter-spacing: 0.06em; text-transform: uppercase; }
.rule { margin: 6px 0; border: 0; border-top: 1px solid var(--line); }
.item { display: flex; align-items: center; gap: 12px; width: 100%; padding: 5px 18px 5px 12px; border: 0; background: none; color: var(--text); font: 15px Inter, system-ui, sans-serif; text-align: left; cursor: pointer; }
.item:hover:not(:disabled) { background: var(--btn-hover); }
.item.on { background: var(--tool-on-bg); color: var(--tool-on-fg); }
.item:disabled { color: var(--disabled); cursor: default; }
.item:focus-visible { outline: 2px solid var(--accent); outline-offset: -2px; }
</style>
