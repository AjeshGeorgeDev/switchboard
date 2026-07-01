<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../../../api'
import Tag from 'primevue/tag'

const toast = useToast()
const smtp = ref<any>({})
const loading = ref(true)

onMounted(async () => {
  try {
    smtp.value = await api.get<any>('/api/admin/smtp-status')
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: 'Failed to load SMTP status',
      detail: e instanceof Error ? e.message : 'Unknown error',
      life: 5000,
    })
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="config-section">
    <header class="config-card-header">
      <div>
        <h2>Email (SMTP)</h2>
        <p class="config-card-lead">
          Server-side email delivery for deployment reports, critical CVEs, and weekly digests.
        </p>
      </div>
      <Tag
        :value="smtp.configured ? 'Configured' : 'Not configured'"
        :severity="smtp.configured ? 'success' : 'secondary'"
      />
    </header>

    <p v-if="smtp.configured" class="smtp-status">
      Connected to <code>{{ smtp.host }}</code>
    </p>
    <p v-else class="config-note">
      Set these environment variables on the Switchboard server:
    </p>
    <ul v-if="!smtp.configured" class="env-list">
      <li><code>SMTP_HOST</code> — mail server hostname</li>
      <li><code>SMTP_PORT</code> — usually <code>587</code> (TLS) or <code>465</code></li>
      <li><code>SMTP_USER</code> / <code>SMTP_PASS</code> — credentials</li>
      <li><code>SMTP_FROM</code> — sender address (e.g. <code>switchboard@example.com</code>)</li>
    </ul>
    <p class="config-note">
      Users can opt in/out per event type under <em>Profile → Notification preferences</em>.
    </p>
  </section>
</template>
