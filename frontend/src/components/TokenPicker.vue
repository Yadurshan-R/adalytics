<!-- The Token box on the Large Trades page: a search with suggestions, like
     Birdeye's "Select a token". Closed, it shows the token picked (logo and
     ticker) or "All tokens". Click it and type: up to 8 matching tokens show
     with logo, ticker and name (the biggest pools first before anything is
     typed). Arrow keys and Enter pick one, Esc or a click elsewhere closes. -->
<script setup>
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'

const model = defineModel({ type: String, default: '' }) // a token id, or '' for all
const props = defineProps({ tokens: { type: Array, default: () => [] } }) // [{ id, ticker, name, has_logo }]

const open = ref(false)
const text = ref('')
const active = ref(0) // the suggestion the arrow keys are on
const root = ref(null)
const input = ref(null)

const picked = computed(() => props.tokens.find((t) => t.id === model.value))
// matches are the suggestions: tickers starting with the text first, then
// tickers or names containing it; "All tokens" is always first.
const matches = computed(() => {
  const q = text.value.trim().toLowerCase()
  const has = (t) => t.ticker.toLowerCase().includes(q) || t.name?.toLowerCase().includes(q)
  const list = q ? props.tokens.filter(has).sort((a, b) => b.ticker.toLowerCase().startsWith(q) - a.ticker.toLowerCase().startsWith(q)) : props.tokens
  return [{ id: '', ticker: 'All tokens' }, ...list.slice(0, 8)]
})

// show opens the search with the text box empty and focused.
async function show() {
  open.value = true
  text.value = ''
  active.value = 0
  await nextTick()
  input.value?.focus()
}
// choose picks a token (or All tokens) and closes.
function choose(t) {
  model.value = t.id
  open.value = false
}
// onKey moves through the suggestions and picks with Enter.
function onKey(e) {
  if (e.key === 'ArrowDown') active.value = Math.min(active.value + 1, matches.value.length - 1)
  else if (e.key === 'ArrowUp') active.value = Math.max(active.value - 1, 0)
  else if (e.key === 'Enter') choose(matches.value[active.value])
  else if (e.key === 'Escape') open.value = false
  else return
  e.preventDefault()
}
const outside = (e) => { if (open.value && !root.value?.contains(e.target)) open.value = false }
onMounted(() => document.addEventListener('pointerdown', outside))
onUnmounted(() => document.removeEventListener('pointerdown', outside))
</script>

<template>
  <div ref="root" class="picker">
    <button v-if="!open" type="button" class="field" aria-haspopup="listbox" @click="show">
      <template v-if="picked">
        <img v-if="picked.has_logo" :src="`/api/tokens/${picked.id}/logo`" alt="" class="logo" width="20" height="20" />
        <span v-else class="logo blank" aria-hidden="true">{{ picked.ticker.slice(0, 1) }}</span>
        <b>{{ picked.ticker }}</b>
      </template>
      <span v-else>All tokens</span>
      <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M4 6l4 4 4-4" /></svg>
    </button>
    <input v-else ref="input" v-model="text" class="field" type="search" placeholder="Search token" aria-label="Search token"
           role="combobox" aria-expanded="true" autocomplete="off" spellcheck="false" @input="active = 0" @keydown="onKey" />
    <ul v-if="open" class="list" role="listbox">
      <li v-for="(t, i) in matches" :key="t.id" role="option" :aria-selected="t.id === model" :class="{ on: i === active }"
          @pointerdown.prevent="choose(t)" @mouseenter="active = i">
        <template v-if="t.id">
          <img v-if="t.has_logo" :src="`/api/tokens/${t.id}/logo`" alt="" class="logo" width="20" height="20" loading="lazy" />
          <span v-else class="logo blank" aria-hidden="true">{{ t.ticker.slice(0, 1) }}</span>
          <b>{{ t.ticker }}</b><span class="name">{{ t.name }}</span>
        </template>
        <span v-else class="all">All tokens</span>
      </li>
      <li v-if="matches.length === 1 && text" class="none">No token matches "{{ text }}"</li>
    </ul>
  </div>
</template>

<style scoped>
.picker { position: relative; }
.field { display: flex; align-items: center; gap: 8px; width: 100%; height: 34px; padding: 0 10px; border: 1px solid var(--line); border-radius: 6px; background: var(--bg, #fff); color: var(--fg); font: inherit; font-size: 14px; text-align: left; cursor: pointer; }
input.field { cursor: text; outline: 2px solid #2962ff33; border-color: #2962ff; }
.field b { font-weight: 600; }
.field svg { width: 14px; height: 14px; margin-left: auto; fill: none; stroke: currentColor; stroke-width: 1.6; stroke-linecap: round; stroke-linejoin: round; }
.list { position: absolute; top: calc(100% + 4px); left: 0; right: 0; z-index: 20; margin: 0; padding: 4px; list-style: none; border: 1px solid var(--line); border-radius: 8px; background: var(--bg, #fff); box-shadow: 0 8px 24px rgba(13, 20, 33, 0.12); }
li { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; font-size: 13px; cursor: pointer; white-space: nowrap; overflow: hidden; }
li.on { background: var(--hover); }
li[aria-selected="true"] b, li[aria-selected="true"] .all { color: #2962ff; }
li b { font-weight: 600; }
.name { overflow: hidden; text-overflow: ellipsis; color: var(--muted); font-size: 12px; }
.none { color: var(--muted); font-size: 13px; cursor: default; }
.logo { width: 20px; height: 20px; flex: none; border-radius: 50%; object-fit: cover; }
.blank { display: inline-flex; align-items: center; justify-content: center; background: var(--chip); color: var(--muted); font-size: 10px; font-weight: 700; }
</style>
