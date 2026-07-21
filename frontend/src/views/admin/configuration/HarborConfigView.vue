<script setup lang="ts">
import { onMounted } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useWebhookEndpoints } from '../../../composables/useWebhookEndpoints'

const {
  harborUrl,
  secretStatus,
  harborApiConfigured,
  loadSecretStatus,
  copy,
} = useWebhookEndpoints()

onMounted(loadSecretStatus)
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Harbor</h2>
        <p class="config-card-lead">
          Receive deployment and scan events from Harbor. Reports land under
          <strong>Security → Reports</strong>; per-CVE details are pulled from Harbor’s API into
          <strong>Security → CVEs</strong> when <code>HARBOR_URL</code> and <code>HARBOR_TOKEN</code> are set.
        </p>
      </div>
      <div class="header-tags">
        <Tag
          :value="secretStatus.harbor ? 'Webhook secret set' : 'No webhook secret'"
          :severity="secretStatus.harbor ? 'success' : 'warn'"
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
        Set <code>HARBOR_URL</code> and <code>HARBOR_TOKEN</code> (<code>username:secret</code> for a
        robot account, or a bearer token) on the server, then restart. Webhooks still create
        deployment reports without these.
      </span>
    </div>

    <div class="endpoint-row">
      <code>{{ harborUrl }}</code>
      <Button icon="pi pi-copy" text label="Copy URL" type="button" @click="copy(harborUrl)" />
    </div>

    <ol class="setup-steps">
      <li>
        <strong>Prerequisites</strong> — Switchboard must be reachable from Harbor at
        <code>APP_BASE_URL</code>. Redis must be running (events are queued asynchronously).
        In production, set <code>APP_BASE_URL</code> to your public URL (e.g.
        <code>https://switchboard.example.com</code>).
      </li>
      <li>
        <strong>Enable Trivy in Harbor</strong> — In Harbor admin: <em>Configuration →
        Interrogation Services</em>, add or enable the Trivy scanner adapter so image scans run
        on push or on schedule.
      </li>
      <li>
        <strong>Create a project webhook</strong> — Open your Harbor project →
        <em>Webhooks</em> → <em>+ New Webhook</em>.
        <ul>
          <li><strong>Name:</strong> <code>switchboard</code></li>
          <li><strong>Notify Type:</strong> HTTP</li>
          <li><strong>Endpoint URL:</strong> paste the webhook URL above</li>
          <li><strong>Auth Header</strong> (optional): if using a shared secret, Harbor cannot
            send HMAC natively — use a reverse proxy or CI relay, or leave
            <code>HARBOR_WEBHOOK_SECRET</code> empty in dev</li>
        </ul>
      </li>
      <li>
        <strong>Select event types</strong>
        <ul>
          <li><code>PUSH_ARTIFACT</code> — new image pushed to the registry</li>
          <li><code>SCANNING_COMPLETED</code> — vulnerability scan finished (summary counts + digest for CVE fetch)</li>
          <li><code>SCANNING_FAILED</code> — scan error (optional, for visibility)</li>
        </ul>
      </li>
      <li>
        <strong>API credentials (CVE details)</strong> — Set <code>HARBOR_URL</code> (e.g.
        <code>https://harbor.example.com</code>) and <code>HARBOR_TOKEN</code> as
        <code>username:secret</code> (robot account recommended) or a bearer token.
        On <code>SCANNING_COMPLETED</code>, Switchboard uses the artifact digest to call Harbor’s
        vulnerabilities API and upsert findings into <em>Security → CVEs</em>.
      </li>
      <li>
        <strong>Test the webhook</strong> — In Harbor, open the webhook → <em>Test</em>.
        Switchboard should return <code>202 Accepted</code> with
        <code>{"status":"accepted"}</code>. Check <em>Security → Reports</em> after a real push/scan,
        and <em>Security → CVEs</em> when API credentials are configured.
      </li>
      <li>
        <strong>Webhook signature (production)</strong> — Set
        <code>HARBOR_WEBHOOK_SECRET</code> in the Switchboard server environment. Requests must
        include header <code>X-Webhook-Signature</code> with the HMAC-SHA256 hex digest of the
        <em>raw request body</em>. Leave the secret empty in local dev to skip verification.
      </li>
    </ol>

    <details class="setup-details">
      <summary>Test with curl (local dev, no secret)</summary>
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
</style>
