<!-- The drawing toolbar beside the chart, laid out like Birdeye's
     (TradingView's): the tool groups (toolList.js) at the top, each button
     showing the last tool picked from its menu; then Measure and zoom; then
     the magnet (weak / strong), stay in drawing mode, lock and hide; then
     remove. On a phone it is a row above the chart. -->
<script setup>
import { ref, computed } from 'vue'
import ToolButton from './ToolButton.vue'
import { GROUPS, LABELS, kindsOf } from '../chart/toolList.js'
import { ICONS } from '../chart/toolIcons.js'
import { MEASURE } from '../chart/useDrawTools.js'

const props = defineProps({
  tool: { type: String, required: true }, // the tool picked
  magnet: { type: String, default: 'off' }, // 'off', 'weak' or 'strong'
  stay: { type: Boolean, default: true },
  locked: { type: Boolean, default: false },
  hidden: { type: Boolean, default: false }, // drawings hidden
  hiddenInd: { type: Boolean, default: false }, // indicators hidden
  count: { type: Number, default: 0 }, // drawings on the chart
  indCount: { type: Number, default: 0 }, // indicators on the chart
})
// pick: a button (picking the tool on again goes back to the cursor);
// arm: a menu item (always picks it).
const emit = defineEmits(['pick', 'arm', 'magnet', 'stay', 'lock', 'hide', 'remove', 'zoom', 'reset'])

// last remembers each group's last tool, which its button shows and picks.
const last = ref(Object.fromEntries(GROUPS.map((g) => [g.key, kindsOf(g)[0]])))
const shownOf = (g) => (kindsOf(g).includes(props.tool) ? props.tool : last.value[g.key])

// item is one menu row (see ToolMenu); its icon is ICONS[icon], by default its key's.
const item = (key, label, { on = false, disabled = false, icon = key } = {}) => ({ key, label, icon: ICONS[icon], on, disabled })

// menuOf is a group's menu, or null when the group has one tool only.
function menuOf(g) {
  if (kindsOf(g).length < 2) return null
  return g.sections.map((s) => ({ title: s.title, items: s.kinds.map((k) => item(k, LABELS[k], { on: k === props.tool })) }))
}

function arm(g, kind) {
  last.value[g.key] = kind
  emit('arm', kind)
}

// The magnet button switches the magnet on (in its last mode) or off, and
// shows that mode either way.
const lastMagnet = ref('weak')
const magnetShown = computed(() => (props.magnet === 'off' ? lastMagnet.value : props.magnet))
const magnetMenu = computed(() => [{ title: '', items: [
  item('weak', 'Weak magnet', { on: props.magnet === 'weak' }),
  item('strong', 'Strong magnet', { on: props.magnet === 'strong' }),
] }])
function chooseMagnet(mode) {
  lastMagnet.value = mode
  emit('magnet', mode)
}

// The hide and remove menus act on the drawings, the indicators or both.
const plural = (n, word) => `${n} ${word}${n === 1 ? '' : 's'}`
const hideMenu = computed(() => [{ title: '', items: [
  item('drawings', 'Hide drawings', { on: props.hidden, icon: 'trend-line' }),
  item('indicators', 'Hide indicators', { on: props.hiddenInd, icon: 'indicators' }),
  item('all', 'Hide all', { on: props.hidden && props.hiddenInd, icon: 'eyeOff' }),
] }])
const removeMenu = computed(() => [{ title: '', items: [
  item('drawings', `Remove ${plural(props.count, 'drawing')}`, { disabled: !props.count, icon: 'trend-line' }),
  item('indicators', `Remove ${plural(props.indCount, 'indicator')}`, { disabled: !props.indCount, icon: 'indicators' }),
  item('all', 'Remove all', { disabled: !props.count && !props.indCount, icon: 'trash' }),
] }])
</script>

<template>
  <div class="toolbar" role="toolbar" aria-label="Chart tools">
    <ToolButton v-for="g in GROUPS" :key="g.key" :icon="ICONS[shownOf(g)]" :label="LABELS[shownOf(g)]"
                :active="g.key === 'cursor' ? tool === 'cursor' : kindsOf(g).includes(tool)"
                :sections="menuOf(g)" :menu-label="g.label" @press="emit('pick', shownOf(g))" @choose="arm(g, $event)" />

    <span class="sep" aria-hidden="true"></span>
    <ToolButton :icon="ICONS.measure" label="Measure" :active="tool === MEASURE" @press="emit('pick', MEASURE)" />
    <ToolButton :icon="ICONS.zoomIn" label="Zoom in" @press="emit('zoom', 0.6)" />
    <ToolButton :icon="ICONS.zoomOut" label="Zoom out" @press="emit('zoom', 1.6)" />
    <ToolButton :icon="ICONS.reset" label="Reset view" @press="emit('reset')" />

    <span class="sep" aria-hidden="true"></span>
    <ToolButton :icon="ICONS[magnetShown]" :label="magnetShown === 'strong' ? 'Strong magnet' : 'Weak magnet'"
                :on="magnet !== 'off'" :sections="magnetMenu" menu-label="Magnets"
                @press="emit('magnet', magnet === 'off' ? lastMagnet : 'off')" @choose="chooseMagnet" />
    <ToolButton :icon="ICONS.stay" label="Stay in drawing mode" :on="stay" @press="emit('stay', !stay)" />
    <ToolButton :icon="ICONS[locked ? 'lock' : 'unlock']" :label="locked ? 'Unlock all drawings' : 'Lock all drawings'"
                :on="locked" @press="emit('lock', !locked)" />
    <ToolButton :icon="ICONS[hidden ? 'eyeOff' : 'eye']" :label="hidden ? 'Show drawings' : 'Hide drawings'" :on="hidden"
                :sections="hideMenu" menu-label="Hide options" @press="emit('hide', 'drawings')" @choose="emit('hide', $event)" />

    <span class="sep" aria-hidden="true"></span>
    <ToolButton :icon="ICONS.trash" :label="`Remove ${plural(count, 'drawing')}`" :disabled="!count"
                :sections="removeMenu" menu-label="Remove options" @press="emit('remove', 'drawings')" @choose="emit('remove', $event)" />
  </div>
</template>

<style scoped>
.toolbar { display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 8px 10px; border-right: 1px solid var(--line); overflow-y: auto; overscroll-behavior: contain; scrollbar-width: none; }
.toolbar::-webkit-scrollbar { display: none; }
.sep { width: 36px; height: 1px; margin: 6px 0; background: var(--line); flex: none; }

@media (max-width: 700px) {
  .toolbar { flex-direction: row; overflow-x: auto; padding: 4px 8px; border-right: 0; border-bottom: 1px solid var(--line); }
  .sep { width: 1px; height: 24px; margin: 0 6px; }
}
</style>
