<script setup>
import { Upload, Search } from 'lucide-vue-next'

defineProps({
  activeFilter: { type: String, required: true },
  searchQuery: { type: String, default: '' }
})

defineEmits(['update:activeFilter', 'update:searchQuery', 'upload'])
</script>

<template>
  <div class="mt-8 mb-6 border-b border-slate-200 dark:border-[#2d3139] flex justify-between items-end pb-px">
    <div class="flex gap-6">
      <button
        :class="[
          'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
          activeFilter === 'all'
            ? 'text-primary'
            : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
        ]"
        @click="$emit('update:activeFilter', 'all')"
      >
        All Content
        <div
          v-if="activeFilter === 'all'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"
        />
      </button>

      <button
        :class="[
          'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
          activeFilter === 'published'
            ? 'text-primary'
            : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
        ]"
        @click="$emit('update:activeFilter', 'published')"
      >
        Published
        <div
          v-if="activeFilter === 'published'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"
        />
      </button>

      <button
        :class="[
          'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
          activeFilter === 'processing'
            ? 'text-primary'
            : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
        ]"
        @click="$emit('update:activeFilter', 'processing')"
      >
        Processing
        <div
          v-if="activeFilter === 'processing'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"
        />
      </button>
    </div>

    <div class="flex items-center gap-3 mb-2">
      <div class="relative hidden sm:block">
        <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
        <input
          :value="searchQuery"
          type="text"
          placeholder="Search library..."
          class="pl-9 pr-4 py-2 w-48 md:w-64 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
          @input="$emit('update:searchQuery', $event.target.value)"
        >
      </div>
      <button
        class="px-4 py-2 bg-primary hover:bg-rose-600 text-white rounded-lg text-sm font-semibold transition-colors flex items-center gap-2 cursor-pointer shadow-sm shrink-0"
        @click="$emit('upload')"
      >
        <Upload class="w-4 h-4" />
        Upload File
      </button>
    </div>
  </div>
</template>
