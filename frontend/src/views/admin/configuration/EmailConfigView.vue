<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Password from 'primevue/password'
import MultiSelect from 'primevue/multiselect'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import TableEmptyState from '../../../components/TableEmptyState.vue'

const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const savingRecipients = ref(false)
const configured = ref(false)
const passConfigured = ref(false)
const form = ref({
  host: '',
  port: 587 as number | null,
  user: '',
  pass: '',
  from: '',
})

const roles = ref<any[]>([])
const digestRoles = ref<string[]>(['security-team'])
const criticalRoles = ref<string[]>(['security-team'])
const preview = ref<any[]>([])

const logItems = ref<any[]>([])
const logTotal = ref(0)
const logPage = ref(0)
const logRows = 20
const logFilter = ref('')
const logLoading = ref(false)
const showRecipients = ref(false)
const selectedLog = ref<any | null>(null)

const logFilters = [
  { label: 'All types', value: '' },
  { label: 'Weekly digest', value: 'weekly_digest' },
  { label: 'Critical CVE', value: 'critical_cve' },
  { label: 'Invite', value: 'invite' },
  { label: 'SMTP test', value: 'smtp_test' },
]

const statusSeverity: Record<string, string> = {
  sent: 'success',
  failed: 'danger',
  skipped: 'secondary',
}

onMounted(async () => {
  await Promise.all([load(), loadRoles(), loadRecipients(), loadLog()])
})

async function load() {
  loading.value = true
  try {
    const data = await api.get<{
      configured?: boolean
      host?: string
      port?: number
      user?: string
      from?: string
      pass_configured?: boolean
    }>('/api/admin/settings/smtp')
    form.value = {
      host: data.host || '',
      port: data.port || 587,
      user: data.user || '',
      pass: '',
      from: data.from || '',
    }
    configured.value = !!data.configured
    passConfigured.value = !!data.pass_configured
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load SMTP settings',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    loading.value = false
  }
}

async function loadRoles() {
  try {
    roles.value = await api.get<any[]>('/api/admin/roles')
  } catch {
    roles.value = []
  }
}

async function loadRecipients() {
  try {
    const data = await api.get<{
      weekly_digest_roles?: string[]
      critical_cve_roles?: string[]
      preview?: any[]
    }>('/api/admin/settings/email-recipients')
    digestRoles.value = data.weekly_digest_roles || ['security-team']
    criticalRoles.value = data.critical_cve_roles || ['security-team']
    preview.value = data.preview || []
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load email recipients',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

async function save() {
  saving.value = true
  try {
    const data = await api.put<{
      configured?: boolean
      host?: string
      port?: number
      user?: string
      from?: string
      pass_configured?: boolean
    }>('/api/admin/settings/smtp', {
      host: form.value.host.trim(),
      port: form.value.port || 587,
      user: form.value.user.trim(),
      pass: form.value.pass,
      from: form.value.from.trim(),
    })
    form.value.pass = ''
    form.value.host = data.host || form.value.host
    form.value.port = data.port || form.value.port
    form.value.user = data.user || form.value.user
    form.value.from = data.from || form.value.from
    configured.value = !!data.configured
    passConfigured.value = !!data.pass_configured
    toast.add({ severity: 'success', summary: 'SMTP settings saved', life: 2500 })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to save SMTP settings',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    saving.value = false
  }
}

async function saveRecipients() {
  savingRecipients.value = true
  try {
    const data = await api.put<{
      weekly_digest_roles?: string[]
      critical_cve_roles?: string[]
      preview?: any[]
    }>('/api/admin/settings/email-recipients', {
      weekly_digest_roles: digestRoles.value,
      critical_cve_roles: criticalRoles.value,
    })
    digestRoles.value = data.weekly_digest_roles || digestRoles.value
    criticalRoles.value = data.critical_cve_roles || criticalRoles.value
    preview.value = data.preview || []
    toast.add({ severity: 'success', summary: 'Recipients saved', life: 2500 })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to save recipients',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    savingRecipients.value = false
  }
}

async function testConnection() {
  testing.value = true
  try {
    const data = await api.post<{ ok?: boolean; message?: string }>('/api/admin/settings/smtp/test', {
      host: form.value.host.trim(),
      port: form.value.port || 587,
      user: form.value.user.trim(),
      pass: form.value.pass,
      from: form.value.from.trim(),
    })
    toast.add({
      severity: 'success',
      summary: 'SMTP test OK',
      detail: data.message || 'Test email sent',
      life: 5000,
    })
    await loadLog()
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'SMTP test failed',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 6000,
    })
    await loadLog()
  } finally {
    testing.value = false
  }
}

async function loadLog() {
  logLoading.value = true
  try {
    const params = new URLSearchParams({
      limit: String(logRows),
      offset: String(logPage.value * logRows),
    })
    if (logFilter.value) params.set('event_type', logFilter.value)
    const data = await api.get<{ items?: any[]; total?: number }>(`/api/admin/email-log?${params}`)
    logItems.value = data.items || []
    logTotal.value = data.total || 0
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load email log',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    logLoading.value = false
  }
}

function onLogPage(e: { page: number }) {
  logPage.value = e.page
  loadLog()
}

function openRecipients(row: any) {
  selectedLog.value = row
  showRecipients.value = true
}

