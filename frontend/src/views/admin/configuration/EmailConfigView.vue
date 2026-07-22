<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Password from 'primevue/password'

const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const configured = ref(false)
const passConfigured = ref(false)
const form = ref({
  host: '',
  port: 587 as number | null,
  user: '',
  pass: '',
  from: '',
})

onMounted(load)

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
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'SMTP test failed',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 6000,
    })
  } finally {
    testing.value = false
  }
}
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Email (SMTP)</h2>
        <p class="config-card-lead">
          Server-side email delivery for deployment reports, critical CVEs, invites, and weekly
          digests.
        </p>
      </div>
      <Tag
        :value="configured ? 'Configured' : 'Not configured'"
        :severity="configured ? 'success' : 'secondary'"
      />
    </header>

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

    <p class="config-note">
      <strong>Send test email</strong> delivers a message to your signed-in account address. Users can
      opt in/out per event type under <em>Profile → Notification preferences</em>. Env vars
      (<code>SMTP_*</code>) still work as a fallback if nothing is saved here.
    </p>
  </section>
</template>

<style scoped>
.config-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  margin-bottom: 1rem;
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
</style>
