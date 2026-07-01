import { ref, computed } from 'vue'
import { useToast } from 'primevue/usetoast'
import { api } from '../api'

export function useWebhookEndpoints() {
  const toast = useToast()
  const secretStatus = ref({ harbor: false, trivy: false })

  const baseUrl = computed(() => window.location.origin.replace(/\/$/, ''))
  const harborUrl = computed(() => `${baseUrl.value}/webhooks/harbor`)
  const trivyUrl = computed(() => `${baseUrl.value}/webhooks/trivy`)

  async function loadSecretStatus() {
    try {
      const status = await api.get<any>('/api/admin/webhook-endpoints')
      secretStatus.value = {
        harbor: !!status.harbor_secret_configured,
        trivy: !!status.trivy_secret_configured,
      }
    } catch {
      // URLs still work from browser origin
    }
  }

  async function copy(text: string) {
    await navigator.clipboard.writeText(text)
    toast.add({ severity: 'success', summary: 'Copied to clipboard', life: 2000 })
  }

  return { harborUrl, trivyUrl, secretStatus, loadSecretStatus, copy }
}
