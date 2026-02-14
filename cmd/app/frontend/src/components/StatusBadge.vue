<script setup lang="ts">
defineProps<{
  state: string
  size?: 'xs' | 'sm' | 'md'
}>()

const styles: Record<string, { bg: string; dot: string }> = {
  running:  { bg: 'bg-emerald-50 text-emerald-700 ring-emerald-200', dot: 'bg-emerald-500' },
  stopped:  { bg: 'bg-gray-50 text-gray-500 ring-gray-200', dot: 'bg-gray-400' },
  failed:   { bg: 'bg-red-50 text-red-600 ring-red-200', dot: 'bg-red-500' },
  starting: { bg: 'bg-amber-50 text-amber-600 ring-amber-200', dot: 'bg-amber-500 animate-pulse' },
  stopping: { bg: 'bg-orange-50 text-orange-600 ring-orange-200', dot: 'bg-orange-500 animate-pulse' },
}

function getStyle(state: string) {
  return styles[state] || styles.stopped
}
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full font-semibold ring-1 ring-inset capitalize"
    :class="[
      getStyle(state).bg,
      size === 'xs' ? 'px-1.5 py-0.5 text-[10px]' :
      size === 'sm' ? 'px-2 py-0.5 text-[11px]' :
      'px-2.5 py-0.5 text-[11px]'
    ]"
  >
    <span class="w-1.5 h-1.5 rounded-full shrink-0" :class="getStyle(state).dot"></span>
    {{ state }}
  </span>
</template>
