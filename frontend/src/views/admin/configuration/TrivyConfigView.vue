<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useWebhookEndpoints } from '../../../composables/useWebhookEndpoints'
import { api } from '../../../api'

const { trivyUrl, secretStatus, loadSecretStatus, copy } = useWebhookEndpoints()
const pullStatus = ref<any>({})

onMounted(async () => {
  await loadSecretStatus()
  try {
    pullStatus.value = await api.get('/api/admin/webhook-endpoints')
  } catch {
    pullStatus.value = {}
  }
})
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Trivy</h2>
        <p class="config-card-lead">
          Receive Trivy JSON scan reports from CI/CD or scanning jobs. Findings appear under
          <strong>Security → CVEs</strong>. Critical CVEs trigger alerts.
        </p>
      </div>
      <Tag
        :value="secretStatus.trivy ? 'Secret configured' : 'No secret'"
        :severity="secretStatus.trivy ? 'success' : 'warn'"
      />
    </header>

    <div class="status-banner" :class="pullStatus.cve_pull_enabled ? 'enabled' : 'disabled'">
      <strong>Weekly CVE pull:</strong>
      <span v-if="!pullStatus.cve_pull_enabled">
        Disabled (webhook-only mode). Set <code>CVE_PULL_ENABLED=true</code> on the server to enable scheduled pulls.
      </span>
      <span v-else-if="!pullStatus.cve_pull_configured">
        Enabled but not configured — set <code>TRIVY_URL</code> and <code>TRIVY_TOKEN</code>. Trivy has no standard global CVE inventory API; webhook ingestion is the recommended path.
      </span>
      <span v-else>
        Enabled (cron: {{ pullStatus.cve_pull_cron }}). Note: Trivy server does not expose a standard “fetch all CVEs” API; scheduled pull will not import findings until a custom integration is added.
      </span>
    </div>

    <div class="endpoint-row">
      <code>{{ trivyUrl }}</code>
      <Button icon="pi pi-copy" text label="Copy URL" @click="copy(trivyUrl)" />
    </div>

    <p class="config-note">
      Harbor's built-in scanner fires events to the <strong>Harbor</strong> tab endpoint.
      To populate the CVE dashboard with full Trivy report detail, POST the standard Trivy JSON
      output to this endpoint from your pipeline.
    </p>

    <ol class="setup-steps">
      <li>
        <strong>Run Trivy in CI</strong> — After building an image, scan it and capture JSON:
        <pre class="code-block">trivy image --format json --output trivy-report.json myregistry/myapp:v1.0.0</pre>
      </li>
      <li>
        <strong>POST the report to Switchboard</strong> — Send the raw JSON file to the Trivy URL.
        Example GitHub Actions / shell step:
        <pre class="code-block"># Without authentication (local dev)
curl -i -X POST "{{ trivyUrl }}" \
  -H "Content-Type: application/json" \
  --data-binary @trivy-report.json

# With HMAC secret (production)
SIG=$(openssl dgst -sha256 -hmac "$TRIVY_WEBHOOK_SECRET" -hex trivy-report.json | awk '{print $2}')
curl -i -X POST "{{ trivyUrl }}" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Signature: $SIG" \
  --data-binary @trivy-report.json</pre>
      </li>
      <li>
        <strong>Set server secret</strong> — Add <code>TRIVY_WEBHOOK_SECRET</code> to the
        Switchboard environment and use the same value in your CI variable store.
      </li>
      <li>
        <strong>Verify</strong> — A successful POST returns <code>202 Accepted</code>. Open
        <em>Security → CVEs</em> to confirm findings. Critical severities also trigger in-app
        and Teams/email notifications (if configured).
      </li>
    </ol>

    <details class="setup-details">
      <summary>Expected JSON format</summary>
      <p class="config-note">
        Switchboard expects standard Trivy JSON with <code>artifact_name</code> (e.g.
        <code>myregistry/myapp:1.0.0</code>) and <code>Results[].Vulnerabilities[]</code>
        containing <code>VulnerabilityID</code>, <code>Severity</code>, <code>PkgName</code>, etc.
      </p>
    </details>
  </section>
</template>

<style scoped>
.status-banner {
  padding: 0.85rem 1rem;
  border-radius: var(--sb-radius-sm);
  margin-bottom: 1rem;
  font-size: 0.9rem;
  line-height: 1.5;
}
.status-banner.disabled {
  background: #fffbeb;
  border: 1px solid #fcd34d;
  color: #92400e;
}
.status-banner.enabled {
  background: #eff6ff;
  border: 1px solid #93c5fd;
  color: #1e40af;
}
.status-banner strong { margin-right: 0.35rem; }
</style>
