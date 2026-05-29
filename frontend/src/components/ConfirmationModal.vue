<script setup>
import { AlertTriangle, Info } from 'lucide-vue-next'

const props = defineProps({
  isOpen: { type: Boolean, required: true },
  title: { type: String, required: true },
  description: { type: String, required: true },
  confirmText: { type: String, default: 'Confirm' },
  cancelText: { type: String, default: 'Cancel' },
  type: { type: String, default: 'danger' } // danger, warning, info
})

const emit = defineEmits(['confirm', 'cancel'])
</script>

<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-[250] flex items-center justify-center p-4 bg-slate-900/15 backdrop-blur-[2px]"
  >
    <div class="bg-white dark:bg-[#1a1d24] rounded-2xl p-6 w-full max-w-sm border border-slate-200 dark:border-[#2d3139] shadow-2xl transform transition-all">
      <div class="flex items-start gap-4 mb-4">
        <!-- Reusable Icon Box based on Type -->
        <div 
          :class="[
            'w-12 h-12 rounded-full flex items-center justify-center shrink-0',
            type === 'danger' ? 'bg-rose-500/10 text-rose-500' :
            type === 'warning' ? 'bg-amber-500/10 text-amber-500' :
            'bg-primary/10 text-primary'
          ]"
        >
          <AlertTriangle
            v-if="type === 'danger' || type === 'warning'"
            class="w-6 h-6"
          />
          <Info
            v-else
            class="w-6 h-6"
          />
        </div>
        
        <div class="flex-1 min-w-0">
          <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-0.5">
            {{ title }}
          </h3>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ description }}
          </p>
        </div>
      </div>

      <div class="flex gap-3 justify-end mt-6">
        <button 
          class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer" 
          @click="emit('cancel')"
        >
          {{ cancelText }}
        </button>
        <button 
          :class="[
            'px-4 py-2 text-white rounded-xl text-sm font-medium transition-colors cursor-pointer shadow-sm',
            type === 'danger' ? 'bg-rose-600 hover:bg-rose-700 shadow-rose-600/20' :
            type === 'warning' ? 'bg-amber-600 hover:bg-amber-700 shadow-amber-600/20' :
            'bg-primary hover:bg-primary/90 shadow-primary/20'
          ]" 
          @click="emit('confirm')"
        >
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>
