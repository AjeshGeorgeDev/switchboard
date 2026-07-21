<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import TableEmptyState from '../../../components/TableEmptyState.vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Checkbox from 'primevue/checkbox'
import Select from 'primevue/select'
import Tag from 'primevue/tag'

type GroupRoleMapping = { group: string; role_id: string }

const toast = useToast()
const providers = ref<any[]>([])
const roles = ref<any[]>([])
const showDialog = ref(false)
const editing = ref<any>(null)
const saving = ref(false)

const emptyForm = () => ({
  name: '',
  display_name: '',
  issuer_url: '',
  client_id: '',
  client_secret: '',
  scopes: ['openid', 'profile', 'email'] as string[],
  auto_provision: true,
  default_role_id: null as string | null,
  is_active: true,
  claim_subject: 'sub',
  claim_email: 'email',
  claim_name: 'name',
  claim_groups: 'groups',
  group_role_mappings: [] as GroupRoleMapping[],
})

const form = ref(emptyForm())

onMounted(async () => {
  await Promise.all([load(), loadRoles()])
})

async function load() {
  try {
    providers.value = await api.get<any[]>('/api/admin/oidc-providers')
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Failed to load OIDC providers', detail: e instanceof Error ? e.message : 'Unknown error', life: 5000 })
  }
}

async function loadRoles() {
  try {
    roles.value = await api.get<any[]>('/api/admin/roles')
  } catch {
    roles.value = []
  }
}

function parseMappings(raw: unknown): GroupRoleMapping[] {
  if (!raw) return []
  if (Array.isArray(raw)) {
    return raw
      .map((m: any) => ({ group: String(m.group || ''), role_id: String(m.role_id || '') }))
      .filter((m) => m.group || m.role_id)
  }
  if (typeof raw === 'string') {
    try {
      return parseMappings(JSON.parse(raw))
    } catch {
      try {
        return parseMappings(JSON.parse(atob(raw)))
      } catch {
        return []
      }
    }
  }
  return []
}

function uuidFromPg(value: unknown): string | null {
  if (!value) return null
  if (typeof value === 'string') return value
  return null
}

function openCreate() {
  editing.value = null
  form.value = emptyForm()
  showDialog.value = true
}

function openEdit(row: any) {
  editing.value = row
  form.value = {
    name: row.name,
    display_name: row.display_name,
    issuer_url: row.issuer_url,
    client_id: row.client_id,
    client_secret: '',
    scopes: row.scopes || ['openid', 'profile', 'email'],
    auto_provision: row.auto_provision,
    default_role_id: uuidFromPg(row.default_role_id),
    is_active: row.is_active,
    claim_subject: row.claim_subject || 'sub',
    claim_email: row.claim_email || 'email',
    claim_name: row.claim_name || 'name',
    claim_groups: row.claim_groups || 'groups',
    group_role_mappings: parseMappings(row.group_role_mappings),
  }
  showDialog.value = true
}

function addMapping() {
  form.value.group_role_mappings.push({ group: '', role_id: '' })
}

function removeMapping(index: number) {
  form.value.group_role_mappings.splice(index, 1)
}

async function save() {
  saving.value = true
  try {
    const mappings = form.value.group_role_mappings
      .map((m) => ({ group: m.group.trim(), role_id: m.role_id }))
      .filter((m) => m.group && m.role_id)

    const body: Record<string, unknown> = {
      display_name: form.value.display_name,
      issuer_url: form.value.issuer_url,
      client_id: form.value.client_id,
      scopes: form.value.scopes,
      auto_provision: form.value.auto_provision,
      default_role_id: form.value.default_role_id,
      is_active: form.value.is_active,
      claim_subject: form.value.claim_subject.trim() || 'sub',
      claim_email: form.value.claim_email.trim() || 'email',
      claim_name: form.value.claim_name.trim() || 'name',
      claim_groups: form.value.claim_groups.trim() || 'groups',
      group_role_mappings: mappings,
    }
    if (form.value.client_secret) body.client_secret = form.value.client_secret

    if (editing.value) {
      await api.patch(`/api/admin/oidc-providers/${editing.value.id}`, body)
    } else {
      await api.post('/api/admin/oidc-providers', {
        name: form.value.name,
        client_secret: form.value.client_secret,
        ...body,
      })
    }
    showDialog.value = false
    await load()
    toast.add({ severity: 'success', summary: 'OIDC provider saved', life: 3000 })
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Save failed', detail: e instanceof Error ? e.message : 'Unknown error', life: 5000 })
  } finally {
    saving.value = false
  }
}

async function remove(row: any) {
  if (!confirm(`Delete OIDC provider "${row.display_name}"?`)) return
  await api.delete(`/api/admin/oidc-providers/${row.id}`)
  await load()
}

