<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Select from 'primevue/select'
import Button from 'primevue/button'

const toast = useToast()
const items = ref<any[]>([])
const total = ref(0)
const action = ref('')
const resourceType = ref('')
const page = ref(0)
const rows = 30

const actionOptions = [
  { label: 'All actions', value: '' },
  { label: 'Login', value: 'auth.login' },
  { label: 'Login failed', value: 'auth.login_failed' },
  { label: 'Logout', value: 'auth.logout' },
  { label: 'User update', value: 'user.update' },
  { label: 'User roles', value: 'user.roles_update' },
  { label: 'Force logout', value: 'user.force_logout' },
  { label: 'Role create', value: 'role.create' },
  { label: 'Role update', value: 'role.update' },
  { label: 'Role delete', value: 'role.delete' },
  { label: 'OIDC provider create', value: 'oidc_provider.create' },
  { label: 'OIDC provider update', value: 'oidc_provider.update' },
  { label: 'OIDC provider delete', value: 'oidc_provider.delete' },
]

const resourceOptions = [
  { label: 'All resources', value: '' },
  { label: 'Auth', value: 'auth' },
  { label: 'User', value: 'user' },
  { label: 'Role', value: 'role' },
  { label: 'OIDC provider', value: 'oidc_provider' },
]

onMounted(load)

async function load() {
  try {
    const params = new URLSearchParams({ limit: String(rows), offset: String(page.value * rows) })
    if (action.value) params.set('action', action.value)
    if (resourceType.value) params.set('resource_type', resourceType.value)
    const data = await api.get<any>(`/api/admin/audit-logs?${params}`)
    items.value = data.items
    total.value = data.total
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Failed to load audit log', detail: e instanceof Error ? e.message : 'Unknown error', life: 5000 })
  }
}

function onPage(e: { page: number }) {
  page.value = e.page
  load()
}

function formatDate(v: string) {
  if (!v) return '—'
  return new Date(v).toLocaleString()
}
</script>

<template>
  <div class="page">
    <PageHeader title="Audit Log" subtitle="Admin actions and authentication events." />

    <div class="toolbar surface-card">
      <Select v-model="action" :options="actionOptions" option-label="label" option-value="value" class="filter" @change="() => { page = 0; load() }" />
      <Select v-model="resourceType" :options="resourceOptions" option-label="label" option-value="value" class="filter" @change="() => { page = 0; load() }" />
      <Button label="Refresh" icon="pi pi-refresh" text @click="load" />
      <span class="total">{{ total }} entries</span>
    </div>

    <DataTable
      :value="items"
      :lazy="true"
      :paginator="true"
      :rows="rows"
      :total-records="total"
      class="surface-card table-card"
      @page="onPage"
    >
      <Column field="created_at" header="Time">
        <template #body="{ data }">{{ formatDate(data.created_at) }}</template>
      </Column>
      <Column field="actor_username" header="Actor" />
      <Column field="action" header="Action" />
      <Column field="resource_type" header="Resource" />
      <Column field="resource_id" header="Resource ID" />
      <Column field="ip_address" header="IP" />
      <template #empty>
        <TableEmptyState title="No audit entries" message="Admin and auth events will appear here." icon="pi-history" />
      </template>
    </DataTable>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.9rem 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}
.filter { min-width: 200px; }
.total { margin-left: auto; color: var(--sb-muted); font-size: 0.875rem; font-weight: 600; }
.table-card { overflow: hidden; }
</style>
