<!-- The size tag after a trade's value: WHALE (₳50K and more) and BIG
     (₳10K to ₳50K) as quiet outlined tags in the text colour, nothing for
     smaller trades. Plain text tags like
     Birdeye's SWAP tag and our Buy / Sell tags, so the table stays calm and
     anyone can read it without hovering. -->
<script setup>
import { computed } from 'vue'
import { sizeOf } from '../tradeSizes.js'

const props = defineProps({ ada: { type: Number, required: true } }) // the trade's ADA
const size = computed(() => sizeOf({ amount_ada: props.ada }))
</script>

<template>
  <span v-if="size.tag" class="tag" :class="size.key" :title="size.hint">{{ size.tag }}</span>
</template>

<style scoped>
.tag { display: inline-block; margin-left: 8px; padding: 1px 5px; border: 1px solid var(--line); border-radius: 4px; font-size: 10px; font-weight: 600; letter-spacing: 0.05em; vertical-align: 1px; }
.whale { border-color: #c9d1dc; color: var(--fg); }
.big { color: var(--muted); }
</style>
