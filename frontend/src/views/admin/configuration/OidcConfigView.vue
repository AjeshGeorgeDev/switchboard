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

const toast = useToast()
const providers = ref<any[]>([])
const roles = ref<any[]>([])
const showDialog = ref(false)
const editing = ref<any>(null)
const saving = ref(false)

const form = ref({
  name: '',
  display_name: '',
  issuer_url: '',
  client_id: '',
  client_secret: '',
  scopes: ['openid', 'profile', 'email'],
  auto_provision: true,
  default_role_id: null as string | null,
  is_active: true,
})

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

function openCreate() {
  editing.value = null
  form.value = {
    name: '',
    display_name: '',
    issuer_url: '',
    client_id: '',
    client_secret: '',
    scopes: ['openid', 'profile', 'email'],
    auto_provision: true,
    default_role_id: null,
    is_active: true,
  }
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
    default_role_id: row.default_role_id?.Valid ? row.default_role_id.Bytes : row.default_role_id || null,
    is_active: row.is_active,
  }
  showDialog.value = true
}

async function save() {
  saving.value = true
  try {
    if (editing.value) {
      const body: Record<string, unknown> = {
        display_name: form.value.display_name,
        issuer_url: form.value.issuer_url,
        client_id: form.value.client_id,
        scopes: form.value.scopes,
        auto_provision: form.value.auto_provision,
        default_role_id: form.value.default_role_id,
        is_active: form.value.is_active,
      }
      if (form.value.client_secret) body.client_secret = form.value.client_secret
      await api.patch(`/api/admin/oidc-providers/${editing.value.id}`, body)
    } else {
      await api.post('/api/admin/oidc-providers', {
        ...form.value,
        default_role_id: form.value.default_role_id || undefined,
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
      <li>Paste issuer URL, client ID, and client secret below.</li>
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

    <Dialog v-model:visible="showDialog" :header="editing ? 'Edit OIDC provider' : 'Add OIDC provider'" modal class="oidc-dialog">
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

<style>
.oidc-dialog { width: min(520px, 95vw); }
</style>
