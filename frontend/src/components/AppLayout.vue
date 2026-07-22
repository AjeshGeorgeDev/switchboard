<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import NotificationBell from './NotificationBell.vue'
import SwitchboardLogo from './SwitchboardLogo.vue'
import Button from 'primevue/button'
import Avatar from 'primevue/avatar'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const primaryNav = computed(() => {
  const items = [{ label: 'My Apps', to: '/launcher', icon: 'pi pi-th-large' }]
  if (auth.canViewSecurity()) {
    items.push(
      { label: 'Overview', to: '/security', icon: 'pi pi-chart-bar' },
      { label: 'CVEs', to: '/security/cves', icon: 'pi pi-shield' },
      { label: 'Reports', to: '/security/reports', icon: 'pi pi-file' },
    )
  }
  items.push({ label: 'Profile', to: '/profile', icon: 'pi pi-user' })
  return items
})

const adminNav = computed(() => {
  if (!auth.isAdmin()) return []
  return [
    { label: 'Catalog', to: '/admin/catalog', icon: 'pi pi-box' },
    { label: 'Users', to: '/admin/users', icon: 'pi pi-users' },
    { label: 'Roles', to: '/admin/roles', icon: 'pi pi-id-card' },
    { label: 'Configuration', to: '/admin/configuration', icon: 'pi pi-cog' },
    { label: 'Audit Log', to: '/admin/audit', icon: 'pi pi-history' },
  ]
})

function isActive(path: string) {
  if (path === '/launcher') return route.path === '/launcher'
  if (path === '/security') return route.path === '/security'
  if (path === '/admin/configuration') return route.path.startsWith('/admin/configuration')
  return route.path.startsWith(path)
}

async function logout() {
  await auth.logout()
  router.push('/')
}
</script>

<template>
  <div class="layout">
    <aside class="side-nav">
      <nav class="nav-section">
        <p class="nav-label">Workspace</p>
        <button
          v-for="item in primaryNav"
          :key="item.to"
          type="button"
          class="nav-link"
          :class="{ active: isActive(item.to) }"
          @click="router.push(item.to)"
        >
          <i :class="item.icon" />
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <nav v-if="adminNav.length" class="nav-section">
        <p class="nav-label">Admin</p>
        <button
          v-for="item in adminNav"
          :key="item.to"
          type="button"
          class="nav-link"
          :class="{ active: isActive(item.to) }"
          @click="router.push(item.to)"
        >
          <i :class="item.icon" />
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </aside>

    <div class="main-column">
      <header class="topbar">
        <button class="brand" type="button" @click="router.push('/launcher')">
          <SwitchboardLogo />
          <span>Switchboard</span>
        </button>

        <div class="topbar-end">
          <NotificationBell v-if="auth.user" />
          <div v-if="auth.user" class="user-chip">
            <Avatar :label="(auth.user.display_name || auth.user.email).charAt(0).toUpperCase()" shape="circle" />
            <span>{{ auth.user.display_name || auth.user.email }}</span>
          </div>
          <Button v-if="auth.user" label="Logout" severity="secondary" text @click="logout" />
        </div>
      </header>

      <main class="content">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  background:
    var(--sb-bg-glow),
    var(--sb-bg);
}

.main-column {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 1.5rem;
  border-radius: 0;
  border: none;
  border-bottom: 1px solid var(--sb-border);
  background: var(--sb-topbar-bg);
  backdrop-filter: blur(12px);
  box-shadow: none;
  position: relative;
}

.topbar::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 2px;
  background: var(--sb-topbar-accent);
  opacity: 0.85;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  border: none;
  background: transparent;
  font: inherit;
  font-weight: 800;
  font-size: 1.05rem;
  color: var(--sb-text);
  cursor: pointer;
  padding: 0;
}

.topbar-end {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--sb-muted);
  font-size: 0.875rem;
  font-weight: 600;
}

.content {
  flex: 1;
  max-width: 1280px;
  width: 100%;
  margin: 0 auto;
  padding: 1.5rem 1.5rem 2.5rem;
}

.side-nav {
  width: 15rem;
  min-height: 100vh;
  position: sticky;
  top: 0;
  align-self: start;
  border-radius: 0;
  border: none;
  border-right: 1px solid var(--sb-sidebar-border);
  padding: 1.25rem 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  background: var(--sb-sidebar-bg);
  box-shadow: none;
}

.nav-section {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-label {
  margin: 0 0.65rem 0.35rem;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--sb-sidebar-muted);
  opacity: 0.9;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  width: 100%;
  border: none;
  background: transparent;
  color: var(--sb-sidebar-text);
  font: inherit;
  font-size: 0.9rem;
  font-weight: 600;
  padding: 0.65rem 0.75rem;
  border-radius: var(--sb-radius-sm);
  cursor: pointer;
  text-align: left;
  transition: background 0.15s, color 0.15s;
}

.nav-link:hover {
  background: var(--sb-sidebar-hover);
  color: var(--sb-nav-active-color);
}

.nav-link.active {
  background: var(--sb-nav-active-bg);
  color: var(--sb-nav-active-color);
  box-shadow: inset 3px 0 0 var(--sb-nav-active-bar);
}

.nav-link i {
  width: 1.1rem;
  text-align: center;
}

@media (max-width: 900px) {
  .layout {
    grid-template-columns: auto minmax(0, 1fr);
  }

  .side-nav {
    width: 4.25rem;
    padding-inline: 0.5rem;
  }

  .nav-label,
  .nav-link span {
    display: none;
  }

  .nav-link {
    justify-content: center;
    padding-inline: 0.5rem;
  }

  .user-chip span {
    display: none;
  }
}
</style>
