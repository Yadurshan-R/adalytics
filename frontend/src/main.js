import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import TokenList from './pages/TokenList.vue'
import TokenPage from './pages/TokenPage.vue'
import LargeTrades from './pages/LargeTrades.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'list', component: TokenList },
    { path: '/token/:ticker', name: 'token', component: TokenPage, props: true },
    { path: '/trades', name: 'trades', component: LargeTrades },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

createApp(App).use(router).mount('#app')
