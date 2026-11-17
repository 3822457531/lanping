<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>()
</script>

<template>
  <div class="ph">
    <div>
      <h2>{{ title }}</h2>
      <p v-if="subtitle">{{ subtitle }}</p>
    </div>
    <div class="actions"><slot /></div>
  </div>
</template>

<style scoped>
.ph { display: flex; justify-content: space-between; align-items: flex-end; gap: 16px; margin-bottom: 16px; }
h2 { margin: 0; font-size: 20px; font-weight: 650; }
p { margin: 4px 0 0; color: var(--lp-muted); font-size: 13px; }
.actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
</style>