function formatDate(value: unknown) {
  if (!value) return '—'
  const d = new Date(String(value))
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString()
}
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Email (SMTP)</h2>
        <p class="config-card-lead">
          Configure delivery, who receives security emails, and review outbound history.
        </p>
      </div>
      <Tag
        :value="configured ? 'Configured' : 'Not configured'"
        :severity="configured ? 'success' : 'secondary'"
      />
    </header>

    <h3 class="subhead">SMTP server</h3>
    <form class="config-form" @submit.prevent="save">
      <div class="field">
        <label for="smtp-host">Host</label>
        <InputText
          id="smtp-host"
          v-model="form.host"
          class="w-full"
          placeholder="smtp.example.com"
          :disabled="loading"
        />
      </div>
      <div class="field">
        <label for="smtp-port">Port</label>
        <InputNumber
          id="smtp-port"
          v-model="form.port"
          class="w-full"
          :use-grouping="false"
          :min="1"
          :max="65535"
          :disabled="loading"
        />
      </div>
      <div class="field">
        <label for="smtp-user">Username</label>
        <InputText id="smtp-user" v-model="form.user" class="w-full" :disabled="loading" />
      </div>
      <div class="field">
        <label for="smtp-pass">
          Password
          <span v-if="passConfigured" class="field-hint">(saved — leave blank to keep)</span>
        </label>
        <Password
          id="smtp-pass"
          v-model="form.pass"
          toggle-mask
          class="w-full"
          input-class="w-full"
          :feedback="false"
          :placeholder="passConfigured ? 'Leave blank to keep current' : ''"
          :disabled="loading"
        />
      </div>
      <div class="field">
        <label for="smtp-from">From address</label>
        <InputText
          id="smtp-from"
          v-model="form.from"
          class="w-full"
          placeholder="switchboard@example.com"
          :disabled="loading"
        />
      </div>
      <div class="form-actions">
        <Button type="submit" label="Save SMTP settings" icon="pi pi-check" :loading="saving" />
        <Button
          type="button"
          label="Send test email"
          icon="pi pi-send"
          severity="secondary"
          outlined
          :loading="testing"
          :disabled="loading || saving"
          @click="testConnection"
        />
      </div>
    </form>

    <h3 class="subhead">Who receives security email</h3>
    <p class="config-note">
      Choose roles for weekly digests and critical CVE alerts. Users in those roles still control
      opt-out under <em>Profile → Notification preferences</em>.
    </p>
    <div class="recipients-form">
      <div class="field">
        <label>Weekly digest roles</label>
        <MultiSelect
          v-model="digestRoles"
          :options="roles"
          option-label="name"
          option-value="name"
          display="chip"
          class="w-full"
          placeholder="Select roles"
        />
      </div>
      <div class="field">
        <label>Critical CVE roles</label>
        <MultiSelect
          v-model="criticalRoles"
          :options="roles"
          option-label="name"
          option-value="name"
          display="chip"
          class="w-full"
          placeholder="Select roles"
        />
      </div>
      <Button
        type="button"
        label="Save recipients"
        icon="pi pi-check"
        :loading="savingRecipients"
        @click="saveRecipients"
      />
    </div>

    <DataTable :value="preview" class="preview-table" data-key="id">
      <Column field="name" header="Resolved users" />
      <Column field="email" header="Email" />
      <Column field="roles" header="Roles">
        <template #body="{ data }">{{ (data.roles || []).join(', ') }}</template>
      </Column>
      <template #empty>
        <TableEmptyState
          title="No matching users"
          message="Assign users to the selected roles to see who can receive security email."
          icon="pi-users"
        />
      </template>
    </DataTable>

    <div class="log-header">
      <h3 class="subhead">Email log</h3>
      <div class="log-tools">
        <Select
          v-model="logFilter"
          :options="logFilters"
          option-label="label"
          option-value="value"
          class="log-filter"
          @change="
            () => {
              logPage = 0
              loadLog()
            }
          "
        />
        <Button type="button" label="Refresh" icon="pi pi-refresh" text @click="loadLog" />
      </div>
    </div>

    <DataTable
      :value="logItems"
      :lazy="true"
      :paginator="true"
      :rows="logRows"
      :total-records="logTotal"
      :loading="logLoading"
      data-key="id"
      class="log-table"
      @page="onLogPage"
    >
      <Column header="Sent">
        <template #body="{ data }">{{ formatDate(data.created_at) }}</template>
      </Column>
      <Column field="event_type" header="Type" />
      <Column field="subject" header="Subject" />
      <Column header="Recipients">
        <template #body="{ data }">
          <Button
            type="button"
            :label="String(data.recipient_count || 0)"
            text
            size="small"
            @click="openRecipients(data)"
          />
        </template>
      </Column>
      <Column header="Status">
        <template #body="{ data }">
          <Tag :value="data.status" :severity="statusSeverity[data.status] || 'secondary'" />
        </template>
      </Column>
      <Column field="error_message" header="Error" />
      <template #empty>
        <TableEmptyState
          title="No outbound email yet"
          message="Digests, critical alerts, invites, and SMTP tests appear here after send attempts."
          icon="pi-envelope"
        />
      </template>
    </DataTable>

    <Dialog
      v-model:visible="showRecipients"
      modal
      header="Recipients"
      :style="{ width: 'min(520px, 92vw)' }"
    >
      <DataTable :value="selectedLog?.recipients || []" data-key="email">
        <Column field="email" header="Email" />
        <Column field="status" header="Status" />
        <Column field="error_message" header="Error" />
      </DataTable>
    </Dialog>
  </section>
</template>

<style scoped>
.config-form,
.recipients-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  margin-bottom: 1.25rem;
  max-width: 36rem;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}
.field label {
  font-size: 0.85rem;
  font-weight: 600;
}
.field-hint {
  font-weight: 400;
  color: var(--sb-muted);
}
.form-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.subhead {
  font-size: 1rem;
  font-weight: 700;
  margin: 1.25rem 0 0.5rem;
}
.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  margin-top: 0.5rem;
}
.log-header .subhead {
  margin: 0;
}
.log-tools {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.log-filter {
  min-width: 160px;
}
.preview-table,
.log-table {
  margin-bottom: 1rem;
}
</style>
