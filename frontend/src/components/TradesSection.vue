<!-- The trades part of the token page, with two tabs like Birdeye's:
     Trades (RecentTrades) and Top Traders (TopTraders). Each keeps its
     filters when the other tab is open (Top Traders is kept with KeepAlive). Clicking a wallet in Top Traders
     opens Trades with only that wallet's trades, over the same time range. -->
<script setup>
import { ref, watch } from 'vue'
import RecentTrades from './RecentTrades.vue'
import TopTraders from './TopTraders.vue'

const props = defineProps({
  id: { type: String, required: true },
  ticker: { type: String, required: true },
})
const TABS = [{ key: 'trades', label: 'Trades' }, { key: 'traders', label: 'Top Traders' }]
const tab = ref('trades')
const recent = ref(null)
watch(() => props.id, () => { tab.value = 'trades' })

// showWallet opens the Trades tab with that wallet's trades over the time
// range picked in Top Traders.
function showWallet(w, hours) {
  tab.value = 'trades'
  recent.value?.showWallet(w, hours)
}
</script>

<template>
  <section class="section">
    <RecentTrades v-show="tab === 'trades'" ref="recent" :id="id" :ticker="ticker">
      <template #title><div class="tabs" role="tablist"><button v-for="t in TABS" :key="t.key" type="button" role="tab" :aria-selected="t.key === tab" @click="tab = t.key">{{ t.label }}</button></div></template>
    </RecentTrades>
    <div v-show="tab === 'traders'" class="traders">
      <KeepAlive :max="1">
        <TopTraders v-if="tab === 'traders'" :key="id" :id="id" :ticker="ticker" @wallet="showWallet">
          <template #title><div class="tabs" role="tablist"><button v-for="t in TABS" :key="t.key" type="button" role="tab" :aria-selected="t.key === tab" @click="tab = t.key">{{ t.label }}</button></div></template>
        </TopTraders>
      </KeepAlive>
    </div>
  </section>
</template>

<style scoped>
.traders { margin-top: 32px; }
.tabs { display: flex; gap: 20px; }
.tabs button { padding: 0 0 6px; border: 0; border-bottom: 2px solid transparent; background: none; color: var(--muted); font: inherit; font-size: 18px; font-weight: 700; cursor: pointer; }
.tabs button:hover { color: var(--fg); }
.tabs button[aria-selected="true"] { color: var(--fg); border-bottom-color: var(--fg); }
.tabs button:focus-visible { outline: 2px solid #2962ff; outline-offset: 2px; }
</style>
