<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import SeverityTag from '../components/SeverityTag.vue'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const items = ref<any[]>([])
const total = ref(0)
const summary = ref<any>({})
const severity = ref('')
const search = ref(typeof route.query.search === 'string' ? route.query.search : '')
const page = ref(0)
const rows = 20
const exporting = ref(false)

const severityOptions = [
  { label: 'All severities', value: '' },
  { label: 'Critical', value: 'critical' },
  { label: 'High', value: 'high' },
  { label: 'Medium', value: 'medium' },
  { label: 'Low', value: 'low' },
]

onMounted(load)

async function load() {
  const params = new URLSearchParams({ limit: String(rows), offset: String(page.value * rows) })
  if (severity.value) params.set('severity', severity.value)
  if (search.value.trim()) params.set('search', search.value.trim())
  const data = await api.get<any>(`/api/security/cves?${params}`)
  items.value = data.items
  total.value = data.total
  summary.value = data.summary
}

function onPage(e: { page: number }) {
  page.value = e.page
  load()
}

function applyFilters() {
  page.value = 0
  router.replace({
    query: {
      ...(search.value.trim() ? { search: search.value.trim() } : {}),
      ...(severity.value ? { severity: severity.value } : {}),
    },
  })
  load()
}

async function exportCsv() {
  exporting.value = true
  try {
    const params = new URLSearchParams()
    if (severity.value) params.set('severity', severity.value)
    if (search.value.trim()) params.set('search', search.value.trim())
    const qs = params.toString()
    await api.download(`/api/security/cves/export${qs ? `?${qs}` : ''}`, 'switchboard-cves.csv')
    toast.add({ severity: 'success', summary: 'CSV downloaded', life: 2500 })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Export failed',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    exporting.value = false
  }
}
</script>

<template>
  <div class="page">
    <PageHeader title="CVE findings" subtitle="Drill-down list from Harbor scans and pulls.">
      <template #actions>
        <Button label="Overview" icon="pi pi-chart-bar" text @click="router.push('/security')" />
        <Button
          label="Export CSV"
          icon="pi pi-download"
          severity="secondary"
          outlined
          :loading="exporting"
          @click="exportCsv"
        />
      </template>
    </PageHeader>

    <div class="stats">
      <div class="stat critical"><span>Critical</span><strong>{{ summary.critical_count || 0 }}</strong></div>
      <div class="stat high"><span>High</span><strong>{{ summary.high_count || 0 }}</strong></div>
      <div class="stat medium"><span>Medium</span><strong>{{ summary.medium_count || 0 }}</strong></div>
      <div class="stat low"><span>Low</span><strong>{{ summary.low_count || 0 }}</strong></div>
    </div>

    <div class="toolbar surface-card">
      <InputText
        v-model="search"
        placeholder="Search image or CVE ID…"
        class="search"
        @keyup.enter="applyFilters"
      />
      <Select
        v-model="severity"
        :options="severityOptions"
        option-label="label"
        option-value="value"
        placeholder="Filter severity"
        class="filter"
        @change="applyFilters"
      />
      <Button label="Search" icon="pi pi-search" @click="applyFilters" />
      <span class="total">{{ total }} findings</span>
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
      <Column field="cve_id" header="CVE" />
      <Column field="image_name" header="Image" />
      <Column field="image_tag" header="Tag" />
      <Column field="severity" header="Severity">
        <template #body="{ data }"><SeverityTag :severity="data.severity" /></template>
      </Column>
      <Column field="package_name" header="Package" />
      <Column header="Fixable">
        <template #body="{ data }">
          <Tag
            :value="data.fixable ? 'Yes' : 'No'"
            :severity="data.fixable ? 'success' : 'secondary'"
          />
        </template>
      </Column>
      <Column field="age_days" header="Age (days)" />
      <Column field="scan_date" header="Scan Date" />
      <template #empty>
        <TableEmptyState
          :title="severity || search ? 'No matching CVEs' : 'No CVE findings'"
          :message="
            severity || search
              ? 'Try different filters or wait for new scan results.'
              : 'Findings appear here from Harbor webhooks when API credentials are configured.'
          "
          icon="pi-shield"
        />
      </template>
    </DataTable>
  </div>
</template>

<style scoped>
.stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.85rem;
}
.stat {
  padding: 1rem 1.1rem;
  border-radius: var(--sb-radius-sm);
  background: var(--sb-surface);
  border: 1px solid var(--sb-border);
  box-shadow: var(--sb-shadow);
}
.stat span {
  display: block;
  font-size: 0.8rem;
  color: var(--sb-muted);
  margin-bottom: 0.35rem;
  font-weight: 600;
}
.stat strong {
  font-size: 1.5rem;
}
.stat.critical strong {
  color: #dc2626;
}
.stat.high strong {
  color: #ea580c;
}
.stat.medium strong {
  color: #ca8a04;
}
.stat.low strong {
  color: #16a34a;
}
.toolbar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.9rem 1rem;
  flex-wrap: wrap;
}
.search {
  flex: 1;
  min-width: 200px;
}
.filter {
  min-width: 180px;
}
.total {
  margin-left: auto;
  color: var(--sb-muted);
  font-size: 0.875rem;
  font-weight: 600;
}
.table-card {
  overflow: hidden;
}
@media (max-width: 800px) {
  .stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
