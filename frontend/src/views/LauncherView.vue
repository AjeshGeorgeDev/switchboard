<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { onBeforeRouteUpdate } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { api } from '../api'
import CatalogGrid from '../components/CatalogGrid.vue'
import { type CatalogSection } from '../utils/catalog'

const auth = useAuthStore()
const apps = ref<any[]>([])
const sections = ref<CatalogSection[]>([])

async function loadApps() {
  const [appList, sectionList] = await Promise.all([
    api.get<any[]>('/api/catalog'),
    api.get<CatalogSection[]>('/api/catalog/sections'),
  ])
  apps.value = appList
  sections.value = sectionList
}

const clock = ref(formatTime())
let timer: ReturnType<typeof setInterval> | undefined

function formatTime() {
  return new Intl.DateTimeFormat(undefined, {
    hour: 'numeric',
    minute: '2-digit',
  }).format(new Date())
}

onMounted(() => {
  loadApps()
  timer = setInterval(() => {
    clock.value = formatTime()
  }, 30_000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

onBeforeRouteUpdate((to) => {
  if (to.path === '/launcher') loadApps()
})

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'Good morning'
  if (hour < 17) return 'Good afternoon'
  return 'Good evening'
})

const displayName = computed(() => {
  const user = auth.user
  if (!user) return 'there'
  if (user.display_name) return user.display_name
  return user.email.split('@')[0]
})
</script>

<template>
  <div class="dashboard">
    <header class="dashboard-hero">
      <div class="hero-text">
        <p class="clock">{{ clock }}</p>
        <h1>{{ greeting }}, {{ displayName }}</h1>
        <p class="tagline">Jump to your internal tools and services.</p>
      </div>
    </header>

    <CatalogGrid :apps="apps" :sections="sections" />
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 1100px;
}

.dashboard-hero {
  padding-top: 0.25rem;
}

.hero-text h1 {
  font-size: clamp(1.75rem, 4vw, 2.25rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  margin-top: 0.15rem;
}

.clock {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--sb-muted);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.tagline {
  margin-top: 0.4rem;
  color: var(--sb-muted);
  font-size: 0.95rem;
}
</style>
