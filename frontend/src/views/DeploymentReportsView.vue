<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Button from 'primevue/button'

const router = useRouter()
const items = ref<any[]>([])
const total = ref(0)
const expanded = ref<any[]>([])
const search = ref('')
const status = ref('')
const page = ref(0)
const rows = 20

const statusOptions = [
  { label: 'All statuses', value: '' },
  { label: 'Success', value: 'success' },
  { label: 'Failed', value: 'failed' },
  { label: 'Partial', value: 'partial' },
]

onMounted(load)

async function load() {
  const params = new URLSearchParams({ limit: String(rows), offset: String(page.value * rows) })
  if (search.value.trim()) params.set('search', search.value.trim())
  if (status.value) params.set('status', status.value)
  const data = await api.get<any>(`/api/security/reports?${params}`)
  items.value = data.items
  total.value = data.total
}

function onPage(e: { page: number }) {
  page.value = e.page
  load()
}

function applyFilters() {
  page.value = 0
  load()
}

function openDetail(row: any) {
  router.push(`/security/reports/${row.id}`)
}

const statusSeverity: Record<string, string> = {
  success: 'success', failed: 'danger', partial: 'warn',
}
</script>

<template>
  <div class="page">
    <PageHeader
      title="Deployment Reports"
      :subtitle="`Harbor webhook events · ${total} total`"
    />

    <div class="toolbar surface-card">
      <InputText v-model="search" placeholder="Search app or image…" class="search" @keyup.enter="applyFilters" />
      <Select v-model="status" :options="statusOptions" option-label="label" option-value="value" class="filter" @change="applyFilters" />
      <Button label="Search" icon="pi pi-search" @click="applyFilters" />
    </div>

    <DataTable
      v-model:expanded-rows="expanded"
      :value="items"
      :lazy="true"
      :paginator="true"
      :rows="rows"
      :total-records="total"
      data-key="id"
      class="surface-card table-card"
      @page="onPage"
    >
      <Column expander style="width: 3rem" />
      <Column field="app_name" header="App" />
      <Column field="image_name" header="Image" />
      <Column field="image_tag" header="Tag" />
      <Column field="status" header="Status">
        <template #body="{ data }"><Tag :value="data.status" :severity="statusSeverity[data.status]" /></template>
      </Column>
      <Column field="critical_count" header="Critical" />
      <Column field="high_count" header="High" />
      <Column field="received_at" header="Received" />
      <Column header="">
        <template #body="{ data }">
          <Button label="Detail" text size="small" @click="openDetail(data)" />
        </template>
      </Column>
      <template #expansion="{ data }">
        <pre class="payload">{{ JSON.stringify(data.raw_payload, null, 2) }}</pre>
      </template>
      <template #empty>
        <TableEmptyState
          title="No deployment reports"
          message="Reports appear when Harbor sends webhooks. Configure under Admin → Configuration → Harbor."
          icon="pi-file"
        />
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
.search { flex: 1; min-width: 200px; }
.filter { min-width: 160px; }
.table-card { overflow: hidden; }
.payload {
  margin: 0;
  padding: 1rem;
  font-size: 0.8rem;
  background: #f8fafc;
  border-radius: var(--sb-radius-sm);
  overflow-x: auto;
}
</style>
