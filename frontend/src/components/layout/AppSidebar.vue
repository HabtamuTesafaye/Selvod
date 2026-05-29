<script setup>
import {
  LayoutDashboard,
  Library,
  BarChart3,
  Palette,
  Activity,
  Settings,
  HelpCircle,
  LogOut
} from 'lucide-vue-next'
import NavLink from '../nav/NavLink.vue'

defineProps({
  activeTab: { type: String, required: true }
})

defineEmits(['navigate'])

const mainLinks = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'library', label: 'Content Library', icon: Library },
  { id: 'analytics', label: 'Analytics', icon: BarChart3 },
  { id: 'player-ui', label: 'Player UI', icon: Palette },
  { id: 'health', label: 'System Health', icon: Activity },
  { id: 'settings', label: 'Settings', icon: Settings },
]

const bottomLinks = [
  { id: 'docs', label: 'Developer Docs', icon: HelpCircle },
  { id: 'logout', label: 'Sign Out', icon: LogOut },
]
</script>

<template>
  <aside class="w-64 flex flex-col bg-white dark:bg-[#1a1d24] border-r border-slate-200 dark:border-[#2d3139] h-full transition-colors duration-300">
    <div class="h-20 flex items-center px-6">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-primary rounded flex items-center justify-center text-white font-bold text-lg leading-none">
          S
        </div>
        <div>
          <h1 class="font-bold text-slate-900 dark:text-white leading-tight">
            Selvod
          </h1>
          <span class="text-xs text-slate-500 dark:text-slate-400 font-medium">VOD Admin</span>
        </div>
      </div>
    </div>

    <nav class="flex-1 px-3 space-y-1 overflow-y-auto">
      <NavLink
        v-for="link in mainLinks"
        :key="link.id"
        :icon="link.icon"
        :label="link.label"
        :active="activeTab === link.id"
        @click="$emit('navigate', link.id)"
      />
    </nav>

    <div class="p-3 border-t border-slate-200 dark:border-[#2d3139] space-y-1">
      <NavLink
        v-for="link in bottomLinks"
        :key="link.id"
        :icon="link.icon"
        :label="link.label"
        @click="$emit('navigate', link.id)"
      />
    </div>
  </aside>
</template>
