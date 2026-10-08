// Shared, auto-refreshing token list from the Go backend (GET /api/tokens).
// It also carries how much history is loaded, for the 24H / 7D / 30D buttons.
import { ref, computed, onMounted, onUnmounted } from 'vue'

const REFRESH_MS = 60_000
const REFRESH_FIRST_LOAD_MS = 3_000 // first load: the loading box moves live
const REFRESH_WHILE_LOADING_MS = 15_000 // 7D / 30D still loading

const tokens = ref([])
const history = ref({ days: 0, done: false, ranges: [], sparkline_days: 1 })
const error = ref('')
const updatedAt = ref(null)
const firstLoad = ref(true)
let timer = null
let users = 0

async function load() {
  try {
    const res = await fetch('/api/tokens', { cache: 'no-store' })
    if (!res.ok) throw new Error(`The backend answered ${res.status}`)
    const data = await res.json()
    tokens.value = data.tokens ?? []
    if (data.history) history.value = data.history
    updatedAt.value = new Date()
    error.value = ''
  } catch (e) {
    error.value = e instanceof TypeError
      ? 'Cannot reach the backend. Is it running? Start it with "./adalytics" in the backend folder.'
      : e.message
  } finally {
    firstLoad.value = false
  }
}

function schedule() {
  timer = setTimeout(async () => {
    await load()
    if (users > 0) schedule()
  }, history.value.loading || firstLoad.value ? REFRESH_FIRST_LOAD_MS : history.value.done ? REFRESH_MS : REFRESH_WHILE_LOADING_MS)
}

// useTokens loads the list when the first page using it opens and refreshes it
// while any page using it is open: every 3 seconds during the first load, every
// 15 seconds while 7D / 30D load, and every minute after that.
export function useTokens() {
  onMounted(() => {
    users++
    if (users === 1) {
      load()
      schedule()
    }
  })
  onUnmounted(() => {
    users--
    if (users === 0) clearTimeout(timer)
  })

  // The backend is still reading history from Koios.
  const loading = computed(() => firstLoad.value || (tokens.value.length > 0 && tokens.value.every((t) => t.status === 'loading')))

  return { tokens, history, error, updatedAt, loading, reload: load }
}
