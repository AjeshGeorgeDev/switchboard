<script setup lang="ts">
import { onMounted } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useWebhookEndpoints } from '../../../composables/useWebhookEndpoints'

const { harborUrl, secretStatus, loadSecretStatus, copy } = useWebhookEndpoints()

onMounted(loadSecretStatus)
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Harbor</h2>
        <p class="config-card-lead">
          Receive deployment and scan events from Harbor. Data appears under <strong>Security → Reports</strong>.
        </p>
      </div>
      <Tag
        :value="secretStatus.harbor ? 'Secret configured' : 'No secret'"
        :severity="secretStatus.harbor ? 'success' : 'warn'"
      />
    </header>

    <div class="endpoint-row">
      <code>{{ harborUrl }}</code>
      <Button icon="pi pi-copy" text label="Copy URL" @click="copy(harborUrl)" />
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
          <li><strong>Endpoint URL:</strong> paste the Harbor URL above</li>
          <li><strong>Auth Header</strong> (optional): if using a shared secret, Harbor cannot
            send HMAC natively — use a reverse proxy or CI relay, or leave
            <code>HARBOR_WEBHOOK_SECRET</code> empty in dev</li>
        </ul>
      </li>
      <li>
        <strong>Select event types</strong>
        <ul>
          <li><code>PUSH_ARTIFACT</code> — new image pushed to the registry</li>
          <li><code>SCANNING_COMPLETED</code> — vulnerability scan finished (includes severity counts)</li>
          <li><code>SCANNING_FAILED</code> — scan error (optional, for visibility)</li>
        </ul>
      </li>
      <li>
        <strong>Test the webhook</strong> — In Harbor, open the webhook → <em>Test</em>.
        Switchboard should return <code>202 Accepted</code> with
        <code>{"status":"accepted"}</code>. Check <em>Security → Reports</em> after a real push/scan.
      </li>
      <li>
        <strong>Authentication (production)</strong> — Set
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
