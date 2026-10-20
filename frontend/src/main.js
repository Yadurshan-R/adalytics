import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import TokenList from './pages/TokenList.vue'
import TokenPage from './pages/TokenPage.vue'
import LargeTrades from './pages/LargeTrades.vue'
import NotFound from './pages/NotFound.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'list', component: TokenList, meta: { title: 'Cardano Tokens' } },
    { path: '/token/:ticker', name: 'token', component: TokenPage, props: true, meta: { title: 'Token' } },
    { path: '/trades', name: 'trades', component: LargeTrades, meta: { title: 'Large Trades' } },
    { path: '/:missing(.*)*', name: 'missing', component: NotFound, meta: { title: 'Page not found' } },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

// Each page names itself in the browser tab, e.g. "Large Trades | Adalytics",
// so several open tabs can be told apart. A token page puts its price in the
// tab instead (TokenPage).
router.afterEach((to) => { document.title = `${to.meta.title} | Adalytics` })

createApp(App).use(router).mount('#app')
