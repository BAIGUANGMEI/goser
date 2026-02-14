<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useServiceStore } from '@/stores/services'
import StatusBadge from '@/components/StatusBadge.vue'

const store = useServiceStore()
const router = useRouter()

const stats = computed(() => [
  { label: 'Total', value: store.totalCount, color: 'from-indigo-500 to-indigo-600', iconBg: 'bg-indigo-400/20' },
  { label: 'Running', value: store.runningCount, color: 'from-emerald-500 to-emerald-600', iconBg: 'bg-emerald-400/20' },
  { label: 'Stopped', value: store.stoppedCount, color: 'from-gray-400 to-gray-500', iconBg: 'bg-gray-400/20' },
  { label: 'Failed', value: store.failedCount, color: 'from-red-500 to-red-600', iconBg: 'bg-red-400/20' },
])

const sortedServices = computed(() =>
  [...store.services].sort((a, b) => {
    const order: Record<string, number> = { running: 0, starting: 1, stopping: 2, failed: 3, stopped: 4 }
    return (order[a.state] ?? 5) - (order[b.state] ?? 5) || a.name.localeCompare(b.name)
  })
)
</script>

<template>
  <div class="p-7">
    <!-- Header -->
    <div class="mb-7">
      <h2 class="text-[22px] font-bold text-gray-800">Dashboard</h2>
      <p class="text-[13px] text-gray-400 mt-0.5">Overview of managed services</p>
    </div>

    <!-- Stats -->
    <div class="grid grid-cols-4 gap-4 mb-7">
      <div v-for="s in stats" :key="s.label" class="card p-4 relative overflow-hidden">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[11px] font-semibold text-gray-400 uppercase tracking-wider">{{ s.label }}</p>
            <p class="text-[28px] font-extrabold text-gray-800 mt-1 leading-none">{{ s.value }}</p>
          </div>
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br flex items-center justify-center" :class="s.color">
            <svg class="w-5 h-5 text-white/90" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/>
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- Daemon Info -->
    <div v-if="store.daemonStatus" class="card p-5 mb-7">
      <div class="flex items-center gap-2 mb-3">
        <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
        <h3 class="text-[12px] font-bold text-gray-600 uppercase tracking-wider">Daemon</h3>
      </div>
      <div class="grid grid-cols-4 gap-6 text-[13px]">
        <div>
          <span class="text-gray-400 text-[11px]">Status</span>
          <p class="font-semibold text-emerald-600 mt-0.5">Running</p>
        </div>
        <div>
          <span class="text-gray-400 text-[11px]">Uptime</span>
          <p class="font-semibold text-gray-700 mt-0.5">{{ store.daemonStatus.uptime }}</p>
        </div>
        <div>
          <span class="text-gray-400 text-[11px]">Listen</span>
          <p class="font-semibold text-gray-700 mt-0.5 font-mono text-[12px]">127.0.0.1:9876</p>
        </div>
        <div>
          <span class="text-gray-400 text-[11px]">Services</span>
          <p class="font-semibold text-gray-700 mt-0.5">
            <span class="text-emerald-600">{{ store.runningCount }}</span> / {{ store.totalCount }}
          </p>
        </div>
      </div>
    </div>

    <div v-else class="card border-red-200 bg-red-50 p-4 mb-7 flex items-center gap-3">
      <div class="w-8 h-8 rounded-lg bg-red-100 flex items-center justify-center shrink-0">
        <svg class="w-4 h-4 text-red-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10"/><path d="M12 8v4m0 4h.01"/>
        </svg>
      </div>
      <div>
        <p class="text-[13px] font-semibold text-red-700">Daemon Offline</p>
        <p class="text-[11px] text-red-500 mt-0.5">Run <code class="bg-red-100 px-1.5 py-0.5 rounded text-[11px] font-mono">goser daemon start</code> to connect</p>
      </div>
    </div>

    <!-- Services table -->
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-[12px] font-bold text-gray-600 uppercase tracking-wider">Services</h3>
      <button @click="router.push('/services')" class="text-[12px] text-indigo-500 hover:text-indigo-600 font-semibold">
        View All &rarr;
      </button>
    </div>

    <div v-if="sortedServices.length === 0" class="card p-12 text-center">
      <svg class="w-10 h-10 mx-auto text-gray-200 mb-3" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/>
      </svg>
      <p class="text-[13px] text-gray-400">No services configured</p>
      <p class="text-[11px] text-gray-300 mt-0.5">Add one via CLI or the Services page</p>
    </div>

    <div v-else class="card overflow-hidden">
      <table class="w-full text-[13px]">
        <thead>
          <tr class="bg-gray-50/70 text-[10px] text-gray-400 uppercase tracking-wider font-semibold">
            <th class="text-left px-4 py-2.5">Name</th>
            <th class="text-left px-4 py-2.5">Status</th>
            <th class="text-left px-4 py-2.5">PID</th>
            <th class="text-left px-4 py-2.5">Uptime</th>
            <th class="text-left px-4 py-2.5">Command</th>
            <th class="text-right px-4 py-2.5">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr
            v-for="svc in sortedServices"
            :key="svc.name"
            class="row-hover cursor-pointer transition-colors"
            @click="router.push(`/services/${svc.name}`)"
          >
            <td class="px-4 py-3 font-semibold text-gray-800">{{ svc.name }}</td>
            <td class="px-4 py-3"><StatusBadge :state="svc.state" size="xs" /></td>
            <td class="px-4 py-3 font-mono text-gray-400 text-[12px]">{{ svc.pid || '--' }}</td>
            <td class="px-4 py-3 text-gray-500">{{ svc.uptime || '--' }}</td>
            <td class="px-4 py-3 font-mono text-gray-400 text-[12px] max-w-[200px] truncate">{{ svc.command }}</td>
            <td class="px-4 py-3 text-right" @click.stop>
              <button v-if="svc.state !== 'running'" @click="store.startService(svc.name)"
                class="text-emerald-500 hover:text-emerald-600 text-[11px] font-semibold mr-3">Start</button>
              <button v-if="svc.state === 'running'" @click="store.stopService(svc.name)"
                class="text-red-500 hover:text-red-600 text-[11px] font-semibold mr-3">Stop</button>
              <button v-if="svc.state === 'running'" @click="store.restartService(svc.name)"
                class="text-indigo-500 hover:text-indigo-600 text-[11px] font-semibold">Restart</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
