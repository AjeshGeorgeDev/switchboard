<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import AppLayout from './components/AppLayout.vue'
import Toast from 'primevue/toast'

const route = useRoute()
const showLayout = computed(() => !['/setup', '/invite', '/'].includes(route.path))
</script>

<template>
  <Toast />
  <AppLayout v-if="showLayout">
    <RouterView />
  </AppLayout>
  <RouterView v-else />
</template>
