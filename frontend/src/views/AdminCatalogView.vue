<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'
import PageHeader from '../components/PageHeader.vue'
import TableEmptyState from '../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import InputNumber from 'primevue/inputnumber'
import MultiSelect from 'primevue/multiselect'
import Checkbox from 'primevue/checkbox'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import { appSectionId, type CatalogSection } from '../utils/catalog'

const toast = useToast()
const apps = ref<any[]>([])
const roles = ref<any[]>([])
const sections = ref<CatalogSection[]>([])
const showDialog = ref(false)
const showSectionsDialog = ref(false)
const saving = ref(false)
const savingSection = ref(false)
const error = ref('')
const sectionError = ref('')
const editing = ref<any>(null)
const editingSection = ref<CatalogSection | null>(null)
const sectionForm = ref({ name: '', sort_order: 0 })
const iconFileInput = ref<HTMLInputElement | null>(null)
const form = ref({
  name: '', description: '', icon_url: '', access_type: 'url',
  target_host: '', target_port: null as number | null,
  is_active: true, is_public: false, sort_order: 0, section_id: null as string | null,
  role_ids: [] as string[],
})

const MAX_ICON_BYTES = 500 * 1024

const accessTypeOptions = [
  { label: 'URL', value: 'url' },
  { label: 'IP / Port', value: 'ip_port' },
]

onMounted(load)

