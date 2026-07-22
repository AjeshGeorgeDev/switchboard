<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useToast } from 'primevue/usetoast'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import { api } from '../../../api'
import { useWebhookEndpoints } from '../../../composables/useWebhookEndpoints'

const toast = useToast()
const {
  harborUrl,
  secretStatus,
  harborApiConfigured,
  loadSecretStatus,
  copy,
} = useWebhookEndpoints()

const loading = ref(true)
const saving = ref(false)
const testing = ref(false)
const form = ref({
  url: '',
  user: '',
  token: '',
  webhook_secret: '',
})
const tokenConfigured = ref(false)
const webhookSecretConfigured = ref(false)

onMounted(async () => {
  await Promise.all([loadSecretStatus(), loadHarbor()])
  loading.value = false
})

async function loadHarbor() {
  try {
    const data = await api.get<{
      url?: string
      user?: string
      token_configured?: boolean
      webhook_secret_configured?: boolean
      api_configured?: boolean
    }>('/api/admin/settings/harbor')
    form.value = {
      url: data.url || '',
      user: data.user || '',
      token: '',
      webhook_secret: '',
    }
    tokenConfigured.value = !!data.token_configured
    webhookSecretConfigured.value = !!data.webhook_secret_configured
    harborApiConfigured.value = !!data.api_configured
    secretStatus.value.harbor = !!data.webhook_secret_configured
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load Harbor settings',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  }
}

async function save() {
  saving.value = true
  try {
    const data = await api.put<{
      url?: string
      user?: string
      token_configured?: boolean
      webhook_secret_configured?: boolean
      api_configured?: boolean
    }>('/api/admin/settings/harbor', {
      url: form.value.url.trim(),
      user: form.value.user.trim(),
      token: form.value.token,
      webhook_secret: form.value.webhook_secret,
    })
    form.value.token = ''
    form.value.webhook_secret = ''
    tokenConfigured.value = !!data.token_configured
    webhookSecretConfigured.value = !!data.webhook_secret_configured
    harborApiConfigured.value = !!data.api_configured
    secretStatus.value.harbor = !!data.webhook_secret_configured
    toast.add({ severity: 'success', summary: 'Harbor settings saved', life: 2500 })
    await loadSecretStatus()
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to save Harbor settings',
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
    const data = await api.post<{ ok?: boolean; message?: string }>('/api/admin/settings/harbor/test', {
      url: form.value.url.trim(),
      user: form.value.user.trim(),
      token: form.value.token,
    })
    toast.add({
      severity: 'success',
      summary: 'Harbor connection OK',
      detail: data.message || 'Connected successfully',
      life: 4000,
    })
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Harbor connection failed',
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
        <h2>Harbor</h2>
        <p class="config-card-lead">
          Receive deployment and scan events from Harbor. Reports land under
          <strong>Security → Reports</strong>; per-CVE details are pulled from Harbor’s API into
          <strong>Security → CVEs</strong> when API credentials are saved below.
        </p>
      </div>
      <div class="header-tags">
        <Tag
          :value="secretStatus.harbor ? 'Webhook HMAC set' : 'Webhook HMAC optional'"
          :severity="secretStatus.harbor ? 'success' : 'secondary'"
        />
        <Tag
          :value="harborApiConfigured ? 'API credentials set' : 'API credentials missing'"
          :severity="harborApiConfigured ? 'success' : 'warn'"
        />
      </div>
    </header>

    <div class="status-banner" :class="harborApiConfigured ? 'enabled' : 'disabled'">
      <strong>CVE enrichment:</strong>
      <span v-if="harborApiConfigured">
        On <code>SCANNING_COMPLETED</code>, Switchboard calls Harbor’s vulnerabilities API with the
        artifact digest and upserts findings into <em>Security → CVEs</em>.
      </span>
      <span v-else>
        Save Harbor URL, robot username, and token below (no backend restart needed). Leave webhook
        HMAC empty for native Harbor webhooks.
      </span>
    </div>

    <div class="endpoint-row">
      <code>{{ harborUrl }}</code>
      <Button icon="pi pi-copy" text label="Copy URL" type="button" @click="copy(harborUrl)" />
    </div>

    <form class="config-form" @submit.prevent="save">
      <div class="field">
        <label for="harbor-url">Harbor URL</label>
        <InputText
          id="harbor-url"
          v-model="form.url"
          class="w-full"
          placeholder="https://harbor.example.com"
          :disabled="loading"
        />
      </div>
      <div class="field">
        <label for="harbor-user">Robot / username</label>
        <InputText
          id="harbor-user"
          v-model="form.user"
          class="w-full"
          placeholder="robot$project+switchboard"
          :disabled="loading"
        />
      </div>
      <div class="field">
        <label for="harbor-token">
          Token / secret
          <span v-if="tokenConfigured" class="field-hint">(saved — leave blank to keep)</span>
        </label>
        <Password
          id="harbor-token"
          v-model="form.token"
          toggle-mask
          class="w-full"
          input-class="w-full"
          :feedback="false"
          :placeholder="tokenConfigured ? 'Leave blank to keep current' : 'Robot secret'"
          :disabled="loading"
        />
      </div>
      <div class="field">
        <label for="harbor-hmac">
          Webhook HMAC secret
          <span class="field-hint">(optional — leave empty for native Harbor)</span>
        </label>
        <Password
          id="harbor-hmac"
          v-model="form.webhook_secret"
          toggle-mask
          class="w-full"
          input-class="w-full"
          :feedback="false"
          :placeholder="
            webhookSecretConfigured ? 'Leave blank to keep current' : 'Optional shared secret'
          "
          :disabled="loading"
        />
      </div>
      <div class="form-actions">
        <Button type="submit" label="Save Harbor settings" icon="pi pi-check" :loading="saving" />
        <Button
          type="button"
          label="Test connection"
          icon="pi pi-bolt"
          severity="secondary"
          outlined
          :loading="testing"
          :disabled="loading || saving"
          @click="testConnection"
        />
      </div>
    </form>

    <ol class="setup-steps">
      <li>
        <strong>Prerequisites</strong> — Switchboard must be reachable from Harbor at
        <code>APP_BASE_URL</code>. Redis must be running (events are queued asynchronously).
      </li>
      <li>
        <strong>Enable Trivy in Harbor</strong> — In Harbor admin: <em>Configuration →
        Interrogation Services</em>, enable the Trivy scanner adapter.
      </li>
      <li>
        <strong>Create a project webhook</strong> — Open your Harbor project →
        <em>Webhooks</em> → <em>+ New Webhook</em>. Paste the endpoint URL above. Leave Harbor’s
        auth header empty unless you use a CI relay that sends <code>X-Webhook-Signature</code>.
      </li>
      <li>
        <strong>Select event types</strong> —
        <code>PUSH_ARTIFACT</code>, <code>SCANNING_COMPLETED</code>, and optionally
        <code>SCANNING_FAILED</code>.
      </li>
      <li>
        <strong>API credentials</strong> — Save Harbor URL, robot user, and secret above. The robot
        needs <em>Read Artifact Addition</em>. No restart required.
      </li>
      <li>
        <strong>Test</strong> — In Harbor, open the webhook → <em>Test</em>. Expect
        <code>202 Accepted</code>. Check <em>Security → Reports</em> and <em>Security → CVEs</em>.
      </li>
    </ol>

    <details class="setup-details">
      <summary>Test with curl (local dev, no HMAC secret)</summary>
      <pre class="code-block">curl -i -X POST "{{ harborUrl }}" \
  -H "Content-Type: application/json" \
  -d '{"repository_name":"myapp","image_name":"registry.example.com/myapp","image_tag":"v1.0.0","status":"success","critical_count":0,"high_count":1}'</pre>
    </details>
  </section>
</template>

<style scoped>
.header-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: flex-end;
}
.status-banner {
  margin-bottom: 1rem;
  padding: 0.85rem 1rem;
  border-radius: 0.75rem;
  font-size: 0.9rem;
  line-height: 1.45;
}
.status-banner.disabled {
  background: color-mix(in srgb, var(--p-orange-100, #ffedd5) 80%, transparent);
  border: 1px solid color-mix(in srgb, var(--p-orange-300, #fdba74) 50%, transparent);
}
.status-banner.enabled {
  background: color-mix(in srgb, var(--p-green-100, #dcfce7) 80%, transparent);
  border: 1px solid color-mix(in srgb, var(--p-green-300, #86efac) 50%, transparent);
}
.status-banner strong {
  margin-right: 0.35rem;
}
.config-form {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  margin-bottom: 1.5rem;
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
