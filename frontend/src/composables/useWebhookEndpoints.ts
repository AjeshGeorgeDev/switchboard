import { ref, computed, unref, type MaybeRef } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'

export function useWebhookEndpoints() {
  const toast = useToast()
  const secretStatus = ref({ harbor: false, trivy: false })
  const harborApiConfigured = ref(false)
  const configuredBaseUrl = ref('')

  const baseUrl = computed(() => {
    const fromApi = configuredBaseUrl.value.replace(/\/$/, '')
    if (fromApi) return fromApi
    return window.location.origin.replace(/\/$/, '')
  })
  const harborUrl = computed(() => `${baseUrl.value}/webhooks/harbor`)
  const trivyUrl = computed(() => `${baseUrl.value}/webhooks/trivy`)

  async function loadSecretStatus() {
    try {
      const status = await api.get<any>('/api/admin/webhook-endpoints')
      secretStatus.value = {
        harbor: !!status.harbor_secret_configured,
        trivy: !!status.trivy_secret_configured,
      }
      harborApiConfigured.value = !!status.harbor_api_configured
      if (typeof status.harbor_url === 'string' && status.harbor_url) {
        configuredBaseUrl.value = status.harbor_url.replace(/\/webhooks\/harbor\/?$/, '')
      }
    } catch {
      // URLs still work from browser origin
    }
  }

  async function copyToClipboard(text: string) {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
      return
    }
    // Fallback for non-secure contexts (plain HTTP) where clipboard API is unavailable
    const ta = document.createElement('textarea')
    ta.value = text
    ta.setAttribute('readonly', '')
    ta.style.position = 'fixed'
    ta.style.left = '-9999px'
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    if (!ok) throw new Error('copy command failed')
  }

  async function copy(text: MaybeRef<string>) {
    const value = String(unref(text) ?? '').trim()
    if (!value) {
      toast.add({ severity: 'warn', summary: 'Nothing to copy', life: 2500 })
      return
    }
    try {
      await copyToClipboard(value)
      toast.add({ severity: 'success', summary: 'Copied to clipboard', life: 2000 })
    } catch {
      toast.add({
        severity: 'error',
        summary: 'Copy failed',
        detail: 'Select the URL and copy it manually.',
        life: 4000,
      })
    }
  }

  return {
    harborUrl,
    trivyUrl,
    secretStatus,
    harborApiConfigured,
    loadSecretStatus,
    copy,
  }
}
