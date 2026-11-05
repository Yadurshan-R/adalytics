<!-- The box for typing a text on the chart: it opens where the Text tool was
     placed or a text was double-clicked (useDrawTools' textBox). Enter or a
     click elsewhere keeps the text; Esc drops it. -->
<script setup>
import { ref } from 'vue'

const props = defineProps({ box: { type: Object, required: true } }) // { x, y, value }
const emit = defineEmits(['submit', 'cancel']) // submit(text)
const text = ref(props.box.value)

// vFocus puts the cursor in the box once it appears. The library places a
// text when the mouse button goes down, before the browser's own mousedown,
// which would take the focus straight back; so the box takes it a moment later.
const vFocus = { mounted: (el) => setTimeout(() => el.focus(), 0) }
</script>

<template>
  <input v-model="text" v-focus class="text-box" :style="{ left: box.x + 'px', top: box.y + 'px' }" maxlength="80"
         aria-label="Text on the chart" @keydown.enter="emit('submit', text)" @keydown.esc.stop="emit('cancel')" @blur="emit('submit', text)" />
</template>

<style scoped>
.text-box { position: absolute; z-index: 6; transform: translateY(-50%); width: 200px; font: 600 14px Inter, system-ui, sans-serif; color: var(--fg); padding: 4px 8px; border: 1px solid var(--accent); border-radius: 6px; outline: none; background: var(--popup); }
</style>
