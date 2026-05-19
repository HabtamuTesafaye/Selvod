<script setup>
import { Upload } from 'lucide-vue-next'

defineProps({
  activeFilter: { type: String, required: true }
})

defineEmits(['update:activeFilter', 'upload'])
</script>

<template>
  <div class="mt-8 mb-6 border-b border-slate-200 dark:border-[#2d3139] flex justify-between items-end pb-px">
    <div class="flex gap-6">
      <button
        @click="$emit('update:activeFilter', 'all')"
        :class="[
          'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
          activeFilter === 'all'
            ? 'text-primary'
            : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
        ]"
      >
        All Content
        <div v-if="activeFilter === 'all'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
      </button>

      <button
        @click="$emit('update:activeFilter', 'published')"
        :class="[
          'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
          activeFilter === 'published'
            ? 'text-primary'
            : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
        ]"
      >
        Published
        <div v-if="activeFilter === 'published'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
      </button>

      <button
        @click="$emit('update:activeFilter', 'processing')"
        :class="[
          'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
          activeFilter === 'processing'
            ? 'text-primary'
            : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
        ]"
      >
        Processing
        <div v-if="activeFilter === 'processing'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
      </button>
    </div>

    <button
      @click="$emit('upload')"
      class="px-4 py-2 bg-primary hover:bg-rose-600 text-white rounded-lg text-sm font-semibold transition-colors flex items-center gap-2 cursor-pointer shadow-sm mb-2"
    >
      <Upload class="w-4 h-4" />
      Upload File
    </button>
  </div>
</template>
