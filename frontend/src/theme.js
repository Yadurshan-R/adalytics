// theme.js keeps the light / dark theme. The choice is saved in the browser,
// so it stays the same on the next visit. A first visit follows the computer's
// own setting (dark if the system is dark). index.html applies the saved theme
// before the page draws, so it never flashes white.
import { ref } from 'vue'

const KEY = 'adalytics-theme'

// theme is 'light' or 'dark'. Components watch it (the chart redraws with it).
export const theme = ref(document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light')

// setTheme switches the whole site to 'light' or 'dark' and remembers it.
export function setTheme(next) {
  theme.value = next === 'dark' ? 'dark' : 'light'
  document.documentElement.dataset.theme = theme.value
  try { localStorage.setItem(KEY, theme.value) } catch { /* private mode: not saved, still works */ }
}

// toggleTheme switches between light and dark.
export const toggleTheme = () => setTheme(theme.value === 'dark' ? 'light' : 'dark')

// color reads one theme colour from style.css, e.g. color('--chart-bg'), for
// the parts drawn in JavaScript (the chart and the drawings on it).
export const color = (name) => getComputedStyle(document.documentElement).getPropertyValue(name).trim()
