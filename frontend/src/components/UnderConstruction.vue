<script setup>
import { Construction } from 'lucide-vue-next'

defineProps({
  title: {
    type: String,
    default: 'Coming Soon'
  },
  description: {
    type: String,
    default: 'This feature is currently under development. Check back later!'
  }
})
</script>

<template>
  <div class="flex flex-col items-center justify-center py-32 text-center">
    <div class="w-24 h-24 mb-6 rounded-3xl bg-slate-100 dark:bg-slate-800 flex items-center justify-center transform rotate-12 shadow-inner">
      <Construction class="w-12 h-12 text-slate-400 dark:text-slate-500" />
    </div>
    <h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100 mb-2">{{ title }}</h2>
    <p class="text-slate-500 dark:text-slate-400 max-w-sm">{{ description }}</p>
  </div>
</template>
