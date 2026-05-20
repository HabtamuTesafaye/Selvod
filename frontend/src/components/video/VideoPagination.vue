<script setup>
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

defineProps({
  currentPage: { type: Number, required: true },
  totalPages: { type: Number, required: true },
  totalItems: { type: Number, required: true },
  itemsPerPage: { type: Number, required: true }
})

defineEmits(['update:currentPage', 'update:itemsPerPage'])
</script>

<template>
  <div class="mt-8 flex flex-col sm:flex-row items-center justify-between gap-4 pt-6 border-t border-slate-200 dark:border-[#2d3139]">
    <div class="flex items-center gap-3 text-sm text-slate-500 dark:text-slate-400">
      <span>Showing {{ (currentPage - 1) * itemsPerPage + 1 }}&ndash;{{ Math.min(currentPage * itemsPerPage, totalItems) }} of {{ totalItems }}</span>
      <select :value="itemsPerPage" @change="$emit('update:itemsPerPage', Number($event.target.value))" class="px-2 py-1 bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-md text-xs cursor-pointer text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-1 focus:ring-primary">
        <option :value="4">4 per page</option>
        <option :value="8">8 per page</option>
        <option :value="12">12 per page</option>
        <option :value="24">24 per page</option>
      </select>
    </div>

    <div class="flex items-center gap-1">
      <button
        @click="$emit('update:currentPage', currentPage - 1)"
        :disabled="currentPage === 1"
        class="w-9 h-9 flex items-center justify-center rounded-lg border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-[#2d3139]/50 disabled:text-slate-300 dark:disabled:text-slate-700 disabled:border-slate-100 dark:disabled:border-[#2d3139]/30 disabled:bg-slate-50/50 dark:disabled:bg-transparent disabled:cursor-not-allowed transition-colors cursor-pointer"
      >
        <ChevronLeft class="w-4 h-4" />
      </button>

      <button
        v-for="page in totalPages"
        :key="page"
        @click="$emit('update:currentPage', page)"
        :class="['w-9 h-9 flex items-center justify-center rounded-lg font-medium transition-colors cursor-pointer text-sm',
          currentPage === page
            ? 'bg-primary text-white'
            : 'border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-[#2d3139]'
        ]"
      >
        {{ page }}
      </button>

      <button
        @click="$emit('update:currentPage', currentPage + 1)"
        :disabled="currentPage === totalPages"
        class="w-9 h-9 flex items-center justify-center rounded-lg border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-[#2d3139]/50 disabled:text-slate-300 dark:disabled:text-slate-700 disabled:border-slate-100 dark:disabled:border-[#2d3139]/30 disabled:bg-slate-50/50 dark:disabled:bg-transparent disabled:cursor-not-allowed transition-colors cursor-pointer"
      >
        <ChevronRight class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>
