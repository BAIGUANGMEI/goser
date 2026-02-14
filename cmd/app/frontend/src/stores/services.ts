import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type ServiceInfo, type DaemonStatus } from '@/api/wails'

export const useServiceStore = defineStore('services', () => {
  const services = ref<ServiceInfo[]>([])
  const daemonStatus = ref<DaemonStatus | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const daemonToggling = ref(false)
  const ws = ref<WebSocket | null>(null)

  const runningCount = computed(() => services.value.filter(s => s.state === 'running').length)
  const stoppedCount = computed(() => services.value.filter(s => s.state === 'stopped').length)
  const failedCount = computed(() => services.value.filter(s => s.state === 'failed').length)
  const totalCount = computed(() => services.value.length)

  async function fetchServices() {
    loading.value = true
    error.value = null
    try {
      services.value = await api.listServices() || []
    } catch (e: any) {
      error.value = e.message || 'Failed to fetch services'
      services.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchDaemonStatus() {
    try {
      daemonStatus.value = await api.getDaemonStatus()
    } catch {
      daemonStatus.value = null
    }
  }

  async function startService(name: string) {
    await api.startService(name)
    await fetchServices()
  }

  async function stopService(name: string) {
    await api.stopService(name)
    await fetchServices()
  }

  async function restartService(name: string) {
    await api.restartService(name)
    await fetchServices()
  }

  async function deleteService(name: string) {
    await api.deleteService(name)
    await fetchServices()
  }

  async function startDaemon() {
    daemonToggling.value = true
    try {
      await api.startDaemon()
      await fetchDaemonStatus()
      await fetchServices()
      connectWebSocket()
    } catch (e: any) {
      error.value = e.message || 'Failed to start daemon'
    } finally {
      daemonToggling.value = false
    }
  }

  async function stopDaemon() {
    daemonToggling.value = true
    try {
      disconnectWebSocket()
      await api.stopDaemon()
      daemonStatus.value = null
      services.value = []
    } catch (e: any) {
      error.value = e.message || 'Failed to stop daemon'
    } finally {
      daemonToggling.value = false
    }
  }

  function connectWebSocket() {
    ws.value = api.connectWebSocket((event) => {
      // Auto-refresh on relevant events
      if (event.type && event.type.startsWith('service.') && event.type !== 'service.log') {
        fetchServices()
      }
    })
  }

  function disconnectWebSocket() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  return {
    services,
    daemonStatus,
    loading,
    error,
    runningCount,
    stoppedCount,
    failedCount,
    totalCount,
    fetchServices,
    fetchDaemonStatus,
    startService,
    stopService,
    restartService,
    deleteService,
    daemonToggling,
    startDaemon,
    stopDaemon,
    connectWebSocket,
    disconnectWebSocket
  }
})
