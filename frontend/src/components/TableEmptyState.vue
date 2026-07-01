<script setup lang="ts">
withDefaults(defineProps<{
  title?: string
  message: string
  icon?: string
}>(), {
  title: 'No data yet',
  icon: 'pi-inbox',
})
</script>

<template>
  <div class="relative overflow-hidden rounded-2xl border border-primary/10 bg-gradient-to-br from-surface-0 via-surface-0 to-primary/5 px-6 py-12 text-center shadow-sm">
    <div class="pointer-events-none absolute -right-8 -top-8 h-24 w-24 rounded-full bg-primary/10 blur-2xl" />

    <div class="relative mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary/15 to-primary/5 ring-1 ring-primary/15">
      <i :class="['pi', icon, 'text-xl text-primary']" />
    </div>
    <h3 class="relative mt-4 text-base font-bold text-color">{{ title }}</h3>
    <p class="relative mx-auto mt-2 max-w-md text-sm leading-relaxed text-muted-color">{{ message }}</p>
  </div>
</template>
