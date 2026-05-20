<script setup>
import { Settings } from 'lucide-vue-next'
import ThemeToggle from '../ThemeToggle.vue'

defineProps({
  activeTab: { type: String, required: true }
})

defineEmits(['openSettings'])
</script>

<template>
  <header class="h-20 border-b border-slate-200 dark:border-[#2d3139] px-8 flex items-center justify-between sticky top-0 bg-slate-50/90 dark:bg-[#111318]/90 backdrop-blur-md z-30 transition-colors duration-300">
    <div class="text-sm font-semibold text-slate-800 dark:text-slate-200">
      Admin Dashboard / <span class="text-primary">{{ activeTab.charAt(0).toUpperCase() + activeTab.slice(1) }}</span>
    </div>

    <div class="flex items-center gap-4">
      <ThemeToggle />
      <button @click="$emit('openSettings')" class="p-2 text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-[#2d3139]/50 rounded-lg transition-colors cursor-pointer" title="System Settings">
        <Settings class="w-5 h-5" />
      </button>
    </div>
  </header>
</template>
