<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import TableEmptyState from '../../../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Select from 'primevue/select'
import Button from 'primevue/button'

const toast = useToast()
const items = ref<any[]>([])
const total = ref(0)
const source = ref('')
const expanded = ref<any[]>([])
const page = ref(0)
const rows = 20

const sourceOptions = [
  { label: 'All sources', value: '' },
  { label: 'Harbor', value: 'harbor' },
  { label: 'Trivy', value: 'trivy' },
]

onMounted(load)

async function load() {
  try {
    const params = new URLSearchParams({ limit: String(rows), offset: String(page.value * rows) })
    if (source.value) params.set('source', source.value)
    const data = await api.get<any>(`/api/admin/webhook-events?${params}`)
    items.value = data.items
    total.value = data.total
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Failed to load webhook events', detail: e instanceof Error ? e.message : 'Unknown error', life: 5000 })
  }
}

function onPage(e: { page: number }) {
  page.value = e.page
  load()
}

const statusSeverity: Record<string, string> = {
  accepted: 'info', processed: 'success', failed: 'danger',
}
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Webhook delivery log</h2>
        <p class="config-card-lead">
          Recent Harbor and Trivy webhook deliveries for debugging integrations. Expand a row to view the full payload.
        </p>
      </div>
      <Select
        v-model="source"
        :options="sourceOptions"
        option-label="label"
        option-value="value"
        class="filter"
        @change="() => { page = 0; load() }"
      />
    </header>

    <DataTable
      v-model:expanded-rows="expanded"
      :value="items"
      :lazy="true"
      :paginator="true"
      :rows="rows"
      :total-records="total"
      data-key="id"
      class="table-card"
      @page="onPage"
    >
      <Column expander style="width: 3rem" />
      <Column field="received_at" header="Received" />
      <Column field="source" header="Source" />
      <Column field="status" header="Status">
        <template #body="{ data }">
          <Tag :value="data.status" :severity="statusSeverity[data.status] || 'secondary'" />
        </template>
      </Column>
      <Column field="payload_preview" header="Preview" />
      <Column field="error_message" header="Error" />
      <template #expansion="{ data }">
        <pre class="payload">{{ JSON.stringify(data.payload, null, 2) }}</pre>
      </template>
      <template #empty>
        <TableEmptyState
          title="No webhook events"
          message="Events appear here when Harbor or Trivy POST to your Switchboard webhook URLs."
          icon="pi-send"
        />
      </template>
    </DataTable>

    <div class="toolbar-footer">
      <Button label="Refresh" icon="pi pi-refresh" text @click="load" />
      <span class="total">{{ total }} events</span>
    </div>
  </section>
</template>

<style scoped>
.filter { min-width: 180px; }
.payload {
  margin: 0;
  padding: 1rem;
  font-size: 0.8rem;
  background: #f8fafc;
  border-radius: var(--sb-radius-sm);
  overflow-x: auto;
  max-height: 400px;
}
.toolbar-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 0.75rem;
}
.total { color: var(--sb-muted); font-size: 0.875rem; font-weight: 600; }
</style>
