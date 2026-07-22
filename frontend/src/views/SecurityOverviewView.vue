<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Tag from 'primevue/tag'

const router = useRouter()
const toast = useToast()
const loading = ref(true)
const exporting = ref(false)
const overview = ref<any>({ stats: {}, top_images: [] })

const stats = computed(() => overview.value.stats || {})

onMounted(load)

async function load() {
  loading.value = true
  try {
    overview.value = await api.get<any>('/api/security/overview?top=15')
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load overview',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    loading.value = false
  }
}

function openImage(imageName: string) {
  router.push({ path: '/security/cves', query: { search: imageName } })
}

async function exportCsv() {
  exporting.value = true
  try {
    await api.download('/api/security/cves/export', 'switchboard-cves.csv')
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
    <PageHeader
      title="Security Overview"
      subtitle="Harbor vulnerability posture without opening Harbor — criticals, aging, and riskiest images."
    >
      <template #actions>
        <Button
          label="Export CSV"
          icon="pi pi-download"
          severity="secondary"
          outlined
          :loading="exporting"
          @click="exportCsv"
        />
        <Button label="All CVEs" icon="pi pi-list" text @click="router.push('/security/cves')" />
      </template>
    </PageHeader>

    <div class="stats">
      <div class="stat critical">
        <span>Critical</span>
        <strong>{{ stats.critical_count || 0 }}</strong>
      </div>
      <div class="stat high">
        <span>High</span>
        <strong>{{ stats.high_count || 0 }}</strong>
      </div>
      <div class="stat new">
        <span>New this week</span>
        <strong>{{ stats.new_this_week || 0 }}</strong>
      </div>
      <div class="stat fixable">
        <span>Fixable criticals</span>
        <strong>{{ stats.fixable_critical || 0 }}</strong>
      </div>
    </div>

    <div class="aging surface-card">
      <h3>Critical + High aging</h3>
      <div class="aging-row">
        <div class="aging-bucket">
          <Tag value="< 7 days" severity="info" />
          <strong>{{ stats.aging_lt_7d || 0 }}</strong>
        </div>
        <div class="aging-bucket">
          <Tag value="7–30 days" severity="warn" />
          <strong>{{ stats.aging_7_to_30d || 0 }}</strong>
        </div>
        <div class="aging-bucket">
          <Tag value="> 30 days" severity="danger" />
          <strong>{{ stats.aging_gt_30d || 0 }}</strong>
        </div>
        <div class="aging-bucket muted">
          <span>Unfixed crit/high</span>
          <strong>{{ stats.unfixed_critical_high || 0 }}</strong>
        </div>
      </div>
    </div>

    <DataTable
      :value="overview.top_images || []"
      :loading="loading"
      data-key="image_name"
      class="surface-card table-card"
      @row-click="(e: any) => openImage(e.data.image_name)"
    >
      <Column field="image_name" header="Riskiest images" />
      <Column field="latest_tag" header="Latest tag" />
      <Column field="critical_count" header="Critical" />
      <Column field="high_count" header="High" />
      <Column field="total_count" header="Total" />
      <Column header="Oldest critical">
        <template #body="{ data }">
          <span v-if="data.oldest_critical_days != null">{{ data.oldest_critical_days }}d</span>
          <span v-else class="muted">—</span>
        </template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No vulnerability data yet"
          message="Findings appear after Harbor SCANNING_COMPLETED webhooks with API credentials configured."
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
  margin-bottom: 1rem;
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
.stat.new strong {
  color: #2563eb;
}
.stat.fixable strong {
  color: #16a34a;
}
.aging {
  padding: 1rem 1.1rem;
  margin-bottom: 1rem;
}
.aging h3 {
  font-size: 0.95rem;
  margin: 0 0 0.75rem;
}
.aging-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
}
.aging-bucket {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.aging-bucket strong {
  font-size: 1.15rem;
}
.aging-bucket.muted span {
  color: var(--sb-muted);
  font-size: 0.85rem;
  font-weight: 600;
}
.table-card {
  overflow: hidden;
  cursor: pointer;
}
.muted {
  color: var(--sb-muted);
}
@media (max-width: 800px) {
  .stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
