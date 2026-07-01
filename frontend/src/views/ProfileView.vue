<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Checkbox from 'primevue/checkbox'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

const loginHistory = ref<any[]>([])
const prefs = ref<any[]>([])
const prefForm = ref([
  { channel: 'email', event_type: 'weekly_digest', enabled: true },
  { channel: 'email', event_type: 'deployment_report', enabled: false },
  { channel: 'teams', event_type: 'deployment_report', enabled: true },
  { channel: 'in_app', event_type: 'deployment_report', enabled: true },
])

function formatDate(value: unknown) {
  if (!value) return '—'
  const d = new Date(String(value))
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString()
}

onMounted(load)

async function load() {
  loginHistory.value = await api.get('/api/profile/login-history')
  try {
    prefs.value = await api.get('/api/profile/notification-preferences')
    if (prefs.value.length) prefForm.value = prefs.value
  } catch { /* use defaults */ }
}

async function savePrefs() {
  await api.patch('/api/profile/notification-preferences', prefForm.value)
}
</script>

<template>
  <div class="page">
    <PageHeader title="Profile" subtitle="Notification preferences and login history." />

    <section class="surface-card section">
      <h2 class="section-title">Notification Preferences</h2>
      <div v-for="(p, i) in prefForm" :key="i" class="pref-row">
        <span class="pref-label">{{ p.channel }} / {{ p.event_type }}</span>
        <Checkbox v-model="p.enabled" binary />
      </div>
      <Button label="Save Preferences" @click="savePrefs" />
    </section>

    <section>
      <h2 class="section-title">Login History</h2>
      <p class="section-lead">Recent sign-ins for your account, including ended sessions.</p>
      <DataTable :value="loginHistory" class="surface-card table-card">
        <Column header="Signed in">
          <template #body="{ data }">{{ formatDate(data.issued_at) }}</template>
        </Column>
        <Column field="ip_address" header="IP" />
        <Column field="user_agent" header="User Agent" />
        <Column header="Status">
          <template #body="{ data }">
            <Tag
              :value="data.revoked ? 'Ended' : 'Active'"
              :severity="data.revoked ? 'secondary' : 'success'"
            />
          </template>
        </Column>
        <template #empty>
          <TableEmptyState
            title="No login history"
            message="Your sign-in events will appear here after you log in."
            icon="pi-history"
          />
        </template>
      </DataTable>
    </section>
  </div>
</template>

<style scoped>
.section { padding: 1.25rem 1.5rem; margin-bottom: 1.5rem; }

.section-title {
  font-size: 1rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

.section-lead {
  color: var(--sb-muted);
  font-size: 0.875rem;
  margin: -0.5rem 0 1rem;
}

.pref-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 420px;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--sb-border);
}

.pref-row:last-of-type {
  border-bottom: none;
  margin-bottom: 1rem;
}

.pref-label {
  font-size: 0.9rem;
  color: #334155;
}

.table-card { overflow: hidden; }
</style>
