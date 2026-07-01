<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import CatalogGrid from '../components/CatalogGrid.vue'
import { type CatalogSection } from '../utils/catalog'
import Select from 'primevue/select'
import Message from 'primevue/message'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const roles = ref<any[]>([])
const apps = ref<any[]>([])
const sections = ref<CatalogSection[]>([])
const loading = ref(false)
const role = ref('')

async function loadRoles() {
  roles.value = await api.get('/api/admin/roles')
  if (!role.value) {
    role.value =
      roles.value.find(r => r.name === 'viewer')?.name ||
      roles.value[0]?.name ||
      ''
  }
}

async function loadPreview() {
  if (!role.value) {
    apps.value = []
    return
  }
  loading.value = true
  try {
    const [appList, sectionList] = await Promise.all([
      api.get(`/api/admin/catalog/preview?role=${encodeURIComponent(role.value)}`),
      api.get<CatalogSection[]>('/api/catalog/sections'),
    ])
    apps.value = appList
    sections.value = sectionList
  } catch (e) {
    apps.value = []
    sections.value = []
    toast.add({
      severity: 'error',
      summary: 'Preview failed',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    loading.value = false
  }
}

function onRoleChange() {
  router.replace({ query: { role: role.value || undefined } })
  loadPreview()
}

onMounted(async () => {
  role.value = typeof route.query.role === 'string' ? route.query.role : ''
  try {
    await loadRoles()
    await loadPreview()
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load preview',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
})
</script>

<template>
  <div class="preview-page">
    <Message severity="info" :closable="false" class="preview-banner">
      Preview mode — showing active apps visible to the <strong>{{ role || 'selected' }}</strong> role.
    </Message>

    <header class="preview-hero">
      <h1>Catalog preview</h1>
      <p class="tagline">See how the home page looks for users with a given role.</p>
    </header>

    <div class="field role-picker">
      <label for="preview-role">Preview as role</label>
      <Select
        id="preview-role"
        v-model="role"
        :options="roles"
        option-label="name"
        option-value="name"
        class="w-full"
        @change="onRoleChange"
      />
    </div>

    <div v-if="loading" class="preview-loading">
      <i class="pi pi-spin pi-spinner" />
      <span>Loading preview…</span>
    </div>
    <CatalogGrid
      v-else
      :apps="apps"
      :sections="sections"
      preview
      :empty-message="`No active apps are assigned to the ${role} role.`"
    />
  </div>
</template>

<style scoped>
.preview-page {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  max-width: 1100px;
}

.preview-banner {
  margin-bottom: 0.25rem;
}

.preview-hero h1 {
  font-size: clamp(1.5rem, 3vw, 2rem);
  font-weight: 800;
  letter-spacing: -0.03em;
}

.tagline {
  margin-top: 0.35rem;
  color: var(--sb-muted);
  font-size: 0.95rem;
}

.role-picker {
  max-width: 320px;
}

.preview-loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--sb-muted);
  padding: 2rem 0;
  justify-content: center;
}
</style>
