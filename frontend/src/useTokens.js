// Shared, auto-refreshing token list from the Go backend (GET /api/tokens).
import { ref, computed, onMounted, onUnmounted } from 'vue'

const REFRESH_MS = 60_000

const tokens = ref([])
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
    updatedAt.value = new Date()
    error.value = ''
  } catch (e) {
    error.value = e instanceof TypeError
      ? 'Cannot reach the backend. Is it running? Start it with "go run ." in the backend folder.'
      : e.message
  } finally {
    firstLoad.value = false
  }
}

// useTokens loads the list when the first page using it opens and refreshes it
// every minute while any page using it is open.
export function useTokens() {
  onMounted(() => {
    users++
    if (users === 1) {
      load()
      timer = setInterval(load, REFRESH_MS)
    }
  })
  onUnmounted(() => {
    users--
    if (users === 0) clearInterval(timer)
  })

  // The backend is still reading history from Koios.
  const loading = computed(() => firstLoad.value || (tokens.value.length > 0 && tokens.value.every((t) => t.status === 'loading')))

  return { tokens, error, updatedAt, loading, reload: load }
}
