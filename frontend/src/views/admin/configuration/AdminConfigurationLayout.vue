<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import PageHeader from '../../../components/PageHeader.vue'
import '../../../styles/configuration.css'

const route = useRoute()
const router = useRouter()

const tabs = [
  { label: 'Identity', to: '/admin/configuration/identity', icon: 'pi pi-key' },
  { label: 'Appearance', to: '/admin/configuration/appearance', icon: 'pi pi-palette' },
  { label: 'Harbor', to: '/admin/configuration/harbor', icon: 'pi pi-server' },
  { label: 'Trivy', to: '/admin/configuration/trivy', icon: 'pi pi-shield' },
  { label: 'Webhooks', to: '/admin/configuration/webhooks', icon: 'pi pi-send' },
  { label: 'Email', to: '/admin/configuration/email', icon: 'pi pi-envelope' },
  { label: 'Teams', to: '/admin/configuration/teams', icon: 'pi pi-microsoft' },
]

function isActive(path: string) {
  return route.path === path
}
</script>

<template>
  <div class="page">
    <PageHeader
      title="Configuration"
      subtitle="Integrations, appearance, and notification delivery."
    />

    <nav class="config-tabs" aria-label="Configuration sections">
      <button
        v-for="tab in tabs"
        :key="tab.to"
        type="button"
        class="config-tab"
        :class="{ active: isActive(tab.to) }"
        @click="router.push(tab.to)"
      >
        <i :class="tab.icon" />
        <span>{{ tab.label }}</span>
      </button>
    </nav>

    <div class="config-content surface-card">
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.config-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 1rem;
  padding: 0.35rem;
  background: color-mix(in srgb, var(--sb-primary) 8%, var(--sb-surface));
  border-radius: var(--sb-radius-sm);
  border: 1px solid var(--sb-border);
}

.config-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border: none;
  background: transparent;
  color: var(--sb-muted);
  font: inherit;
  font-size: 0.875rem;
  font-weight: 600;
  padding: 0.55rem 0.9rem;
  border-radius: calc(var(--sb-radius-sm) - 2px);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.config-tab:hover {
  background: var(--sb-surface);
  color: var(--sb-text);
}

.config-tab.active {
  background: var(--sb-surface);
  color: var(--sb-primary);
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
}

.config-tab i {
  font-size: 0.85rem;
}

.config-content {
  padding: 1.25rem 1.5rem;
}
</style>
