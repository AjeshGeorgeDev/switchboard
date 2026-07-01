<script setup lang="ts">
import Tag from 'primevue/tag'

defineProps<{ severity: string }>()

const map: Record<string, string> = {
  critical: 'danger',
  high: 'warn',
  medium: 'info',
  low: 'success',
  unknown: 'secondary',
}
</script>

<template>
  <Tag :value="severity" :severity="map[severity] || 'secondary'" />
</template>
