<!-- The icon of a drawing tool: an SVG path on a 20 × 20 grid, with small
     circles for the points the visitor clicks. -->
<script setup>
defineProps({ tool: { type: Object, required: true } }) // { icon, dots }
</script>

<template>
  <svg class="icon" viewBox="0 0 20 20" aria-hidden="true">
    <path :d="tool.icon" />
    <circle v-for="(d, i) in tool.dots || []" :key="i" :cx="d[0]" :cy="d[1]" r="1.6" class="dot" />
  </svg>
</template>

<style scoped>
.icon { width: 30px; height: 30px; flex: none; fill: none; stroke: currentColor; stroke-width: 0.9; stroke-linecap: round; stroke-linejoin: round; }
.dot { fill: #fff; }
</style>
