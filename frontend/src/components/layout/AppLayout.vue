<script setup>
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

defineProps({
  activeTab: { type: String, required: true }
})

defineEmits(['navigate', 'openSettings'])
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-slate-50 dark:bg-[#111318] text-slate-900 dark:text-slate-100 font-sans transition-colors duration-300">
    <AppSidebar
      :active-tab="activeTab"
      @navigate="$emit('navigate', $event)"
    />

    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">
      <AppHeader
        :active-tab="activeTab"
        @open-settings="$emit('openSettings')"
      />

      <main class="flex-1 px-8 py-8 min-h-0 overflow-hidden">
        <slot />
      </main>
    </div>
  </div>
</template>
