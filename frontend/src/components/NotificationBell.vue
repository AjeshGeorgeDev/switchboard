<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '../api'
import OverlayPanel from 'primevue/overlaypanel'
import Badge from 'primevue/badge'
import Button from 'primevue/button'

const panel = ref()
const unread = ref(0)
const items = ref<any[]>([])
let timer: number

async function load() {
  const data = await api.get<{ items: any[]; unread_count: number }>('/api/notifications?unread=true')
  items.value = data.items
  unread.value = data.unread_count
}

async function markRead(id: string) {
  await api.patch(`/api/notifications/${id}/read`, {})
  await load()
}

function toggle(e: Event) {
  panel.value.toggle(e)
  load()
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 60000)
})
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div class="bell-wrap">
    <Button icon="pi pi-bell" severity="secondary" text rounded aria-label="Notifications" @click="toggle" />
    <Badge v-if="unread" :value="unread" severity="danger" class="badge" />
    <OverlayPanel ref="panel" class="notif-panel">
      <div class="panel-title">Notifications</div>
      <div v-for="n in items" :key="n.id" class="notif" @click="markRead(n.id)">
        <strong>{{ n.title }}</strong>
        <p>{{ n.body }}</p>
      </div>
      <p v-if="!items.length" class="empty">You're all caught up.</p>
    </OverlayPanel>
  </div>
</template>

<style scoped>
.bell-wrap {
  position: relative;
}

.badge {
  position: absolute;
  top: 0;
  right: 0;
  transform: translate(25%, -25%);
}

.panel-title {
  font-weight: 700;
  margin-bottom: 0.75rem;
}

.notif {
  padding: 0.65rem 0;
  border-bottom: 1px solid var(--sb-border);
  cursor: pointer;
}

.notif:last-child {
  border-bottom: none;
}

.notif p {
  margin-top: 0.25rem;
  color: var(--sb-muted);
  font-size: 0.85rem;
}

.empty {
  color: var(--sb-muted);
  font-size: 0.875rem;
}
</style>
