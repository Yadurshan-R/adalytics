<!-- The drawing toolbar beside the chart, laid out like Birdeye's: drawing
     tools (the line tools grouped under one button), then the ruler and zoom,
     then the magnet, stay in drawing mode, lock and hide drawings, then delete
     all. Big icons; the column scrolls when it does not fit. -->
<script setup>
import ToolGroup from './ToolGroup.vue'
import ToolIcon from './ToolIcon.vue'

const props = defineProps({
  tool: { type: String, required: true },
  magnet: { type: Boolean, default: false },
  hidden: { type: Boolean, default: false },
  locked: { type: Boolean, default: false },
  stay: { type: Boolean, default: false },
  count: { type: Number, default: 0 },
})
const emit = defineEmits(['update:tool', 'update:magnet', 'update:hidden', 'update:locked', 'update:stay', 'zoom', 'reset', 'clear'])

// LINES are grouped under one button; TOOLS are the other drawing tools.
// Each has an SVG path for its icon (20 × 20) and dots for its points.
const LINES = [
  { key: 'trend', label: 'Trend line', icon: 'M4.5 15.5 15.5 4.5', dots: [[4.5, 15.5], [15.5, 4.5]] },
  { key: 'ray', label: 'Ray', icon: 'M4.5 15.5 18 2', dots: [[4.5, 15.5], [10.5, 9.5]] },
  { key: 'extended', label: 'Extended line', icon: 'M2 18 18 2', dots: [[7, 13], [13, 7]] },
  { key: 'hline', label: 'Horizontal line', icon: 'M2 10h16', dots: [[10, 10]] },
  { key: 'vline', label: 'Vertical line', icon: 'M10 2v16', dots: [[10, 10]] },
]
const CURSOR = { key: 'cursor', label: 'Cross', icon: 'M10 2.5v5.5M10 12v5.5M2.5 10H8M12 10h5.5' }
// In the same order as TradingView's and Birdeye's bars: Fib right after the lines.
const TOOLS = [
  { key: 'fib', label: 'Fib retracement', icon: 'M5.5 4H18M2 8h16M2 12h16M2 16h12.5', dots: [[3.5, 4], [16.5, 16]] },
  { key: 'channel', label: 'Parallel channel', icon: 'M3 12 12 3M8 17 17 8', dots: [[3, 12], [12, 3], [17, 8]] },
  { key: 'rect', label: 'Rectangle', icon: 'M4 5h12v10H4z', dots: [[4, 5], [16, 15]] },
  { key: 'text', label: 'Text', icon: 'M4 5.5V3.5h12v2M10 3.5v13M7.5 16.5h5' },
]
const RULER = { key: 'ruler', label: 'Measure', icon: 'M3 14 14 3l3 3L6 17zM6.5 10.5l1.5 1.5M9 8l1.5 1.5M11.5 5.5l1.5 1.5' }

// pick chooses a tool; picking the one already on goes back to the cursor.
const pick = (key) => emit('update:tool', props.tool === key && key !== 'cursor' ? 'cursor' : key)
</script>

