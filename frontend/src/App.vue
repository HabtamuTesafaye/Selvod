<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { getAdminKey, clearCredentials } from './lib/credentials'
import Login from './pages/Login.vue'
import Dashboard from './pages/Dashboard.vue'
import { Activity, X } from 'lucide-vue-next'

const isAuthenticated = ref(!!getAdminKey())
const notifications = ref([])

const addNotification = (message, type = 'success') => {
  const id = Date.now() + Math.random().toString(36).substring(2, 9)
  notifications.value.push({ id, message, type })
  setTimeout(() => {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }, 4500)
}

const handleLoginSuccess = () => {
  isAuthenticated.value = true
}

const handleLogout = () => {
  isAuthenticated.value = false
}

const handleNotification = (payload) => {
  addNotification(payload.message, payload.type)
}

const handleUnauthorized = () => {
  clearCredentials()
  isAuthenticated.value = false
}

onMounted(() => {
  window.addEventListener('unauthorized', handleUnauthorized)
})

onUnmounted(() => {
  window.removeEventListener('unauthorized', handleUnauthorized)
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#111318]">
    <Login 
      v-slot="Login"
      v-if="!isAuthenticated" 
      @success="handleLoginSuccess" 
      @notify="handleNotification" 
    />
    <Dashboard 
      v-else 
      @logout="handleLogout" 
      @notify="handleNotification" 
    />

    <!-- Toast Notifications (Bottom Right) -->
    <div class="fixed bottom-6 right-6 z-[250] flex flex-col gap-3 w-80 max-w-full">
      <div 
        v-for="n in notifications" 
        :key="n.id" 
        class="p-4 rounded-xl shadow-lg border text-sm flex items-start gap-3 bg-[#1a1d24] border-[#2d3139] text-white transform transition-all"
      >
        <div 
          :class="['w-6 h-6 rounded-full flex items-center justify-center shrink-0 mt-0.5', 
            n.type === 'success' ? 'bg-emerald-500/20 text-emerald-400' : 
            n.type === 'error' ? 'bg-rose-500/20 text-rose-400' : 
            'bg-blue-500/20 text-blue-400']"
        >
          <Activity class="w-3.5 h-3.5" />
        </div>
        <div class="flex-1">
          <p class="font-semibold text-[13px] mb-0.5">
            {{ n.type === 'success' ? 'System Update' : n.type === 'error' ? 'Task Failed' : 'System Notice' }}
          </p>
          <p class="text-slate-400 text-xs leading-relaxed">{{ n.message }}</p>
        </div>
        <button @click="notifications = notifications.filter(x => x.id !== n.id)" class="text-slate-500 hover:text-white transition-colors">
          <X class="w-4 h-4" />
        </button>
      </div>
    </div>
  </div>
</template>
