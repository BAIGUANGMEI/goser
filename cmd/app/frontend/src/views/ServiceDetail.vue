<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type ServiceInfo, type LogEntry, type ServiceConfig } from '@/api/wails'
import StatusBadge from '@/components/StatusBadge.vue'
import LogViewer from '@/components/LogViewer.vue'
import ConfigEditor from '@/components/ConfigEditor.vue'

const props = defineProps<{ name: string }>()
const router = useRouter()

const service = ref<ServiceInfo | null>(null)
const logs = ref<LogEntry[]>([])
const loading = ref(true)
const activeTab = ref<'overview' | 'config' | 'logs'>('overview')
const editError = ref('')
const editSuccess = ref('')
const configEditorRef = ref<InstanceType<typeof ConfigEditor> | null>(null)

let pollTimer: ReturnType<typeof setInterval>

// Snapshot the config ONCE when switching to config tab, not reactively
const configSnapshot = ref<Partial<ServiceConfig>>({})

function takeConfigSnapshot() {
  if (!service.value) return
  configSnapshot.value = {
    name: service.value.name,
    command: service.value.command,
    args: [...(service.value.args || [])],
    working_dir: service.value.working_dir || '',
    env: { ...(service.value.env || {}) },
    auto_start: service.value.auto_start,
    auto_restart: service.value.auto_restart,
  }
}

async function fetchData() {
  // Skip polling when user is editing config to avoid disrupting the form
  if (activeTab.value === 'config') return

  try {
    service.value = await api.getService(props.name)
    if (activeTab.value === 'logs') {
      logs.value = await api.getLogs(props.name, 300) || []
    }
  } catch { /* ignore */ }
  finally { loading.value = false }
}

async function handleStart() {
  await api.startService(props.name); await fetchData()
}
async function handleStop() {
  await api.stopService(props.name); await fetchData()
}
async function handleRestart() {
  await api.restartService(props.name); await fetchData()
}
async function handleDelete() {
  if (confirm(`Remove service "${props.name}"? This will stop the process and delete its configuration.`)) {
    await api.deleteService(props.name)
    router.push('/services')
  }
}
async function handleSaveConfig(config: ServiceConfig) {
  editError.value = ''
  editSuccess.value = ''
  try {
    await api.updateService(props.name, config)
    editSuccess.value = 'Configuration saved successfully'
    setTimeout(() => editSuccess.value = '', 3000)
    // Refresh service data and update the editor snapshot
    service.value = await api.getService(props.name)
    takeConfigSnapshot()
    if (configEditorRef.value) {
      configEditorRef.value.resetFromConfig(configSnapshot.value)
    }
  } catch (e: any) {
    editError.value = e.message || 'Failed to save'
  }
}
async function switchTab(tab: 'overview' | 'config' | 'logs') {
  if (tab === 'config') {
    // Fetch fresh data before entering config tab, then snapshot it
    service.value = await api.getService(props.name)
    takeConfigSnapshot()
  }
  activeTab.value = tab
  if (tab === 'logs') logs.value = await api.getLogs(props.name, 300) || []
  if (tab !== 'config') {
    // Leaving config tab: resume polling by fetching immediately
    loading.value = false
    await fetchData()
  }
}

onMounted(async () => {
  // Initial fetch (activeTab is 'overview' so fetchData will run)
  try {
    service.value = await api.getService(props.name)
  } catch { /* ignore */ }
  finally { loading.value = false }
  pollTimer = setInterval(fetchData, 3000)
})
onUnmounted(() => clearInterval(pollTimer))
</script>