const roleOptions = () => roles.value.map((r) => ({ label: r.name, value: r.id }))
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Identity (OIDC)</h2>
        <p class="config-card-lead">
          Configure OpenID Connect providers for enterprise sign-in. Users see active providers on the sign-in page.
        </p>
      </div>
      <Button label="Add provider" icon="pi pi-plus" @click="openCreate" />
    </header>

    <ol class="setup-steps compact">
      <li>Register Switchboard as an OIDC client in your IdP (Azure AD, Okta, etc.).</li>
      <li>Set the redirect URI to <code>/api/auth/oidc/{name}/callback</code> (replace <code>{name}</code> with the provider slug).</li>
      <li>Paste issuer URL, client ID, and client secret below. Map claim keys and groups if your IdP uses non-standard names.</li>
    </ol>

    <DataTable :value="providers" class="table-card">
      <Column field="name" header="Slug" />
      <Column field="display_name" header="Display Name" />
      <Column field="issuer_url" header="Issuer URL" />
      <Column field="is_active" header="Active">
        <template #body="{ data }">
          <Tag :value="data.is_active ? 'Active' : 'Inactive'" :severity="data.is_active ? 'success' : 'secondary'" />
        </template>
      </Column>
      <Column field="auto_provision" header="Auto-provision" />
      <Column header="">
        <template #body="{ data }">
          <Button icon="pi pi-pencil" text @click="openEdit(data)" />
          <Button icon="pi pi-trash" text severity="danger" @click="remove(data)" />
        </template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No OIDC providers"
          message="Add a provider to enable enterprise sign-in alongside local accounts."
          icon="pi-key"
        />
      </template>
    </DataTable>

    <Dialog
      v-model:visible="showDialog"
      :header="editing ? 'Edit OIDC provider' : 'Add OIDC provider'"
      modal
      class="oidc-dialog"
      :style="{ width: 'min(640px, 95vw)' }"
      :content-style="{ overflow: 'auto', maxHeight: 'min(78vh, 900px)' }"
    >
      <div class="field" v-if="!editing">
        <label>Slug (URL name)</label>
        <InputText v-model="form.name" class="w-full" placeholder="azure" />
      </div>
      <div class="field">
        <label>Display name</label>
        <InputText v-model="form.display_name" class="w-full" placeholder="Azure AD" />
      </div>
      <div class="field">
        <label>Issuer URL</label>
        <InputText v-model="form.issuer_url" class="w-full" placeholder="https://login.microsoftonline.com/{tenant}/v2.0" />
      </div>
      <div class="field">
        <label>Client ID</label>
        <InputText v-model="form.client_id" class="w-full" />
      </div>
      <div class="field">
        <label>Client secret</label>
        <InputText v-model="form.client_secret" class="w-full" type="password" :placeholder="editing ? 'Leave blank to keep current' : ''" />
      </div>

      <h3 class="section-label">Claim keys</h3>
      <p class="hint">Override IdP claim names when they differ from the defaults.</p>
      <div class="claim-grid">
        <div class="field">
          <label>Subject</label>
          <InputText v-model="form.claim_subject" class="w-full" placeholder="sub" />
        </div>
        <div class="field">
          <label>Email</label>
          <InputText v-model="form.claim_email" class="w-full" placeholder="email" />
        </div>
        <div class="field">
          <label>Display name</label>
          <InputText v-model="form.claim_name" class="w-full" placeholder="name" />
        </div>
        <div class="field">
          <label>Groups</label>
          <InputText v-model="form.claim_groups" class="w-full" placeholder="groups" />
        </div>
      </div>

      <h3 class="section-label">Group → role mappings</h3>
      <p class="hint">
        Matching groups add roles on every login (additive). Default role is used only on first login when no group matches.
        Include a groups-capable scope if your IdP requires it.
      </p>
      <div v-for="(mapping, index) in form.group_role_mappings" :key="index" class="mapping-row">
        <InputText v-model="mapping.group" class="w-full" placeholder="IdP group name or ID" />
        <Select
          v-model="mapping.role_id"
          :options="roleOptions()"
          option-label="label"
          option-value="value"
          placeholder="Role"
          class="w-full"
        />
        <Button icon="pi pi-trash" text severity="danger" @click="removeMapping(index)" />
      </div>
      <Button label="Add mapping" icon="pi pi-plus" severity="secondary" outlined size="small" class="mb-3" @click="addMapping" />

      <div class="field">
        <label>Default role</label>
        <Select v-model="form.default_role_id" :options="roleOptions()" option-label="label" option-value="value" placeholder="None" class="w-full" show-clear />
      </div>
      <div class="field-row">
        <Checkbox v-model="form.auto_provision" input-id="oidc-auto" binary />
        <label for="oidc-auto">Auto-provision new users</label>
      </div>
      <div class="field-row">
        <Checkbox v-model="form.is_active" input-id="oidc-active" binary />
        <label for="oidc-active">Active</label>
      </div>
      <Button :label="editing ? 'Update' : 'Create'" :loading="saving" @click="save" />
    </Dialog>
  </section>
</template>

<style scoped>
.section-label {
  margin: 1.25rem 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 700;
}
.hint {
  margin: 0 0 0.75rem;
  color: var(--sb-muted, var(--p-text-muted-color));
  font-size: 0.8rem;
  line-height: 1.4;
}
.claim-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}
.mapping-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 0.5rem;
}
.mb-3 {
  margin-bottom: 1rem;
}
@media (max-width: 640px) {
  .claim-grid,
  .mapping-row {
    grid-template-columns: 1fr;
  }
}
</style>

<style>
.oidc-dialog { width: min(640px, 95vw); }
</style>
