<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useServiceStore } from '@/stores/services'
import logoUrl from '@/assets/logo.png'

const route = useRoute()
const store = useServiceStore()

const navItems = [
  { path: '/', label: 'Dashboard', icon: 'grid' },
  { path: '/services', label: 'Services', icon: 'server' },
  { path: '/logs', label: 'Logs', icon: 'terminal' },
  { path: '/settings', label: 'Settings', icon: 'settings' }
]

function isActive(item: { path: string }): boolean {
  if (item.path === '/') return route.path === '/'
  return route.path.startsWith(item.path)
}

let pollTimer: ReturnType<typeof setInterval>

onMounted(() => {
  store.fetchServices()
  store.fetchDaemonStatus()
  store.connectWebSocket()
  pollTimer = setInterval(() => {
    store.fetchServices()
    store.fetchDaemonStatus()
  }, 5000)
})

onUnmounted(() => {
  store.disconnectWebSocket()
  clearInterval(pollTimer)
})
</script>

<template>
  <div class="flex h-screen bg-[#f5f6fa] text-gray-800 select-none">
    <!-- Sidebar -->
    <aside class="w-[220px] bg-white border-r border-gray-200/80 flex flex-col">
      <!-- Brand -->
      <div class="wails-drag px-5 pt-5 pb-4">
        <div class="flex items-center gap-2.5">
          <img :src="logoUrl" alt="GoSer" class="w-8 h-8 rounded-lg" />
          <div>
            <h1 class="text-[15px] font-bold text-gray-800 tracking-tight">GoSer</h1>
            <p class="text-[10px] text-gray-400 font-medium -mt-0.5">Service Manager</p>
          </div>
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-3 py-2 space-y-0.5">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="flex items-center gap-2.5 px-3 py-[9px] rounded-[10px] text-[13px] font-medium transition-all duration-150"
          :class="isActive(item)
            ? 'bg-indigo-50 text-indigo-600 shadow-sm shadow-indigo-100'
            : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'"
        >
          <svg v-if="item.icon === 'grid'" class="w-[16px] h-[16px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>
          </svg>
          <svg v-else-if="item.icon === 'server'" class="w-[16px] h-[16px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <rect x="2" y="3" width="20" height="7" rx="2"/><rect x="2" y="14" width="20" height="7" rx="2"/><circle cx="6" cy="6.5" r="1"/><circle cx="6" cy="17.5" r="1"/>
          </svg>
          <svg v-else-if="item.icon === 'terminal'" class="w-[16px] h-[16px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <rect x="2" y="4" width="20" height="16" rx="2"/><path d="M7 9l3 3-3 3"/><path d="M13 15h4"/>
          </svg>
          <svg v-else-if="item.icon === 'settings'" class="w-[16px] h-[16px]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
          </svg>
          {{ item.label }}
        </router-link>
      </nav>

      <!-- Daemon Control -->
      <div class="px-4 py-3.5 border-t border-gray-100 space-y-2.5">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="relative flex h-2 w-2 shrink-0">
              <span v-if="store.daemonStatus" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-2 w-2" :class="store.daemonStatus ? 'bg-green-500' : 'bg-gray-300'"></span>
            </span>
            <span class="text-[11px] font-medium" :class="store.daemonStatus ? 'text-gray-600' : 'text-gray-400'">
              Daemon
            </span>
          </div>
          <!-- Toggle Switch -->
          <button
            @click="store.daemonStatus ? store.stopDaemon() : store.startDaemon()"
            :disabled="store.daemonToggling"
            class="relative inline-flex h-[20px] w-[36px] shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out focus:ring-2 focus:ring-indigo-300 focus:ring-offset-1 disabled:opacity-50 disabled:cursor-wait"
            :class="store.daemonStatus ? 'bg-emerald-500' : 'bg-gray-300'"
          >
            <span
              class="pointer-events-none inline-block h-[16px] w-[16px] transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out mt-[2px]"
              :class="store.daemonStatus ? 'translate-x-[18px]' : 'translate-x-[2px]'"
            >
              <!-- Loading spinner when toggling -->
              <svg v-if="store.daemonToggling" class="w-2.5 h-2.5 m-[3px] animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
            </span>
          </button>
        </div>
        <div v-if="store.daemonStatus" class="flex items-center justify-between pl-4">
          <span class="text-[10px] text-gray-400">Uptime</span>
          <span class="text-[10px] text-gray-400 font-mono">{{ store.daemonStatus.uptime }}</span>
        </div>
      </div>
    </aside>

    <!-- Main -->
    <main class="flex-1 overflow-auto">
      <router-view />
    </main>
  </div>
</template>
