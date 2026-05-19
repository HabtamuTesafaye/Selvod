<script setup>
import { 
  LayoutDashboard, 
  Library, 
  BarChart3, 
  Radio, 
  Activity, 
  Settings, 
  HelpCircle, 
  LogOut,
  Upload
} from 'lucide-vue-next'

const props = defineProps({
  activeTab: { type: String, required: true }
})

const emit = defineEmits(['navigate', 'upload'])

const mainLinks = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'library', label: 'Content Library', icon: Library },
  { id: 'analytics', label: 'Analytics', icon: BarChart3 },
  { id: 'live', label: 'Live Streams', icon: Radio },
  { id: 'health', label: 'System Health', icon: Activity },
  { id: 'settings', label: 'Settings', icon: Settings },
]

const bottomLinks = [
  { id: 'support', label: 'Support', icon: HelpCircle },
  { id: 'logout', label: 'Logout', icon: LogOut },
]
</script>

<template>
  <aside class="w-64 flex flex-col bg-white dark:bg-[#1a1d24] border-r border-slate-200 dark:border-[#2d3139] h-screen sticky top-0 transition-colors duration-300">
    <!-- Header / Logo -->
    <div class="h-20 flex items-center px-6">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-primary rounded flex items-center justify-center text-white font-bold text-lg leading-none">
          S
        </div>
        <div>
          <h1 class="font-bold text-slate-900 dark:text-white leading-tight">Selvod</h1>
          <span class="text-xs text-slate-500 dark:text-slate-400 font-medium">VOD Admin</span>
        </div>
      </div>
    </div>

    <!-- Upload Button -->
    <div class="px-4 mb-6">
      <button 
        @click="emit('upload')"
        class="w-full bg-primary hover:bg-rose-600 text-white py-2.5 rounded-lg flex items-center justify-center gap-2 font-medium transition-colors cursor-pointer"
      >
        <Upload class="w-4 h-4" />
        Upload Video
      </button>
    </div>

    <!-- Main Navigation -->
    <nav class="flex-1 px-3 space-y-1 overflow-y-auto">
      <button
        v-for="link in mainLinks"
        :key="link.id"
        @click="emit('navigate', link.id)"
        :class="[
          'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors relative cursor-pointer',
          activeTab === link.id 
            ? 'text-slate-900 dark:text-white bg-slate-100 dark:bg-[#2d3139]/50' 
            : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-50 dark:hover:bg-[#2d3139]/30'
        ]"
      >
        <!-- Active Marker (Left border) -->
        <div v-if="activeTab === link.id" class="absolute left-0 top-1.5 bottom-1.5 w-1 bg-accent rounded-r-full"></div>
        <component :is="link.icon" class="w-5 h-5" />
        {{ link.label }}
      </button>
    </nav>

    <!-- Bottom Navigation -->
    <div class="p-3 border-t border-slate-200 dark:border-[#2d3139] space-y-1">
      <button
        v-for="link in bottomLinks"
        :key="link.id"
        @click="emit('navigate', link.id)"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-50 dark:hover:bg-[#2d3139]/30 transition-colors cursor-pointer"
      >
        <component :is="link.icon" class="w-5 h-5" />
        {{ link.label }}
      </button>
    </div>
  </aside>
</template>
