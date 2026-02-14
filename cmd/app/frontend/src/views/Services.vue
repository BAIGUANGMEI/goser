<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useServiceStore } from '@/stores/services'
import ServiceCard from '@/components/ServiceCard.vue'
import ConfigEditor from '@/components/ConfigEditor.vue'
import type { ServiceConfig } from '@/api/wails'
import { api } from '@/api/wails'

const store = useServiceStore()
const router = useRouter()

const search = ref('')
const filter = ref<'all' | 'running' | 'stopped' | 'failed'>('all')
const showCreateModal = ref(false)
const error = ref('')

const filteredServices = computed(() => {
  let list = store.services
  if (filter.value !== 'all') {
    list = list.filter(s => s.state === filter.value)
  }
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(s => s.name.toLowerCase().includes(q) || s.command.toLowerCase().includes(q))
  }
  return list
})

const filterCounts = computed(() => ({
  all: store.services.length,
  running: store.services.filter(s => s.state === 'running').length,
  stopped: store.services.filter(s => s.state === 'stopped').length,
  failed: store.services.filter(s => s.state === 'failed').length,
}))

async function handleCreate(config: ServiceConfig) {
  error.value = ''
  try {
    await api.createService(config)
    showCreateModal.value = false
    await store.fetchServices()
  } catch (e: any) {
    error.value = e.message || 'Failed to create service'
  }
}
</script>

<template>
  <div class="p-7">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-[22px] font-bold text-gray-800">Services</h2>
        <p class="text-[13px] text-gray-400 mt-0.5">Manage and configure service processes</p>
      </div>
      <button @click="showCreateModal = true"
        class="flex items-center gap-2 px-4 py-2.5 bg-indigo-500 hover:bg-indigo-600 text-white text-[12px] font-semibold rounded-xl transition-colors shadow-sm shadow-indigo-200">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M12 4v16m8-8H4"/></svg>
        Add Service
      </button>
    </div>

    <!-- Filters + Search -->
    <div class="flex items-center gap-4 mb-5">
      <div class="flex bg-white rounded-xl border border-gray-200 p-0.5">
        <button v-for="f in (['all', 'running', 'stopped', 'failed'] as const)" :key="f"
          @click="filter = f"
          class="px-3.5 py-1.5 text-[11px] font-semibold rounded-lg transition-colors capitalize"
          :class="filter === f ? 'bg-indigo-50 text-indigo-600 shadow-sm' : 'text-gray-400 hover:text-gray-600'">
          {{ f }} <span class="ml-0.5 text-[10px] opacity-60">({{ filterCounts[f] }})</span>
        </button>
      </div>
      <div class="flex-1 relative">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-300" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
        </svg>
        <input v-model="search" type="text" placeholder="Search services..."
          class="w-full pl-10 pr-4 py-2.5 text-[13px] rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow" />
      </div>
    </div>

    <!-- Grid -->
    <div v-if="filteredServices.length === 0 && !store.loading" class="card p-16 text-center">
      <svg class="w-12 h-12 mx-auto text-gray-200 mb-3" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/>
      </svg>
      <p class="text-[13px] text-gray-400">{{ search ? 'No services match your search' : 'No services yet' }}</p>
      <p class="text-[11px] text-gray-300 mt-1" v-if="!search">Click "Add Service" to get started</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <ServiceCard
        v-for="svc in filteredServices"
        :key="svc.name"
        :service="svc"
        @start="store.startService"
        @stop="store.stopService"
        @restart="store.restartService"
        @click="(name: string) => router.push(`/services/${name}`)"
      />
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50">
        <div class="bg-white rounded-2xl shadow-2xl w-full max-w-[650px] p-6 m-4 max-h-[85vh] overflow-auto">
          <div class="mb-5">
            <h3 class="text-[18px] font-bold text-gray-800">Add New Service</h3>
            <p class="text-[12px] text-gray-400 mt-0.5">Configure a new process to manage</p>
          </div>
          <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-[12px] text-red-600">{{ error }}</div>
          <ConfigEditor mode="create" @save="handleCreate" @cancel="showCreateModal = false" />
        </div>
      </div>
    </Teleport>
  </div>
</template>
