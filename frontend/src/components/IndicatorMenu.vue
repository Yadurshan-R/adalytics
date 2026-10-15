<!-- The Indicators button in the chart bar and its list. Each indicator can be
     switched on and off; the picked ones are kept while moving between tokens. -->
<script setup>
import { ref, onBeforeUnmount } from 'vue'
import { INDICATORS } from '../chart/indicators.js'

const props = defineProps({ modelValue: { type: Array, required: true } }) // picked keys
const emit = defineEmits(['update:modelValue'])

const open = ref(false)
const root = ref(null)

// toggle switches one indicator on or off.
function toggle(key) {
  const on = props.modelValue.includes(key)
  emit('update:modelValue', on ? props.modelValue.filter((k) => k !== key) : [...props.modelValue, key])
}

// outside closes the list when the visitor clicks anywhere else.
const outside = (e) => !root.value?.contains(e.target) && close()
function close() {
  open.value = false
  document.removeEventListener('pointerdown', outside)
}
function show() {
  if (open.value) return close()
  open.value = true
  setTimeout(() => document.addEventListener('pointerdown', outside), 0)
}
onBeforeUnmount(close)
</script>

<template>
  <div ref="root" class="wrap">
    <button type="button" class="open" :aria-expanded="open" @click="show">
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M7.5 16.5c1.8 0 2-1.5 2.4-4.5l.8-6c.4-2.6 1-3.5 2.8-3.5M6 8.5h7M11.5 11l4.5 5M16 11l-4.5 5" /></svg>
      Indicators<span v-if="modelValue.length" class="n num">{{ modelValue.length }}</span>
    </button>
    <div v-if="open" class="menu" role="menu">
      <button v-for="i in INDICATORS" :key="i.key" type="button" role="menuitemcheckbox" class="item"
              :aria-checked="modelValue.includes(i.key)" @click="toggle(i.key)">
        <span class="dot" :style="{ background: i.color }"></span>
        <span class="name">{{ i.name }}</span>
        <svg v-if="modelValue.includes(i.key)" class="tick" viewBox="0 0 20 20" aria-hidden="true"><path d="M4 10.5l4 4 8-9" /></svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.wrap { position: relative; }
.open { display: inline-flex; align-items: center; gap: 6px; height: 30px; font: inherit; font-size: 13px; font-weight: 500; color: #6a6d78; background: none; border: 0; border-radius: 4px; padding: 0 8px; cursor: pointer; }
.open:hover, .open[aria-expanded="true"] { background: #f0f3fa; color: #131722; }
.open:focus-visible { outline: 2px solid #2962ff; outline-offset: -2px; }
.open svg { width: 17px; height: 17px; fill: none; stroke: currentColor; stroke-width: 1.5; stroke-linecap: round; stroke-linejoin: round; }
.n { min-width: 18px; padding: 0 5px; border-radius: 9px; background: #2962ff; color: #fff; font-size: 12px; line-height: 18px; text-align: center; }
.menu { position: absolute; z-index: 20; top: calc(100% + 6px); left: 0; min-width: 240px; padding: 6px 0; background: #fff; border-radius: 8px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(19, 23, 34, 0.06), 0 8px 24px rgba(0, 0, 0, 0.12); }
.item { display: flex; align-items: center; gap: 10px; width: 100%; padding: 8px 14px; border: 0; background: none; color: #131722; font: 14px Inter, system-ui, sans-serif; text-align: left; cursor: pointer; }
.item:hover { background: #f0f3fa; }
.dot { width: 10px; height: 10px; border-radius: 50%; flex: none; }
.name { flex: 1; }
.tick { width: 18px; height: 18px; fill: none; stroke: #2962ff; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
</style>
