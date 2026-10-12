<!-- The loading page, shown while the backend does its first load (about 3 to
     5 minutes after it starts). It shows how far it has got, how long is left,
     each step in plain words, and a live feed of what the backend is doing.
     The list opens by itself when loading finishes. -->
<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import LoadingActivity from './LoadingActivity.vue'
import { STEPS, formatDuration, n, stepCount } from '../loadingSteps.js'

const props = defineProps({
  // From /api/tokens → history.loading:
  // { percent, steps, elapsed_seconds, left_seconds, requests, activity }
  loading: { type: Object, required: true },
})

// The backend is asked for news every 3 seconds. In between, the clock keeps
// moving here, so the times count smoothly instead of jumping.
const now = ref(Date.now())
const receivedAt = ref(Date.now())
let clock = null
onMounted(() => { clock = setInterval(() => (now.value = Date.now()), 1000) })
onUnmounted(() => clearInterval(clock))
watch(() => props.loading, () => (receivedAt.value = Date.now()))

// sinceUpdate is how many seconds ago the last progress arrived.
const sinceUpdate = computed(() => Math.max(0, (now.value - receivedAt.value) / 1000))

// elapsed is how long the first load has been running.
const elapsed = computed(() => (props.loading.elapsed_seconds ?? 0) + sinceUpdate.value)

// leftText is the time-left estimate, counting down between updates.
const leftText = computed(() => {
  const left = props.loading.left_seconds
  if (!left) return 'Working out the time left…'
  const s = left - sinceUpdate.value
  return s > 5 ? `About ${formatDuration(s)} left` : 'Almost done…'
})

// activeStep is the step running now, for the heading under the percentage.
const activeStep = computed(() => props.loading.steps?.find((s) => s.state === 'active'))
</script>

<template>
  <div class="page">
    <header class="intro">
      <h1>Getting live Cardano data ready</h1>
      <p>
        Adalytics reads the last 24 hours of Minswap trades straight from the Cardano blockchain through the
        Koios API. This happens once, when the backend starts. The token list opens by itself when it is done.
      </p>
    </header>

    <section class="progress" role="progressbar" aria-label="Loading the last 24 hours"
             :aria-valuenow="loading.percent" aria-valuemin="0" aria-valuemax="100">
      <div class="pct num">{{ loading.percent }}%</div>
      <div class="now">{{ activeStep ? STEPS[activeStep.key]?.title : 'Starting' }}…</div>
      <div class="bar"><span :style="{ width: Math.max(loading.percent, 1) + '%' }"></span></div>
      <dl class="stats">
        <div><dt>Time left</dt><dd>{{ leftText }}</dd></div>
        <div><dt>Running for</dt><dd class="num">{{ formatDuration(elapsed) }}</dd></div>
        <div><dt>Koios requests</dt><dd class="num">{{ n(loading.requests) }}</dd></div>
      </dl>
    </section>

    <div class="grid">
      <section class="steps" aria-labelledby="steps-title">
        <h2 id="steps-title">Steps</h2>
        <ol>
          <li v-for="(s, i) in loading.steps" :key="s.key" :class="s.state">
            <span class="icon" aria-hidden="true">
              <svg v-if="s.state === 'done'" viewBox="0 0 16 16"><path d="M6.4 11.2 3.2 8l-1 1 4.2 4.2 7.6-7.6-1-1z" /></svg>
              <span v-else class="dot">{{ i + 1 }}</span>
            </span>
            <div class="body">
              <div class="row">
                <span class="title">{{ STEPS[s.key]?.title ?? s.key }}</span>
                <span class="count num">{{ stepCount(s) }}</span>
              </div>
              <p>{{ STEPS[s.key]?.text }}</p>
            </div>
          </li>
        </ol>
      </section>

      <LoadingActivity :items="loading.activity ?? []" />
    </div>

    <p class="foot">
      After this, 7 and 30 days of history keep loading in the background. You can use the site while that happens.
    </p>
  </div>
</template>

<style scoped>
.page { max-width: 920px; margin: 8px auto 0; display: flex; flex-direction: column; gap: 20px; }
.intro h1 { margin: 0 0 6px; font-size: 24px; letter-spacing: -0.02em; }
.intro p { margin: 0; color: var(--muted); line-height: 1.55; max-width: 680px; }

.progress { padding: 26px 22px; border-radius: 12px; background: #fafbfd; text-align: center; }
.pct { font-size: 44px; font-weight: 700; letter-spacing: -0.02em; line-height: 1; }
.now { margin: 8px 0 14px; font-weight: 600; color: var(--muted); }
.bar { height: 8px; border-radius: 4px; background: var(--chip); overflow: hidden; max-width: 520px; margin: 0 auto; }
.bar > span { display: block; height: 100%; border-radius: 4px; background: var(--bull); transition: width 0.6s ease; }
.stats { display: flex; justify-content: center; flex-wrap: wrap; gap: 8px 36px; margin: 18px 0 0; }
.stats dt { font-size: 12px; color: var(--muted); }
.stats dd { margin: 2px 0 0; font-weight: 600; }

.grid { display: grid; grid-template-columns: 1.1fr 1fr; gap: 20px; align-items: start; }
@media (max-width: 760px) { .grid { grid-template-columns: 1fr; } }

.steps h2 { margin: 0 0 10px; font-size: 13px; font-weight: 600; color: var(--muted); text-transform: uppercase; letter-spacing: 0.04em; }
.steps ol { list-style: none; margin: 0; padding: 0; }
.steps li { display: flex; gap: 12px; padding: 10px 0; color: #a6b0c3; }
.steps li.done, .steps li.active { color: var(--fg); }
.body { flex: 1; min-width: 0; }
.row { display: flex; justify-content: space-between; gap: 12px; }
.title { font-weight: 600; }
.count { color: var(--muted); }
.body p { margin: 3px 0 0; font-size: 13px; line-height: 1.45; color: var(--muted); }
.steps li.todo .body p { color: #b9c1cf; }

.icon { width: 22px; height: 22px; flex: none; display: inline-flex; align-items: center; justify-content: center; }
.icon svg { width: 18px; height: 18px; fill: var(--bull); }
.dot { width: 22px; height: 22px; border-radius: 50%; border: 2px solid #d5dbe5; font-size: 11px; font-weight: 700;
       display: inline-flex; align-items: center; justify-content: center; color: #a6b0c3; }
.active .dot { border-color: var(--bull); color: var(--bull); animation: pulse 1.2s ease-in-out infinite; }
@keyframes pulse { 50% { box-shadow: 0 0 0 4px rgba(22, 199, 132, 0.18); } }

.foot { margin: 0; font-size: 13px; color: var(--muted); text-align: center; }
@media (prefers-reduced-motion: reduce) { .bar > span { transition: none; } .active .dot { animation: none; } }
</style>