<template>
  <div class="p-7">
    <!-- Back + Header -->
    <div class="flex items-start gap-3 mb-6">
      <button @click="router.push('/services')"
        class="mt-1 p-1.5 hover:bg-gray-100 rounded-lg transition-colors shrink-0">
        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M15 19l-7-7 7-7"/></svg>
      </button>
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-3">
          <h2 class="text-[22px] font-bold text-gray-800 truncate">{{ name }}</h2>
          <StatusBadge v-if="service" :state="service.state" />
        </div>
        <p v-if="service" class="text-[12px] text-gray-400 mt-0.5 font-mono truncate">
          {{ service.command }} {{ (service.args || []).join(' ') }}
        </p>
      </div>
      <div class="flex gap-2 shrink-0">
        <button v-if="service && service.state !== 'running' && service.state !== 'starting'" @click="handleStart"
          class="px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-white text-[12px] font-semibold rounded-lg transition-colors shadow-sm shadow-emerald-200 flex items-center gap-1.5">
          <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>Start
        </button>
        <button v-if="service?.state === 'running'" @click="handleStop"
          class="px-4 py-2 bg-red-500 hover:bg-red-600 text-white text-[12px] font-semibold rounded-lg transition-colors shadow-sm shadow-red-200 flex items-center gap-1.5">
          <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24"><rect x="6" y="6" width="12" height="12" rx="1"/></svg>Stop
        </button>
        <button v-if="service?.state === 'running'" @click="handleRestart"
          class="px-4 py-2 bg-white hover:bg-gray-50 text-gray-600 text-[12px] font-semibold rounded-lg border border-gray-200 transition-colors flex items-center gap-1.5">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15"/></svg>
          Restart
        </button>
        <button @click="handleDelete"
          class="px-3 py-2 hover:bg-red-50 text-gray-400 hover:text-red-500 text-[12px] font-semibold rounded-lg transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-0.5 mb-6 bg-gray-100 p-0.5 rounded-xl w-fit">
      <button v-for="tab in [
        { key: 'overview', label: 'Overview' },
        { key: 'config', label: 'Configuration' },
        { key: 'logs', label: 'Logs' },
      ]" :key="tab.key"
        @click="switchTab(tab.key as any)"
        class="px-4 py-[7px] text-[12px] font-semibold rounded-[10px] transition-all"
        :class="activeTab === tab.key
          ? 'bg-white text-gray-800 shadow-sm'
          : 'text-gray-400 hover:text-gray-600'">
        {{ tab.label }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-20 text-gray-300 text-[13px]">Loading...</div>

    <!-- Overview Tab -->
    <div v-else-if="activeTab === 'overview' && service" class="space-y-5">
      <!-- Process Info Grid -->
      <div class="card p-5">
        <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-wider mb-4">Process Information</h3>
        <div class="grid grid-cols-4 gap-4">
          <div v-for="item in [
            { label: 'PID', value: service.pid || '--', mono: true },
            { label: 'Uptime', value: service.uptime || '--' },
            { label: 'Restarts', value: service.restart_count },
            { label: 'Exit Code', value: service.exit_code ?? '--', mono: true },
            { label: 'Started', value: service.started_at ? new Date(service.started_at).toLocaleString() : '--' },
            { label: 'Stopped', value: service.stopped_at ? new Date(service.stopped_at).toLocaleString() : '--' },
            { label: 'Auto Start', value: service.auto_start ? 'Enabled' : 'Disabled' },
            { label: 'Auto Restart', value: service.auto_restart ? 'Enabled' : 'Disabled' },
          ]" :key="item.label" class="bg-gray-50/70 rounded-lg px-3 py-2.5">
            <span class="block text-[10px] text-gray-400 font-semibold uppercase tracking-wider mb-0.5">{{ item.label }}</span>
            <span class="text-[13px] font-bold text-gray-700" :class="item.mono ? 'font-mono' : ''">{{ item.value }}</span>
          </div>
        </div>
      </div>

      <!-- Command -->
      <div class="card p-5">
        <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-wider mb-3">Command</h3>
        <div class="bg-gray-50 rounded-lg px-4 py-3 font-mono text-[13px] text-gray-700">
          {{ service.command }} {{ (service.args || []).join(' ') }}
        </div>
        <div v-if="service.working_dir" class="mt-3">
          <span class="text-[10px] text-gray-400 font-semibold uppercase tracking-wider">Working Directory</span>
          <div class="bg-gray-50 rounded-lg px-4 py-2.5 font-mono text-[12px] text-gray-600 mt-1">{{ service.working_dir }}</div>
        </div>
      </div>

      <!-- Environment -->
      <div v-if="service.env && Object.keys(service.env).length" class="card p-5">
        <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-wider mb-3">Environment Variables</h3>
        <div class="space-y-1.5">
          <div v-for="(val, key) in service.env" :key="key"
            class="flex items-center bg-gray-50 rounded-lg px-4 py-2 font-mono text-[12px]">
            <span class="text-indigo-600 font-semibold min-w-[140px]">{{ key }}</span>
            <span class="text-gray-300 mx-2">=</span>
            <span class="text-gray-600 truncate">{{ val }}</span>
          </div>
        </div>
      </div>

      <!-- Error -->
      <div v-if="service.error" class="card border-red-200 bg-red-50 p-5">
        <h3 class="text-[11px] font-bold text-red-500 uppercase tracking-wider mb-2">Error</h3>
        <p class="text-[13px] text-red-600 font-mono">{{ service.error }}</p>
      </div>
    </div>

    <!-- Config Tab -->
    <div v-else-if="activeTab === 'config'" class="card p-6">
      <div class="mb-5">
        <h3 class="text-[15px] font-bold text-gray-800">Edit Configuration</h3>
        <p class="text-[12px] text-gray-400 mt-0.5">Modify service parameters. Changes take effect on next restart.</p>
      </div>
      <div v-if="editError" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-[12px] text-red-600">{{ editError }}</div>
      <div v-if="editSuccess" class="mb-4 p-3 bg-emerald-50 border border-emerald-200 rounded-lg text-[12px] text-emerald-600 flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M5 13l4 4L19 7"/></svg>
        {{ editSuccess }}
      </div>
      <ConfigEditor ref="configEditorRef" :config="configSnapshot" mode="edit" @save="handleSaveConfig" @cancel="switchTab('overview')" />
    </div>

    <!-- Logs Tab -->
    <div v-else-if="activeTab === 'logs'" class="h-[calc(100vh-210px)]">
      <LogViewer :logs="logs" />
    </div>
  </div>
</template>
