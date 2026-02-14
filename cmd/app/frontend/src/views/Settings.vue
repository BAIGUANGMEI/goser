<script setup lang="ts">
import { ref } from 'vue'
import { useServiceStore } from '@/stores/services'

const store = useServiceStore()

const settings = ref({
  daemonAddress: '127.0.0.1:9876',
  refreshInterval: 5,
})
</script>

<template>
  <div class="p-7 max-w-2xl">
    <div class="mb-7">
      <h2 class="text-[22px] font-bold text-gray-800">Settings</h2>
      <p class="text-[13px] text-gray-400 mt-0.5">Application preferences</p>
    </div>

    <!-- Connection -->
    <div class="card p-5 mb-5">
      <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-wider mb-4">Connection</h3>
      <div class="space-y-4">
        <div>
          <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Daemon Address</label>
          <input v-model="settings.daemonAddress" type="text"
            class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono" />
          <p class="text-[10px] text-gray-400 mt-1.5">The host:port where the GoSer daemon is listening</p>
        </div>
        <div>
          <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Refresh Interval</label>
          <div class="flex items-center gap-2">
            <input v-model.number="settings.refreshInterval" type="number" min="1" max="60"
              class="w-24 px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow" />
            <span class="text-[12px] text-gray-400">seconds</span>
          </div>
        </div>
      </div>
    </div>

    <!-- About -->
    <div class="card p-5">
      <h3 class="text-[11px] font-bold text-gray-400 uppercase tracking-wider mb-4">About</h3>
      <div class="space-y-2.5">
        <div class="flex items-center justify-between text-[13px]">
          <span class="text-gray-500">Version</span>
          <span class="font-mono font-semibold text-gray-700">1.0.0</span>
        </div>
        <div class="flex items-center justify-between text-[13px]">
          <span class="text-gray-500">Daemon</span>
          <span class="font-semibold" :class="store.daemonStatus ? 'text-emerald-500' : 'text-red-400'">
            {{ store.daemonStatus ? 'Connected' : 'Disconnected' }}
          </span>
        </div>
        <div v-if="store.daemonStatus" class="flex items-center justify-between text-[13px]">
          <span class="text-gray-500">Daemon Uptime</span>
          <span class="font-mono text-gray-700">{{ store.daemonStatus.uptime }}</span>
        </div>
        <div class="flex items-center justify-between text-[13px]">
          <span class="text-gray-500">Tech Stack</span>
          <span class="text-gray-700">Go + Wails + Vue 3</span>
        </div>
      </div>
    </div>
  </div>
</template>
