<!-- The Trader cell in Recent Trades, like Birdeye's: the wallet's short
     stake address and a filter icon. Clicking either shows only that
     wallet's trades; clicking again shows everyone's. A trade whose wallet
     is not known shows a dash. -->
<script setup>
import { computed } from 'vue'

const props = defineProps({
  wallet: { type: String, default: '' }, // stake address, or '' when not known
  on: { type: Boolean, default: false }, // this wallet is the one picked
})
defineEmits(['pick'])

// short is the start and end of the address: stake1u82z…r2dn.
const short = computed(() => `${props.wallet.slice(0, 10)}…${props.wallet.slice(-4)}`)
</script>

<template>
  <span v-if="!wallet" class="muted">—</span>
  <button v-else type="button" class="trader" :class="{ on }" :title="on ? 'Show all wallets' : `Only ${wallet}`" @click="$emit('pick', wallet)">
    <span class="addr"><span class="head">{{ short.slice(0, 10) }}</span>{{ short.slice(10) }}</span>
    <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M2.5 3h11l-4.25 5v4.5l-2.5 1.5V8z" /></svg>
  </button>
</template>

<style scoped>
.trader { display: inline-flex; align-items: center; gap: 6px; padding: 2px 4px; margin: -2px -4px; border: 0; border-radius: 4px; background: none; color: #131722; font: 13px ui-monospace, SFMono-Regular, Menlo, monospace; cursor: pointer; }
.trader svg { width: 13px; height: 13px; fill: none; stroke: #9aa0ac; stroke-width: 1.3; stroke-linejoin: round; }
.trader:hover { background: #f0f3fa; }
.trader:hover svg, .trader.on svg { stroke: #2962ff; }
.trader.on { color: #2962ff; }
.trader:focus-visible { outline: 2px solid #2962ff; outline-offset: 1px; }
@media (max-width: 700px) { .head { display: none; } }
</style>
