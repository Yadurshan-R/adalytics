<!-- The menu button in the top bar, like Birdeye's: a ☰ icon left of the
     logo that opens a short list of the site's pages (Tokens, Large Trades).
     Who uses it: anyone moving between pages; the bar itself stays clean.
     The list closes on a pick, a click anywhere else, or Esc. -->
<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const open = ref(false)
const root = ref(null)

// PAGES are the menu's entries: name, one line on what it shows (on hover),
// and a small line icon: a list for Tokens, two arrows for trades.
const PAGES = [
  { to: '/', names: ['list', 'token'], label: 'Tokens', hint: 'Every Cardano token on Minswap',
    icon: 'M3 4h10M3 8h10M3 12h6' },
  { to: '/trades', names: ['trades'], label: 'Large Trades', hint: 'Big trades, up to 30 days back',
    icon: 'M3 5.5h9.5M10 3l2.5 2.5L10 8M13 10.5H3.5M6 8l-2.5 2.5L6 13' },
]

// outside closes the list when the visitor clicks anywhere else; onKey on Esc.
const outside = (e) => { if (open.value && !root.value?.contains(e.target)) open.value = false }
const onKey = (e) => { if (e.key === 'Escape') open.value = false }
onMounted(() => { document.addEventListener('pointerdown', outside); document.addEventListener('keydown', onKey) })
onUnmounted(() => { document.removeEventListener('pointerdown', outside); document.removeEventListener('keydown', onKey) })
</script>

<template>
  <div ref="root" class="menu">
    <button type="button" class="toggle" :aria-expanded="open" aria-label="Menu" @click="open = !open">
      <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M3 5.5h14M3 10h14M3 14.5h14" /></svg>
    </button>
    <nav v-if="open" class="list" aria-label="Pages">
      <RouterLink v-for="p in PAGES" :key="p.to" :to="p.to" :title="p.hint" class="item" :class="{ on: p.names.includes(route.name) }" @click="open = false">
        <svg viewBox="0 0 16 16" aria-hidden="true"><path :d="p.icon" /></svg>{{ p.label }}
      </RouterLink>
    </nav>
  </div>
</template>

<style scoped>
.menu { position: relative; margin-right: 12px; }
.toggle { display: inline-flex; align-items: center; justify-content: center; width: 40px; height: 40px; border: 1px solid var(--line); border-radius: 8px; background: none; color: var(--fg); cursor: pointer; }
.toggle:hover, .toggle[aria-expanded="true"] { background: var(--hover); }
.toggle svg { width: 20px; height: 20px; fill: none; stroke: currentColor; stroke-width: 1.6; stroke-linecap: round; }
.list { position: absolute; top: calc(100% + 6px); left: 0; z-index: 20; width: 200px; padding: 4px; border: 1px solid var(--line); border-radius: 8px; background: var(--bg); box-shadow: 0 8px 24px rgba(13, 20, 33, 0.12); }
.item { display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 6px; color: var(--fg); font-size: 14px; font-weight: 500; }
.item:hover { background: var(--hover); }
.item.on { color: #2962ff; font-weight: 600; }
.item svg { width: 16px; height: 16px; flex: none; fill: none; stroke: currentColor; stroke-width: 1.5; stroke-linecap: round; stroke-linejoin: round; opacity: 0.8; }
</style>
