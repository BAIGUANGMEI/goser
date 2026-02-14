<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { ServiceConfig } from '@/api/wails'

const props = defineProps<{
  config?: Partial<ServiceConfig>
  mode: 'create' | 'edit'
}>()

const emit = defineEmits<{
  save: [config: ServiceConfig]
  cancel: []
}>()

const form = ref<ServiceConfig>({
  name: '',
  command: '',
  args: [],
  working_dir: '',
  env: {},
  auto_start: false,
  auto_restart: true,
  max_restarts: 5,
  restart_delay: 5000000000,
  stop_signal: 'SIGTERM',
  stop_timeout: 10000000000,
  log_file: 'auto',
  depends_on: []
})

const argsText = ref('')
const envText = ref('')
const dependsText = ref('')
const activeSection = ref<'basic' | 'env' | 'restart' | 'advanced'>('basic')
const initialized = ref(false)

const restartDelaySeconds = computed({
  get: () => Math.round((form.value.restart_delay || 5000000000) / 1000000000),
  set: (v: number) => { form.value.restart_delay = v * 1000000000 }
})

const stopTimeoutSeconds = computed({
  get: () => Math.round((form.value.stop_timeout || 10000000000) / 1000000000),
  set: (v: number) => { form.value.stop_timeout = v * 1000000000 }
})

// Only populate the form ONCE when config first arrives.
// Subsequent prop changes from polling are ignored so the user's edits are preserved.
watch(() => props.config, (cfg) => {
  if (cfg && !initialized.value) {
    form.value = { ...form.value, ...cfg }
    argsText.value = (cfg.args || []).join('\n')
    envText.value = Object.entries(cfg.env || {}).map(([k, v]) => `${k}=${v}`).join('\n')
    dependsText.value = (cfg.depends_on || []).join('\n')
    initialized.value = true
  }
}, { immediate: true })

// Allow parent to reset the form (e.g. after a successful save) by exposing a method
function resetFromConfig(cfg: Partial<ServiceConfig>) {
  form.value = { ...form.value, ...cfg }
  argsText.value = (cfg.args || []).join('\n')
  envText.value = Object.entries(cfg.env || {}).map(([k, v]) => `${k}=${v}`).join('\n')
  dependsText.value = (cfg.depends_on || []).join('\n')
}

defineExpose({ resetFromConfig })

function handleSave() {
  form.value.args = argsText.value.split('\n').map(s => s.trim()).filter(Boolean)
  const env: Record<string, string> = {}
  envText.value.split('\n').map(s => s.trim()).filter(Boolean).forEach(line => {
    const idx = line.indexOf('=')
    if (idx > 0) env[line.substring(0, idx)] = line.substring(idx + 1)
  })
  form.value.env = env
  form.value.depends_on = dependsText.value.split('\n').map(s => s.trim()).filter(Boolean)
  emit('save', { ...form.value })
}

