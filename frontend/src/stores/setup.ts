import { ref } from 'vue'
import { defineStore } from 'pinia'
import { api } from '../api'

export const useSetupStore = defineStore('setup', () => {
  const complete = ref<boolean | null>(null)

  async function fetchStatus() {
    const data = await api.get<{ complete: boolean }>('/api/setup/status')
    complete.value = data.complete
    return data.complete
  }

  function markComplete() {
    complete.value = true
  }

  return { complete, fetchStatus, markComplete }
})
