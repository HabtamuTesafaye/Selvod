<script setup>
defineProps({
  icon: { required: true },
  label: { type: String, required: true },
  active: { type: Boolean, default: false }
})

defineEmits(['click'])
</script>

<template>
  <button
    @click="$emit('click')"
    :class="[
      'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors relative cursor-pointer',
      active
        ? 'text-slate-900 dark:text-white bg-slate-100 dark:bg-[#2d3139]/50'
        : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-50 dark:hover:bg-[#2d3139]/30'
    ]"
  >
    <div v-if="active" class="absolute left-0 top-1.5 bottom-1.5 w-1 bg-accent rounded-r-full"></div>
    <component :is="icon" class="w-5 h-5 shrink-0" />
    {{ label }}
  </button>
</template>
