<!-- The live feed on the loading page: the backend's most recent messages
     ("Read 1,500 of 3,200 transactions"), newest at the top, with the time each
     happened. It shows people the backend is working and what it is doing. -->
<script setup>
import { clockTime } from '../loadingSteps.js'

defineProps({
  items: { type: Array, default: () => [] }, // [{ time, text }], newest first
})
</script>

<template>
  <section class="feed" aria-labelledby="feed-title">
    <h2 id="feed-title">What the backend is doing</h2>
    <p v-if="!items.length" class="empty">Waiting for the first message…</p>
    <ol v-else aria-live="polite">
      <li v-for="(a, i) in items" :key="a.time + a.text" :class="{ latest: i === 0 }">
        <time class="num" :datetime="a.time">{{ clockTime(a.time) }}</time>
        <span>{{ a.text }}</span>
      </li>
    </ol>
  </section>
</template>

<style scoped>
.feed { border: 1px solid var(--line); border-radius: 12px; padding: 16px 18px; background: #fff; }
h2 { margin: 0 0 10px; font-size: 13px; font-weight: 600; color: var(--muted); text-transform: uppercase; letter-spacing: 0.04em; }
.empty { margin: 0; color: var(--muted); }
ol { list-style: none; margin: 0; padding: 0; max-height: 260px; overflow-y: auto; }
li { display: flex; gap: 12px; padding: 6px 0; font-size: 13px; line-height: 1.45; color: var(--muted); border-top: 1px solid var(--line); }
li:first-child { border-top: none; }
li.latest { color: var(--fg); font-weight: 500; animation: fade-in 0.4s ease; }
time { flex: none; color: #a6b0c3; font-size: 12px; padding-top: 1px; }
@keyframes fade-in { from { opacity: 0; transform: translateY(-3px); } }
@media (prefers-reduced-motion: reduce) { li.latest { animation: none; } }
</style>
