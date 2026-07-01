<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import Tag from 'primevue/tag'
import Button from 'primevue/button'

const route = useRoute()
const router = useRouter()
const report = ref<any>(null)
const loading = ref(true)

onMounted(load)

async function load() {
  loading.value = true
  try {
    report.value = await api.get<any>(`/api/security/reports/${route.params.id}`)
  } finally {
    loading.value = false
  }
}

const statusSeverity: Record<string, string> = {
  success: 'success', failed: 'danger', partial: 'warn',
}
</script>

<template>
  <div class="page">
    <PageHeader title="Report Detail" subtitle="Deployment scan report from Harbor webhook.">
      <template #actions>
        <Button label="Back to reports" icon="pi pi-arrow-left" text @click="router.push('/security/reports')" />
      </template>
    </PageHeader>

    <div v-if="loading" class="surface-card loading">Loading…</div>

    <div v-else-if="report" class="detail surface-card">
      <div class="meta-grid">
        <div><span>App</span><strong>{{ report.app_name }}</strong></div>
        <div><span>Image</span><strong>{{ report.image_name }}:{{ report.image_tag }}</strong></div>
        <div><span>Status</span><Tag :value="report.status" :severity="statusSeverity[report.status]" /></div>
        <div><span>Triggered by</span><strong>{{ report.triggered_by || '—' }}</strong></div>
        <div><span>Received</span><strong>{{ report.received_at }}</strong></div>
        <div><span>Severity counts</span><strong>C:{{ report.critical_count }} H:{{ report.high_count }} M:{{ report.medium_count }} L:{{ report.low_count }}</strong></div>
      </div>
      <h3>Raw payload</h3>
      <pre class="payload">{{ JSON.stringify(report.raw_payload, null, 2) }}</pre>
    </div>
  </div>
</template>

<style scoped>
.loading { padding: 2rem; text-align: center; color: var(--sb-muted); }
.detail { padding: 1.25rem 1.5rem; }
.meta-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.meta-grid span { display: block; font-size: 0.8rem; color: var(--sb-muted); font-weight: 600; margin-bottom: 0.25rem; }
.payload {
  margin: 0;
  padding: 1rem;
  font-size: 0.8rem;
  background: #f8fafc;
  border-radius: var(--sb-radius-sm);
  overflow-x: auto;
}
</style>
