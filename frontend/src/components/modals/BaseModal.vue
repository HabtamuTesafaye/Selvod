<script setup>
import { X } from 'lucide-vue-next'

defineProps({
  isOpen: { type: Boolean, required: true },
  title: { type: String, default: '' },
  description: { type: String, default: '' },
  maxWidth: { type: String, default: 'max-w-sm' }
})

defineEmits(['close'])
</script>

<template>
  <Teleport to="body">
    <div
      v-if="isOpen"
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/15 backdrop-blur-[2px]"
      @click.self="$emit('close')"
    >
      <div :class="['bg-white dark:bg-[#1a1d24] rounded-2xl p-6 w-full border border-slate-200 dark:border-[#2d3139] shadow-2xl', maxWidth]">
        <div
          v-if="title"
          class="flex justify-between items-center mb-4"
        >
          <h3 class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
            <slot name="icon" />
            {{ title }}
          </h3>
          <button
            class="p-1 hover:bg-slate-100 dark:hover:bg-[#2d3139] rounded-lg text-slate-400 transition-colors cursor-pointer"
            @click="$emit('close')"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <p
          v-if="description"
          class="text-xs text-slate-500 dark:text-slate-400 mb-6"
        >
          {{ description }}
        </p>

        <slot />
      </div>
    </div>
  </Teleport>
</template>
