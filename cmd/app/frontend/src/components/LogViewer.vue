<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { LogEntry } from '@/api/wails'

const props = defineProps<{
  logs: LogEntry[]
  loading?: boolean
}>()

const container = ref<HTMLDivElement | null>(null)
const autoScroll = ref(true)

watch(() => props.logs.length, async () => {
  if (autoScroll.value) {
    await nextTick()
    if (container.value) {
      container.value.scrollTop = container.value.scrollHeight
    }
  }
})

function formatTime(ts: string): string {
  try { return new Date(ts).toLocaleTimeString('en-US', { hour12: false }) }
  catch { return '' }
}

// Detect truly error-level lines by content, not by stream.
// Many programs (Python, uvicorn, npm) write normal logs to stderr.
const errorPatterns = /\b(error|fatal|panic|exception|traceback|failed|critical)\b/i

function isErrorLine(entry: LogEntry): boolean {
  return errorPatterns.test(entry.line)
}
</script>

<template>
  <div class="flex flex-col h-full rounded-xl overflow-hidden border border-gray-200 bg-white">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-4 py-2.5 bg-gray-50 border-b border-gray-200">
      <div class="flex items-center gap-2">
        <svg class="w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <rect x="2" y="4" width="20" height="16" rx="2"/><path d="M7 9l3 3-3 3"/>
        </svg>
        <span class="text-[12px] text-gray-500 font-semibold">Output</span>
        <span class="text-[10px] text-gray-300 font-mono">{{ logs.length }} lines</span>
      </div>
      <label class="flex items-center gap-1.5 text-[11px] text-gray-400 cursor-pointer">
        <input type="checkbox" v-model="autoScroll" class="w-3 h-3 rounded border-gray-300 text-indigo-500 focus:ring-indigo-500" />
        Auto-scroll
      </label>
    </div>

    <!-- Log Content -->
    <div
      ref="container"
      class="flex-1 overflow-auto bg-[#1e1e2e] p-3 min-h-[200px] max-h-[600px]"
    >
      <div v-if="loading" class="text-gray-500 text-center py-12 text-[12px]">Loading logs...</div>
      <div v-else-if="logs.length === 0" class="text-gray-600 text-center py-12 text-[12px]">No log output yet</div>
      <div v-else>
        <div
          v-for="(entry, i) in logs"
          :key="i"
          class="log-line whitespace-pre-wrap break-all"
        >
          <span class="text-gray-600 select-none">{{ formatTime(entry.timestamp) }}</span>
          <span class="text-gray-600 select-none mx-1">|</span>
          <span :class="isErrorLine(entry) ? 'text-red-400' : 'text-gray-300'">{{ entry.line }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