async function load() {
  try {
    const [appList, roleList, sectionList] = await Promise.all([
      api.get<any[]>('/api/admin/applications'),
      api.get<any[]>('/api/admin/roles'),
      api.get<CatalogSection[]>('/api/admin/catalog-sections'),
    ])
    apps.value = appList
    roles.value = roleList
    sections.value = sectionList
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load catalog',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

function defaultRoleIds() {
  const viewer = roles.value.find(r => r.name === 'viewer')
  return viewer ? [viewer.id] : []
}

function sectionName(sectionId: string | null) {
  if (!sectionId) return '—'
  return sections.value.find(s => s.id === sectionId)?.name ?? '—'
}

function clearIcon() {
  form.value.icon_url = ''
  if (iconFileInput.value) iconFileInput.value.value = ''
}

function onIconFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  if (!file.type.startsWith('image/')) {
    error.value = 'Icon must be an image file (PNG, JPEG, WebP, SVG, or GIF).'
    input.value = ''
    return
  }
  if (file.size > MAX_ICON_BYTES) {
    error.value = 'Icon image must be 500 KB or smaller.'
    input.value = ''
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    form.value.icon_url = typeof reader.result === 'string' ? reader.result : ''
    error.value = ''
  }
  reader.onerror = () => {
    error.value = 'Could not read the selected image.'
    input.value = ''
  }
  reader.readAsDataURL(file)
}

function openCreate() {
  editing.value = null
  error.value = ''
  form.value = {
    name: '', description: '', icon_url: '', access_type: 'url',
    target_host: '', target_port: null, is_active: true, is_public: false, sort_order: 0,
    section_id: null,
    role_ids: defaultRoleIds(),
  }
  if (iconFileInput.value) iconFileInput.value.value = ''
  showDialog.value = true
}

async function openEdit(app: any) {
  editing.value = app
  error.value = ''
  let roleIds: string[] = []
  try {
    const appRoles = await api.get<any[]>(`/api/admin/applications/${app.id}/roles`)
    roleIds = appRoles.map(r => r.id)
  } catch {
    // keep empty if fetch fails
  }
  form.value = {
    name: app.name,
    description: typeof app.description === 'string' ? app.description : app.description?.String || '',
    icon_url: typeof app.icon_url === 'string' ? app.icon_url : app.icon_url?.String || '',
    access_type: app.access_type,
    target_host: app.target_host,
    target_port: app.target_port?.Int32 ?? app.target_port ?? null,
    is_active: Boolean(app.is_active),
    is_public: Boolean(app.is_public),
    sort_order: app.sort_order,
    section_id: appSectionId(app),
    role_ids: roleIds,
  }
  if (iconFileInput.value) iconFileInput.value.value = ''
  showDialog.value = true
}

function validate() {
  if (!form.value.name.trim()) {
    error.value = 'Name is required.'
    return false
  }
  if (!form.value.target_host.trim()) {
    error.value = 'Target host is required (e.g. https://grafana.example.com or 10.0.0.5).'
    return false
  }
  if (form.value.access_type === 'ip_port' && !form.value.target_port) {
    error.value = 'Port is required for IP/port applications.'
    return false
  }
  return true
}

function buildPayload() {
  return {
    name: form.value.name.trim(),
    description: form.value.description,
    icon_url: form.value.icon_url.trim(),
    access_type: form.value.access_type,
    target_host: form.value.target_host.trim(),
    target_port: form.value.target_port,
    is_active: Boolean(form.value.is_active),
    is_public: Boolean(form.value.is_public),
    sort_order: form.value.sort_order,
    section_id: form.value.section_id,
    role_ids: form.value.role_ids,
  }
}

async function save() {
  error.value = ''
  if (!validate()) return

  const appId = editing.value?.id
  const wasEdit = Boolean(appId)
  const payload = buildPayload()

  saving.value = true
  try {
    if (wasEdit) {
      await api.patch(`/api/admin/applications/${appId}`, payload)
    } else {
      await api.post('/api/admin/applications', payload)
    }
    showDialog.value = false
    editing.value = null
    await load()
    toast.add({
      severity: 'success',
      summary: wasEdit ? 'Application updated' : 'Application created',
      life: 3000,
    })
    if (!payload.role_ids.length) {
      toast.add({
        severity: 'warn',
        summary: 'No roles assigned',
        detail: 'Assign at least one role or other users will not see this app on the launcher.',
        life: 6000,
      })
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Save failed'
    toast.add({
      severity: 'error',
      summary: 'Could not save application',
      detail: error.value,
      life: 6000,
    })
  } finally {
    saving.value = false
  }
}

async function remove(app: any) {
  try {
    await api.delete(`/api/admin/applications/${app.id}`)
    await load()
    toast.add({ severity: 'success', summary: 'Application deleted', life: 3000 })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

const previewHref = computed(() => {
  const role = roles.value.find(r => r.name === 'viewer')?.name || roles.value[0]?.name || ''
  return role
    ? `/admin/catalog/preview?role=${encodeURIComponent(role)}`
    : '/admin/catalog/preview'
})

const sectionSelectOptions = computed(() => [
  { label: 'No section', value: null },
  ...sections.value.map(s => ({ label: s.name, value: s.id })),
])

function openSectionsDialog() {
  sectionError.value = ''
  editingSection.value = null
  sectionForm.value = { name: '', sort_order: sections.value.length }
  showSectionsDialog.value = true
}

function openEditSection(section: CatalogSection) {
  sectionError.value = ''
  editingSection.value = section
  sectionForm.value = { name: section.name, sort_order: section.sort_order }
}

function resetSectionForm() {
  editingSection.value = null
  sectionForm.value = { name: '', sort_order: sections.value.length }
  sectionError.value = ''
}

async function saveSection() {
  sectionError.value = ''
  if (!sectionForm.value.name.trim()) {
    sectionError.value = 'Section name is required.'
    return
  }

  savingSection.value = true
  try {
    const payload = {
      name: sectionForm.value.name.trim(),
      sort_order: sectionForm.value.sort_order,
    }
    const wasEdit = Boolean(editingSection.value)
    if (editingSection.value) {
      await api.patch(`/api/admin/catalog-sections/${editingSection.value.id}`, payload)
    } else {
      await api.post('/api/admin/catalog-sections', payload)
    }
    await load()
    resetSectionForm()
    toast.add({
      severity: 'success',
      summary: wasEdit ? 'Section updated' : 'Section created',
      life: 3000,
    })
  } catch (e) {
    sectionError.value = e instanceof Error ? e.message : 'Save failed'
  } finally {
    savingSection.value = false
  }
}

async function removeSection(section: CatalogSection) {
  try {
    await api.delete(`/api/admin/catalog-sections/${section.id}`)
    await load()
    if (editingSection.value?.id === section.id) resetSectionForm()
    toast.add({ severity: 'success', summary: 'Section deleted', life: 3000 })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Delete failed',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}
</script>

<template>
  <div class="page">
    <PageHeader title="Catalog Admin" subtitle="Manage launcher applications and role access.">
      <template #actions>
        <Button label="Sections" icon="pi pi-folder" severity="secondary" outlined @click="openSectionsDialog" />
        <Button
          label="Preview"
          icon="pi pi-eye"
          severity="secondary"
          class="preview-btn"
          as="a"
          :href="previewHref"
          target="_blank"
          rel="noopener"
        />
        <Button label="Add Application" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <DataTable :value="apps" paginator :rows="10" class="surface-card table-card">
      <Column field="name" header="Name" />
      <Column header="Section">
        <template #body="{ data }">
          {{ sectionName(appSectionId(data)) }}
        </template>
      </Column>
      <Column field="access_type" header="Type" />
      <Column field="target_host" header="Target" />
      <Column header="Active">
        <template #body="{ data }">
          <Tag :value="data.is_active ? 'Active' : 'Inactive'" :severity="data.is_active ? 'success' : 'secondary'" />
        </template>
      </Column>
      <Column header="Public">
        <template #body="{ data }">
          <Tag :value="data.is_public ? 'On homepage' : 'Sign-in only'" :severity="data.is_public ? 'info' : 'secondary'" />
        </template>
      </Column>
      <Column header="Actions">
        <template #body="{ data }">
          <Button icon="pi pi-pencil" text @click="openEdit(data)" />
          <Button icon="pi pi-trash" text severity="danger" @click="remove(data)" />
        </template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No applications"
          message="Add your first launcher application to get started."
        />
      </template>
    </DataTable>

    <Dialog
      v-model:visible="showDialog"
      :header="editing ? 'Edit Application' : 'New Application'"
      modal
      class="app-dialog"
      :style="{ width: 'min(720px, 96vw)' }"
      :content-style="{ overflow: 'auto', maxHeight: 'min(78vh, 900px)' }"
    >
      <Message v-if="error" severity="error" class="mb-3">{{ error }}</Message>

      <div class="field">
        <label for="app-name">Name</label>
        <InputText id="app-name" v-model="form.name" class="w-full" />
      </div>
      <div class="field">
        <label for="app-desc">Description</label>
        <InputText id="app-desc" v-model="form.description" class="w-full" />
      </div>
      <div class="field">
        <label for="app-icon">Icon</label>
        <div class="icon-row">
          <div class="icon-preview" aria-hidden="true">
            <img v-if="form.icon_url" :src="form.icon_url" alt="" />
            <span v-else>{{ (form.name || '?').charAt(0).toUpperCase() }}</span>
          </div>
          <div class="icon-controls">
            <InputText
              id="app-icon"
              v-model="form.icon_url"
              class="w-full"
              placeholder="https://… or upload an image"
            />
            <div class="icon-actions">
              <input
                ref="iconFileInput"
                type="file"
                accept="image/png,image/jpeg,image/webp,image/svg+xml,image/gif"
                class="sr-only"
                @change="onIconFileSelected"
              />
              <Button
                type="button"
                label="Upload image"
                icon="pi pi-upload"
                severity="secondary"
                outlined
                size="small"
                @click="iconFileInput?.click()"
              />
              <Button
                v-if="form.icon_url"
                type="button"
                label="Clear"
                icon="pi pi-times"
                severity="secondary"
                text
                size="small"
                @click="clearIcon"
              />
            </div>
            <small class="hint">Paste an image URL/link, or upload a small icon (max 500 KB).</small>
          </div>
        </div>
      </div>
      <div class="field">
        <label for="app-type">Access Type</label>
        <Select
          id="app-type"
          v-model="form.access_type"
          :options="accessTypeOptions"
          option-label="label"
          option-value="value"
          class="w-full"
        />
      </div>
      <div class="field">
        <label for="app-target">Target Host</label>
        <InputText
          id="app-target"
          v-model="form.target_host"
          class="w-full"
          :placeholder="form.access_type === 'url' ? 'https://app.example.com' : '10.0.0.5'"
        />
      </div>
      <div v-if="form.access_type === 'ip_port'" class="field">
        <label for="app-port">Port</label>
        <InputNumber id="app-port" v-model="form.target_port" class="w-full" :use-grouping="false" />
      </div>
      <div class="field">
        <label for="app-section">Section</label>
        <Select
          id="app-section"
          v-model="form.section_id"
          :options="sectionSelectOptions"
          option-label="label"
          option-value="value"
          placeholder="Choose a section"
          class="w-full"
          show-clear
        />
        <small class="hint">Group apps under headings like Internal Tools or Email on the launcher.</small>
      </div>
      <div class="field">
        <label for="app-sort">Sort Order</label>
        <InputNumber id="app-sort" v-model="form.sort_order" class="w-full" :use-grouping="false" />
      </div>
      <div class="field">
        <label for="app-roles">Roles</label>
        <MultiSelect
          id="app-roles"
          v-model="form.role_ids"
          :options="roles"
          option-label="name"
          option-value="id"
          display="chip"
          placeholder="Select roles that can launch this app"
          class="w-full"
        />
        <small class="hint">Users only see apps linked to their role on the launcher.</small>
      </div>
      <div class="field-row">
        <Checkbox v-model="form.is_active" input-id="app-active" :binary="true" />
        <label for="app-active">Active</label>
      </div>
      <div class="field-row">
        <Checkbox v-model="form.is_public" input-id="app-public" :binary="true" />
        <label for="app-public">Show on public homepage</label>
      </div>
      <small class="hint">Public apps appear on the homepage without signing in. They must also be active.</small>

      <template #footer>
        <Button label="Cancel" severity="secondary" text :disabled="saving" @click="showDialog = false" />
        <Button label="Save" icon="pi pi-check" :loading="saving" @click="save" />
      </template>
    </Dialog>

    <Dialog
      v-model:visible="showSectionsDialog"
      header="Catalog Sections"
      modal
      class="sections-dialog"
    >
      <p class="sections-intro">Sections appear as grouped headings on the homepage and launcher.</p>
      <Message v-if="sectionError" severity="error" class="mb-3">{{ sectionError }}</Message>

      <div class="field">
        <label for="section-name">Name</label>
        <InputText id="section-name" v-model="sectionForm.name" class="w-full" placeholder="e.g. Internal Tools" />
      </div>
      <div class="field">
        <label for="section-sort">Sort Order</label>
        <InputNumber id="section-sort" v-model="sectionForm.sort_order" class="w-full" :use-grouping="false" />
      </div>
      <div class="section-form-actions">
        <Button
          v-if="editingSection"
          label="Cancel edit"
          severity="secondary"
          text
          @click="resetSectionForm"
        />
        <Button
          :label="editingSection ? 'Update section' : 'Add section'"
          icon="pi pi-check"
          :loading="savingSection"
          @click="saveSection"
        />
      </div>

      <DataTable :value="sections" class="sections-table">
        <Column field="name" header="Name" />
        <Column field="sort_order" header="Order" />
        <Column header="Actions">
          <template #body="{ data }">
            <Button icon="pi pi-pencil" text @click="openEditSection(data)" />
            <Button icon="pi pi-trash" text severity="danger" @click="removeSection(data)" />
          </template>
        </Column>
        <template #empty>
          <TableEmptyState
            title="No sections"
            message="Add sections to group applications on the catalog."
          />
        </template>
      </DataTable>
    </Dialog>

  </div>
</template>

<style scoped>
.table-card { overflow: hidden; }
.mb-3 { margin-bottom: 1rem; }
.preview-btn {
  margin-right: 0.5rem;
}
.hint {
  color: var(--sb-muted);
  font-size: 0.8rem;
}
.icon-row {
  display: flex;
  gap: 0.85rem;
  align-items: flex-start;
}
.icon-preview {
  width: 3.25rem;
  height: 3.25rem;
  border-radius: 0.85rem;
  border: 1px solid var(--sb-border, var(--p-content-border-color));
  background: var(--sb-surface, var(--p-surface-100));
  display: grid;
  place-items: center;
  overflow: hidden;
  flex-shrink: 0;
  font-weight: 700;
  color: var(--sb-muted, var(--p-text-muted-color));
}
.icon-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.icon-controls {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}
.icon-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.sections-intro {
  margin-bottom: 1rem;
  color: var(--sb-muted);
  font-size: 0.9rem;
}
.section-form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.sections-table {
  margin-top: 0.5rem;
}
</style>

<style>
.app-dialog {
  width: min(720px, 96vw);
  max-width: 96vw;
}

.sections-dialog {
  width: min(640px, 96vw);
}
</style>
