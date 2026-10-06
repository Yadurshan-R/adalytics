<script setup>
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <header class="topbar">
    <div class="container bar">
      <RouterLink to="/" class="brand">
        <img src="/logo.png" alt="" class="logo" />Adalytics
      </RouterLink>
    </div>
  </header>
  <main class="container page">
    <RouterView />
  </main>
</template>

<style scoped>
.topbar { border-bottom: 1px solid var(--line); }
.bar { display: flex; align-items: center; min-height: 64px; }
.brand { display: inline-flex; align-items: center; gap: 8px; font-weight: 700; font-size: 20px; letter-spacing: -0.02em; }
.logo { width: 34px; height: 34px; }
.page { padding-block: 28px 48px; }
</style>