const sections = [
  { key: 'basic', label: 'Basic', icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
  { key: 'env', label: 'Environment', icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
  { key: 'restart', label: 'Restart Policy', icon: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15' },
  { key: 'advanced', label: 'Advanced', icon: 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4' },
]
</script>

<template>
  <div class="flex gap-5">
    <!-- Section tabs -->
    <div class="w-[160px] shrink-0 space-y-0.5">
      <button
        v-for="s in sections"
        :key="s.key"
        @click="activeSection = s.key as any"
        class="w-full flex items-center gap-2 px-3 py-2 text-[12px] font-medium rounded-lg transition-colors text-left"
        :class="activeSection === s.key
          ? 'bg-indigo-50 text-indigo-600'
          : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'"
      >
        <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" :d="s.icon"/>
        </svg>
        {{ s.label }}
      </button>
    </div>

    <!-- Form content -->
    <div class="flex-1 min-w-0">
      <!-- Basic -->
      <div v-show="activeSection === 'basic'" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Service Name</label>
            <input v-model="form.name" :disabled="mode === 'edit'" type="text" placeholder="my-service"
              class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow disabled:bg-gray-50 disabled:text-gray-400" />
          </div>
          <div>
            <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Command</label>
            <input v-model="form.command" type="text" placeholder="python, node, java..."
              class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono" />
          </div>
        </div>
        <div>
          <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Arguments <span class="font-normal text-gray-400">(one per line)</span></label>
          <textarea v-model="argsText" rows="3" placeholder="server.js&#10;--host&#10;0.0.0.0"
            class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono resize-none"></textarea>
        </div>
        <div>
          <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Working Directory</label>
          <input v-model="form.working_dir" type="text" placeholder="C:/projects/my-app"
            class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono" />
        </div>
      </div>

      <!-- Environment -->
      <div v-show="activeSection === 'env'" class="space-y-4">
        <div>
          <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Environment Variables <span class="font-normal text-gray-400">(KEY=VALUE per line)</span></label>
          <textarea v-model="envText" rows="6" placeholder="NODE_ENV=production&#10;PORT=3000&#10;DATABASE_URL=postgres://..."
            class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono resize-none"></textarea>
        </div>
        <div class="bg-indigo-50 rounded-lg px-4 py-3">
          <p class="text-[11px] text-indigo-600">System PATH and default environment variables are inherited automatically. Only add variables specific to this service.</p>
        </div>
      </div>

      <!-- Restart Policy -->
      <div v-show="activeSection === 'restart'" class="space-y-4">
        <div class="grid grid-cols-2 gap-6">
          <label class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors">
            <input v-model="form.auto_start" type="checkbox" class="w-4 h-4 rounded border-gray-300 text-indigo-500 focus:ring-indigo-500" />
            <div>
              <span class="text-[13px] font-medium text-gray-700 block">Auto Start</span>
              <span class="text-[11px] text-gray-400">Start when daemon launches</span>
            </div>
          </label>
          <label class="flex items-center gap-3 p-3 rounded-lg border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors">
            <input v-model="form.auto_restart" type="checkbox" class="w-4 h-4 rounded border-gray-300 text-indigo-500 focus:ring-indigo-500" />
            <div>
              <span class="text-[13px] font-medium text-gray-700 block">Auto Restart</span>
              <span class="text-[11px] text-gray-400">Restart on unexpected exit</span>
            </div>
          </label>
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Max Restarts</label>
            <input v-model.number="form.max_restarts" type="number" min="0" max="100"
              class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow" />
            <p class="text-[10px] text-gray-400 mt-1">Stop trying after this many failures</p>
          </div>
          <div>
            <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Restart Delay (seconds)</label>
            <input v-model.number="restartDelaySeconds" type="number" min="1" max="300"
              class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow" />
            <p class="text-[10px] text-gray-400 mt-1">Wait before restarting</p>
          </div>
        </div>
      </div>

      <!-- Advanced -->
      <div v-show="activeSection === 'advanced'" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Stop Timeout (seconds)</label>
            <input v-model.number="stopTimeoutSeconds" type="number" min="1" max="300"
              class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow" />
            <p class="text-[10px] text-gray-400 mt-1">Force kill after timeout</p>
          </div>
          <div>
            <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Log File</label>
            <input v-model="form.log_file" type="text" placeholder="auto"
              class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono" />
            <p class="text-[10px] text-gray-400 mt-1">"auto" uses ~/.goser/logs/&lt;name&gt;.log</p>
          </div>
        </div>
        <div>
          <label class="block text-[11px] font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Dependencies <span class="font-normal text-gray-400">(service names, one per line)</span></label>
          <textarea v-model="dependsText" rows="3" placeholder="database-service&#10;redis"
            class="w-full px-3 py-2 text-[13px] rounded-lg border border-gray-200 bg-white focus:ring-2 focus:ring-indigo-200 focus:border-indigo-400 transition-shadow font-mono resize-none"></textarea>
          <p class="text-[10px] text-gray-400 mt-1">These services will be started first</p>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex justify-end gap-2.5 pt-5 mt-5 border-t border-gray-100">
        <button @click="emit('cancel')"
          class="px-4 py-2 text-[12px] font-semibold text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors">
          Cancel
        </button>
        <button @click="handleSave"
          class="px-5 py-2 text-[12px] font-semibold text-white bg-indigo-500 hover:bg-indigo-600 rounded-lg transition-colors shadow-sm">
          {{ mode === 'create' ? 'Create Service' : 'Save Changes' }}
        </button>
      </div>
    </div>
  </div>
</template>
