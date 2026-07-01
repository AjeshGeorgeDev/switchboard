<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  message: string
  icon?: string
  compact?: boolean
}>(), {
  icon: 'pi-inbox',
  compact: false,
})
</script>

<template>
  <div
    class="mx-auto max-w-lg text-center"
    :class="compact ? 'py-10' : 'py-16'"
  >
    <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-surface-100 text-muted-color">
      <i :class="['pi', icon, 'text-xl']" />
    </div>
    <h3 class="mt-4 text-base font-semibold text-color">{{ title }}</h3>
    <p class="mt-2 text-sm leading-relaxed text-muted-color">{{ message }}</p>
  </div>
</template>
