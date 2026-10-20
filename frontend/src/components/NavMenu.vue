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

// PAGES are the menu's entries: name, one line on what it shows, the icon path.
const PAGES = [
  { to: '/', names: ['list', 'token'], label: 'Tokens', hint: 'Every Cardano token on Minswap',
    icon: 'M3 4h10M3 8h10M3 12h6' },
  { to: '/trades', names: ['trades'], label: 'Large Trades', hint: 'Big trades, up to 30 days back',
    icon: 'M2 9c1.5-3 4-4 7-4 2 0 3.5 1 4.5 2.5L15 6v4l-1.5-1C12.5 11 10.5 12 8 12c-3 0-5-1-6-3Z' },
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
      <RouterLink v-for="p in PAGES" :key="p.to" :to="p.to" class="item" :class="{ on: p.names.includes(route.name) }" @click="open = false">
        <span class="badge"><svg viewBox="0 0 16 16" aria-hidden="true"><path :d="p.icon" /></svg></span>
        <span><b>{{ p.label }}</b><small>{{ p.hint }}</small></span>
      </RouterLink>
    </nav>
  </div>
</template>

<style scoped>
.menu { position: relative; margin-right: 12px; }
.toggle { display: inline-flex; align-items: center; justify-content: center; width: 40px; height: 40px; border: 1px solid var(--line); border-radius: 8px; background: none; color: var(--fg); cursor: pointer; }
.toggle:hover, .toggle[aria-expanded="true"] { background: var(--hover); }
.toggle svg { width: 20px; height: 20px; fill: none; stroke: currentColor; stroke-width: 1.6; stroke-linecap: round; }
.list { position: absolute; top: calc(100% + 8px); left: 0; z-index: 20; width: 280px; padding: 6px; border: 1px solid var(--line); border-radius: 10px; background: var(--bg); box-shadow: 0 8px 24px rgba(13, 20, 33, 0.12); }
.item { display: flex; align-items: center; gap: 12px; padding: 10px; border-radius: 8px; color: var(--fg); }
.item:hover { background: var(--hover); }
.item.on { background: #2962ff0d; }
.item.on b { color: #2962ff; }
.item b { display: block; font-size: 14px; font-weight: 600; }
.item small { display: block; margin-top: 2px; font-size: 12px; color: var(--muted); }
.badge { display: inline-flex; align-items: center; justify-content: center; flex: none; width: 34px; height: 34px; border-radius: 8px; background: var(--chip); }
.badge svg { width: 16px; height: 16px; fill: none; stroke: var(--fg); stroke-width: 1.5; stroke-linecap: round; stroke-linejoin: round; }
.item.on .badge svg { stroke: #2962ff; }
</style>
