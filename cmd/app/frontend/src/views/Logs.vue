<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useServiceStore } from '@/stores/services'
import { useLogStore } from '@/stores/logs'
import LogViewer from '@/components/LogViewer.vue'

const serviceStore = useServiceStore()
const logStore = useLogStore()
const selectedService = ref('')

onMounted(() => {
  if (serviceStore.services.length === 0) serviceStore.fetchServices()
})

watch(selectedService, (name) => {
  if (name) logStore.fetchLogs(name, 500)
  else logStore.clearLogs()
})

function refresh() {
  if (selectedService.value) logStore.fetchLogs(selectedService.value, 500)
}
</script>

<template>
  <div class="p-7 h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between mb-5">
      <div>
        <h2 class="text-[22px] font-bold text-gray-800">Logs</h2>
        <p class="text-[13px] text-gray-400 mt-0.5">Real-time service output</p>
      </div>
      <div class="flex items-center gap-3">
        <select v-model="selectedService"
          class="px-3 py-2.5 text-[13px] rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 min-w-[220px] font-medium text-gray-600">
          <option value="">Select a service...</option>
          <option v-for="svc in serviceStore.services" :key="svc.name" :value="svc.name">
            {{ svc.name }} ({{ svc.state }})
          </option>
        </select>
        <button @click="refresh" :disabled="!selectedService"
          class="p-2.5 rounded-xl border border-gray-200 bg-white hover:bg-gray-50 text-gray-400 hover:text-gray-600 transition-colors disabled:opacity-40">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 min-h-0">
      <div v-if="!selectedService" class="card flex items-center justify-center h-full">
        <div class="text-center py-16">
          <svg class="w-10 h-10 mx-auto text-gray-200 mb-3" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <rect x="2" y="4" width="20" height="16" rx="2"/><path d="M7 9l3 3-3 3"/><path d="M13 15h4"/>
          </svg>
          <p class="text-[13px] text-gray-400">Select a service to view logs</p>
        </div>
      </div>
      <LogViewer v-else :logs="logStore.logs" :loading="logStore.loading" />
    </div>
  </div>
</template>
