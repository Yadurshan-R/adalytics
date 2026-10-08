<!-- The first load's live progress: one percentage and bar, and the four steps
     the backend goes through (done ✓, happening now ●, still to come ○). -->
<script setup>
const props = defineProps({
  loading: { type: Object, required: true }, // { percent, steps: [{ key, state, done, total }] }
})

const LABELS = {
  trades_found: 'Trades found',
  reading: 'Reading trades',
  quiet: 'Quiet tokens',
  details: 'Token details',
}

const n = (v) => Number(v ?? 0).toLocaleString('en-US')
function count(s) {
  if (s.state === 'todo') return ''
  if (s.key === 'trades_found') return n(s.done)
  if (s.key === 'details') return ''
  return s.total ? `${n(s.done)} / ${n(s.total)}` : n(s.done)
}
</script>

<template>
  <div class="box" role="progressbar" aria-label="Loading 24H" :aria-valuenow="loading.percent" aria-valuemin="0" aria-valuemax="100">
    <span class="title">Loading 24H</span>
    <span class="pct num">{{ loading.percent }}%</span>
    <span class="bar"><span :style="{ width: Math.max(loading.percent, 1) + '%' }"></span></span>

    <ul class="steps">
      <li v-for="s in loading.steps" :key="s.key" :class="s.state">
        <span class="icon" aria-hidden="true">
          <svg v-if="s.state === 'done'" viewBox="0 0 16 16"><path d="M6.4 11.2 3.2 8l-1 1 4.2 4.2 7.6-7.6-1-1z" /></svg>
          <span v-else class="dot"></span>
        </span>
        <span class="label">{{ LABELS[s.key] ?? s.key }}</span>
        <span class="count num">{{ count(s) }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.box { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 48px 16px; border-radius: 12px; background: #fafbfd; }
.title { font-size: 15px; font-weight: 600; color: var(--muted); }
.pct { font-size: 40px; font-weight: 700; letter-spacing: -0.02em; }
.bar { width: min(320px, 80%); height: 8px; border-radius: 4px; background: var(--chip); overflow: hidden; }
.bar > span { display: block; height: 100%; border-radius: 4px; background: var(--bull); transition: width 0.6s ease; }

.steps { list-style: none; margin: 18px 0 0; padding: 0; width: min(360px, 100%); }
.steps li { display: flex; align-items: center; gap: 12px; padding: 9px 4px; font-size: 14px; color: #a6b0c3; }
.steps li.done, .steps li.active { color: var(--fg); }
.label { flex: 1; }
.steps li.active .label { font-weight: 600; }
.count { color: var(--muted); }

.icon { width: 18px; height: 18px; display: inline-flex; align-items: center; justify-content: center; flex: none; }
.icon svg { width: 16px; height: 16px; fill: var(--bull); }
.dot { width: 9px; height: 9px; border-radius: 50%; border: 2px solid #c9d1de; }
.active .dot { border-color: var(--bull); background: var(--bull); animation: pulse 1.2s ease-in-out infinite; }
@keyframes pulse { 50% { opacity: 0.35; } }
@media (prefers-reduced-motion: reduce) { .bar > span { transition: none; } .active .dot { animation: none; } }
</style>
