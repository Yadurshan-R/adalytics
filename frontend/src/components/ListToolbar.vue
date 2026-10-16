<!-- The bar above the token list, like Birdeye's: Top Volume / Gainers /
     Losers as underlined tabs and the search box. Press "/" anywhere to jump to the search box, Esc to clear it. -->
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

defineProps({
  views: { type: Array, required: true }, // [{ key, label }]
  view: { type: String, required: true },
  query: { type: String, required: true },
})
const emit = defineEmits(['update:view', 'update:query'])

const search = ref(null)
// onKey makes "/" jump to the search box from anywhere on the page.
function onKey(e) {
  if (e.key === '/' && document.activeElement !== search.value && !/INPUT|TEXTAREA/.test(document.activeElement?.tagName)) {
    e.preventDefault()
    search.value?.focus()
  }
}
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="toolbar">
    <nav class="tabs" aria-label="Order the list">
      <button v-for="v in views" :key="v.key" type="button" :aria-pressed="view === v.key" @click="emit('update:view', v.key)">{{ v.label }}</button>
    </nav>
    <label class="search">
      <svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="9" cy="9" r="5.5" /><path d="M13 13l4 4" /></svg>
      <input ref="search" :value="query" type="search" placeholder="Search token or policy ID" aria-label="Search token"
             @input="emit('update:query', $event.target.value)" @keydown.esc="emit('update:query', '')" />
      <kbd>/</kbd>
    </label>
  </div>
</template>

<style scoped>
.toolbar { display: flex; align-items: center; gap: 8px 20px; flex-wrap: wrap; margin-bottom: 8px; border-bottom: 1px solid var(--line); }
.tabs { display: flex; gap: 20px; }
.tabs button { padding: 10px 0 11px; border: 0; border-bottom: 2px solid transparent; margin-bottom: -1px; background: none; color: var(--muted); font: inherit; font-size: 15px; font-weight: 600; cursor: pointer; }
.tabs button:hover { color: var(--fg); }
.tabs button[aria-pressed="true"] { color: var(--fg); border-bottom-color: #2962ff; }
button:focus-visible { outline: 2px solid #2962ff; outline-offset: -2px; }
.search { margin-left: auto; display: flex; align-items: center; gap: 8px; width: min(300px, 100%); height: 34px; padding: 0 10px; margin-bottom: 6px; border: 1px solid var(--line); border-radius: 6px; }
.search:focus-within { border-color: #2962ff; }
.search svg { width: 16px; height: 16px; flex: none; fill: none; stroke: var(--muted); stroke-width: 1.6; stroke-linecap: round; }
.search input { flex: 1; min-width: 0; border: 0; outline: none; background: none; font: inherit; font-size: 13px; color: var(--fg); }
kbd { padding: 0 6px; border: 1px solid var(--line); border-radius: 4px; color: var(--muted); font: 11px ui-monospace, monospace; }
@media (max-width: 700px) {
  .search { margin-left: 0; width: 100%; }
  .tabs { gap: 16px; }
}
</style>
