<script setup>
defineProps({
  icon: { required: true },
  label: { type: String, required: true },
  value: { type: String, default: '' },
  badge: { type: String, default: '' },
  badgeColor: { type: String, default: 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-900/30' },
  iconBg: { type: String, default: 'bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400' },
  progress: { type: Number, default: 0 },
  progressLabel: { type: String, default: '' }
})
</script>

<template>
  <div class="bg-white dark:bg-[#1a1d24] p-5 rounded-xl border border-slate-200 dark:border-[#2d3139] shadow-sm flex flex-col justify-between">
    <div class="flex justify-between items-start mb-6">
      <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', iconBg]">
        <component :is="icon" class="w-5 h-5" />
      </div>
      <span v-if="badge" :class="['text-xs font-semibold px-2 py-1 rounded', badgeColor]">{{ badge }}</span>
    </div>

    <div>
      <h3 class="text-slate-500 dark:text-slate-400 text-sm font-medium mb-1">{{ label }}</h3>
      <div class="text-2xl font-bold text-slate-900 dark:text-white">{{ value }}</div>

      <div v-if="progress > 0 && progressLabel" class="mt-2">
        <div class="flex justify-between items-baseline mb-1">
          <span class="text-sm text-slate-500 dark:text-slate-400">{{ progressLabel }}</span>
        </div>
        <div class="w-full bg-slate-100 dark:bg-[#2d3139] h-2 rounded-full overflow-hidden">
          <div class="bg-primary h-full rounded-full" :style="{ width: `${progress}%` }"></div>
        </div>
      </div>
    </div>
  </div>
</template>
