<script setup lang="ts">
import StatusBadge from './StatusBadge.vue'
import type { ServiceInfo } from '@/api/wails'

const props = defineProps<{
  service: ServiceInfo
}>()

const emit = defineEmits<{
  start: [name: string]
  stop: [name: string]
  restart: [name: string]
  click: [name: string]
}>()

function cmdDisplay(svc: ServiceInfo): string {
  let cmd = svc.command
  if (svc.args?.length) cmd += ' ' + svc.args.join(' ')
  return cmd.length > 45 ? cmd.substring(0, 42) + '...' : cmd
}
</script>

<template>
  <div
    class="card p-4 hover:shadow-md hover:border-gray-300/80 transition-all duration-200 cursor-pointer group"
    @click="emit('click', service.name)"
  >
    <!-- Header -->
    <div class="flex items-start justify-between mb-3.5">
      <div class="min-w-0 flex-1">
        <h3 class="text-[13px] font-bold text-gray-800 truncate group-hover:text-indigo-600 transition-colors">{{ service.name }}</h3>
        <p class="text-[11px] text-gray-400 mt-0.5 font-mono truncate">{{ cmdDisplay(service) }}</p>
      </div>
      <StatusBadge :state="service.state" size="xs" class="ml-2 shrink-0" />
    </div>

    <!-- Metrics -->
    <div class="grid grid-cols-3 gap-3 mb-3.5">
      <div class="bg-gray-50/80 rounded-lg px-2.5 py-2">
        <span class="block text-[10px] text-gray-400 font-medium uppercase tracking-wider">PID</span>
        <span class="text-[13px] font-bold text-gray-700 font-mono">{{ service.pid || '--' }}</span>
      </div>
      <div class="bg-gray-50/80 rounded-lg px-2.5 py-2">
        <span class="block text-[10px] text-gray-400 font-medium uppercase tracking-wider">Uptime</span>
        <span class="text-[13px] font-bold text-gray-700">{{ service.uptime || '--' }}</span>
      </div>
      <div class="bg-gray-50/80 rounded-lg px-2.5 py-2">
        <span class="block text-[10px] text-gray-400 font-medium uppercase tracking-wider">Restarts</span>
        <span class="text-[13px] font-bold text-gray-700">{{ service.restart_count }}</span>
      </div>
    </div>

    <!-- Actions -->
    <div class="flex gap-1.5" @click.stop>
      <button
        v-if="service.state !== 'running' && service.state !== 'starting'"
        @click="emit('start', service.name)"
        class="flex-1 px-3 py-[7px] bg-emerald-500 hover:bg-emerald-600 text-white text-[11px] font-semibold rounded-lg transition-colors flex items-center justify-center gap-1.5"
      >
        <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
        Start
      </button>
      <button
        v-if="service.state === 'running'"
        @click="emit('stop', service.name)"
        class="flex-1 px-3 py-[7px] bg-red-500 hover:bg-red-600 text-white text-[11px] font-semibold rounded-lg transition-colors flex items-center justify-center gap-1.5"
      >
        <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><rect x="6" y="6" width="12" height="12" rx="1"/></svg>
        Stop
      </button>
      <button
        v-if="service.state === 'running'"
        @click="emit('restart', service.name)"
        class="flex-1 px-3 py-[7px] bg-white hover:bg-gray-50 text-gray-600 text-[11px] font-semibold rounded-lg border border-gray-200 transition-colors flex items-center justify-center gap-1.5"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path d="M1 4v6h6M23 20v-6h-6"/><path d="M20.49 9A9 9 0 005.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 013.51 15"/></svg>
        Restart
      </button>
    </div>
  </div>
</template>
