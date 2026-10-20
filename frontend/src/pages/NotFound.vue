<!-- The page for an address that does not exist, e.g. /nope: a short
     message and a link back to the token list, instead of a blank page. -->
<script setup>
import { RouterLink } from 'vue-router'
</script>

<template>
  <section class="missing">
    <p class="code">404</p>
    <h1>Page not found</h1>
    <p class="muted">This address does not exist on Adalytics. It may have a typo, or the page has moved.</p>
    <div class="links">
      <RouterLink to="/" class="primary">See all tokens</RouterLink>
    </div>
  </section>
</template>

<style scoped>
.missing { max-width: 420px; margin: 72px auto; text-align: center; }
.code { margin: 0; font-size: 13px; font-weight: 700; letter-spacing: 0.1em; color: #2962ff; }
h1 { margin: 8px 0 10px; font-size: 24px; letter-spacing: -0.02em; }
.muted { margin: 0; font-size: 14px; line-height: 1.5; }
.links { display: flex; justify-content: center; margin-top: 24px; }
.links a { display: inline-flex; align-items: center; height: 38px; padding: 0 16px; border-radius: 6px; font-size: 14px; font-weight: 600; }
.primary { background: #2962ff; color: #fff; }
.primary:hover { background: #1e4fd8; }
</style>
