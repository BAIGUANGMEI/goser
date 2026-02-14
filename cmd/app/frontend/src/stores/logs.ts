import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type LogEntry } from '@/api/wails'

export const useLogStore = defineStore('logs', () => {
  const logs = ref<LogEntry[]>([])
  const loading = ref(false)
  const selectedService = ref<string>('')

  async function fetchLogs(serviceName: string, n: number = 200) {
    loading.value = true
    selectedService.value = serviceName
    try {
      logs.value = await api.getLogs(serviceName, n) || []
    } catch {
      logs.value = []
    } finally {
      loading.value = false
    }
  }

  function clearLogs() {
    logs.value = []
    selectedService.value = ''
  }

  return {
    logs,
    loading,
    selectedService,
    fetchLogs,
    clearLogs
  }
})
