<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import TableEmptyState from '../../../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'

const toast = useToast()
const items = ref<any[]>([])
const total = ref(0)
const source = ref('')
const page = ref(0)
const rows = 20
const showPayload = ref(false)
const selected = ref<any | null>(null)
const loadingPayload = ref(false)

const sourceOptions = [
  { label: 'All sources', value: '' },
  { label: 'Harbor', value: 'harbor' },
  { label: 'Trivy', value: 'trivy' },
]

const payloadText = computed(() => {
  const raw = selected.value?.payload
  if (raw == null) return ''
  try {
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    return JSON.stringify(parsed, null, 2)
  } catch {
    return typeof raw === 'string' ? raw : JSON.stringify(raw, null, 2)
  }
})

onMounted(load)

async function load() {
  try {
    const params = new URLSearchParams({ limit: String(rows), offset: String(page.value * rows) })
    if (source.value) params.set('source', source.value)
    const data = await api.get<any>(`/api/admin/webhook-events?${params}`)
    items.value = data.items
    total.value = data.total
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load webhook events',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

function onPage(e: { page: number }) {
  page.value = e.page
  load()
}

async function openPayload(row: any) {
  selected.value = row
  showPayload.value = true
  loadingPayload.value = true
  try {
    const event = await api.get<any>(`/api/admin/webhook-events/${row.id}`)
    selected.value = event
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load payload',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    loadingPayload.value = false
  }
}

async function copyPayload() {
  const text = payloadText.value
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    toast.add({ severity: 'success', summary: 'Payload copied', life: 2000 })
  } catch {
    toast.add({ severity: 'error', summary: 'Copy failed', life: 3000 })
  }
}

const statusSeverity: Record<string, string> = {
  accepted: 'info',
  processed: 'success',
  failed: 'danger',
}
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Webhook delivery log</h2>
        <p class="config-card-lead">
          Recent Harbor and Trivy webhook deliveries for debugging integrations. Open a payload to
          inspect the full JSON.
        </p>
      </div>
      <Select
        v-model="source"
        :options="sourceOptions"
        option-label="label"
        option-value="value"
        class="filter"
        @change="
          () => {
            page = 0
            load()
          }
        "
      />
    </header>

    <DataTable
      :value="items"
      :lazy="true"
      :paginator="true"
      :rows="rows"
      :total-records="total"
      data-key="id"
      class="table-card"
      @page="onPage"
    >
      <Column field="received_at" header="Received" />
      <Column field="source" header="Source" />
      <Column field="status" header="Status">
        <template #body="{ data }">
          <Tag :value="data.status" :severity="statusSeverity[data.status] || 'secondary'" />
        </template>
      </Column>
      <Column field="error_message" header="Error" />
      <Column header="" style="width: 8rem">
        <template #body="{ data }">
          <Button
            label="Payload"
            icon="pi pi-code"
            text
            size="small"
            type="button"
            @click="openPayload(data)"
          />
        </template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No webhook events"
          message="Events appear here when Harbor or Trivy POST to your Switchboard webhook URLs."
          icon="pi-send"
        />
      </template>
    </DataTable>

    <div class="toolbar-footer">
      <Button label="Refresh" icon="pi pi-refresh" text type="button" @click="load" />
      <span class="total">{{ total }} events</span>
    </div>

    <Dialog
      v-model:visible="showPayload"
      modal
      :header="selected ? `Webhook payload · ${selected.source}` : 'Webhook payload'"
      :style="{ width: 'min(720px, 92vw)' }"
      :breakpoints="{ '960px': '95vw' }"
    >
      <div v-if="loadingPayload" class="payload-loading">Loading…</div>
      <pre v-else class="payload">{{ payloadText || '(empty)' }}</pre>
      <template #footer>
        <Button label="Copy" icon="pi pi-copy" text type="button" :disabled="!payloadText" @click="copyPayload" />
        <Button label="Close" type="button" @click="showPayload = false" />
      </template>
    </Dialog>
  </section>
</template>

<style scoped>
.filter {
  min-width: 180px;
}
.payload-loading {
  color: var(--sb-muted);
  padding: 1rem 0;
}
.payload {
  margin: 0;
  padding: 1rem;
  font-size: 0.8rem;
  line-height: 1.45;
  background: var(--sb-bg, #f8fafc);
  border: 1px solid var(--sb-border, #e2e8f0);
  border-radius: var(--sb-radius-sm, 6px);
  overflow: auto;
  max-height: min(60vh, 480px);
  white-space: pre-wrap;
  word-break: break-word;
}
.toolbar-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 0.75rem;
}
.total {
  color: var(--sb-muted);
  font-size: 0.875rem;
  font-weight: 600;
}
</style>