<template>
  <div class="toolbar" role="toolbar" aria-label="Chart tools">
    <button type="button" class="btn" :aria-pressed="tool === 'cursor'" :aria-label="CURSOR.label" :title="CURSOR.label" @click="pick('cursor')">
      <ToolIcon :tool="CURSOR" />
    </button>
    <ToolGroup :tools="LINES" :tool="tool" @pick="pick" />
    <button v-for="t in TOOLS" :key="t.key" type="button" class="btn" :aria-pressed="tool === t.key"
            :aria-label="t.label" :title="t.label" @click="pick(t.key)">
      <ToolIcon :tool="t" />
    </button>

    <span class="sep" aria-hidden="true"></span>

    <button type="button" class="btn" :aria-pressed="tool === RULER.key" :aria-label="RULER.label" :title="RULER.label" @click="pick(RULER.key)">
      <ToolIcon :tool="RULER" />
    </button>
    <button type="button" class="btn" aria-label="Zoom in" title="Zoom in" @click="emit('zoom', 0.6)">
      <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" /><path d="M12.5 12.5 17 17M6 8.5h5M8.5 6v5" /></svg>
    </button>
    <button type="button" class="btn" aria-label="Zoom out" title="Zoom out" @click="emit('zoom', 1.6)">
      <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" /><path d="M12.5 12.5 17 17M6 8.5h5" /></svg>
    </button>
    <button type="button" class="btn" aria-label="Reset view" title="Reset view" @click="emit('reset')">
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M3 7V3h4M17 7V3h-4M3 13v4h4M17 13v4h-4" /></svg>
    </button>

    <span class="sep" aria-hidden="true"></span>

    <button type="button" class="btn" :aria-pressed="magnet" aria-label="Magnet" title="Magnet" @click="emit('update:magnet', !magnet)">
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M5 3v7a5 5 0 0 0 10 0V3M5 6.5h3.5M11.5 6.5H15M8.5 3v7a1.5 1.5 0 0 0 3 0V3" /></svg>
    </button>
    <button type="button" class="btn" :aria-pressed="stay" aria-label="Stay in drawing mode" title="Stay in drawing mode"
            @click="emit('update:stay', !stay)">
      <svg viewBox="0 0 20 20" aria-hidden="true">
        <path d="M3 15l.8-3.2 7-7 2.4 2.4-7 7zM9.4 6.2l2.4 2.4" />
        <rect x="12" y="13" width="6" height="4.5" rx="1" /><path d="M13.5 13v-1.2a1.5 1.5 0 0 1 3 0V13" />
      </svg>
    </button>
    <button type="button" class="btn" :aria-pressed="locked" :aria-label="locked ? 'Unlock drawings' : 'Lock drawings'"
            :title="locked ? 'Unlock drawings' : 'Lock drawings'" @click="emit('update:locked', !locked)">
      <svg viewBox="0 0 20 20" aria-hidden="true">
        <rect x="4.5" y="9" width="11" height="8" rx="1.5" />
        <path :d="locked ? 'M7 9V6.5a3 3 0 0 1 6 0V9' : 'M7 9V6.5a3 3 0 0 1 5.8-1.1'" />
      </svg>
    </button>
    <button type="button" class="btn" :aria-pressed="hidden" aria-label="Hide drawings" title="Hide drawings"
            @click="emit('update:hidden', !hidden)">
      <svg viewBox="0 0 20 20" aria-hidden="true">
        <path d="M2 10s3-5.5 8-5.5S18 10 18 10s-3 5.5-8 5.5S2 10 2 10z" /><circle cx="10" cy="10" r="2.5" />
        <path v-if="hidden" d="M3 3l14 14" />
      </svg>
    </button>
    <span class="sep" aria-hidden="true"></span>

    <button type="button" class="btn" :disabled="!count" aria-label="Delete all drawings" title="Delete all drawings" @click="emit('clear')">
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M3.5 5.5h13M8 5.5V3.5h4v2M5.5 5.5l1 11h7l1-11M8.5 8.5v5M11.5 8.5v5" /></svg>
    </button>
  </div>
</template>

<style scoped>
.toolbar { display: flex; flex-direction: column; align-items: center; gap: 2px; padding: 6px 4px; border-right: 1px solid var(--line); overflow-y: auto; overscroll-behavior: contain; scrollbar-width: none; }
.toolbar::-webkit-scrollbar { display: none; }
.btn { flex: none; display: inline-flex; align-items: center; justify-content: center; width: 46px; height: 38px; border: 0; border-radius: 6px; background: none; color: #131722; cursor: pointer; }
.btn:hover:not(:disabled) { background: #f0f3fa; color: #131722; }
.btn[aria-pressed="true"] { color: #2962ff; }
.btn:disabled { color: #b2b5be; cursor: default; }
.btn:focus-visible { outline: 2px solid #2962ff; outline-offset: 1px; }
.btn svg { width: 30px; height: 30px; fill: none; stroke: currentColor; stroke-width: 0.9; stroke-linecap: round; stroke-linejoin: round; }
.sep { width: 32px; height: 1px; margin: 4px 0; background: var(--line); flex: none; }

@media (max-width: 700px) {
  .toolbar { flex-direction: row; overflow-x: auto; border-right: 0; border-bottom: 1px solid var(--line); scrollbar-width: none; }
  .btn { width: 40px; height: 40px; }
  .btn svg { width: 24px; height: 24px; }
  .sep { width: 1px; height: 24px; margin: 0 4px; }
}
</style>
