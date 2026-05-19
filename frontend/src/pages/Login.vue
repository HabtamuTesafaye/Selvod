<script setup>
import { ref } from 'vue'
import { listVideos } from '../api/videos'
import { saveCredentials, clearCredentials } from '../lib/credentials'
import { Activity, Eye, EyeOff, Loader2, AlertTriangle } from 'lucide-vue-next'

const emit = defineEmits(['success', 'notify'])

const apiKey = ref('')
const playbackKey = ref('')
const rememberMe = ref(false)

const showAdminKey = ref(false)
const showPlaybackKey = ref(false)
const isLoggingIn = ref(false)
const loginError = ref('')

const login = async () => {
  if (!apiKey.value && !playbackKey.value) {
    loginError.value = "Admin API Key and Playback Scope Key are required."
    emit('notify', { message: "Admin API Key and Playback Scope Key are required.", type: "error" })
    return
  }
  if (!apiKey.value) {
    loginError.value = "Admin API Key is required."
    emit('notify', { message: "Admin API Key is required.", type: "error" })
    return
  }
  if (!playbackKey.value) {
    loginError.value = "Playback Scope Key is required."
    emit('notify', { message: "Playback Scope Key is required.", type: "error" })
    return
  }
  
  loginError.value = ''
  isLoggingIn.value = true
  
  // Save credentials temporarily so the listVideos request is sent with the correct headers
  saveCredentials({ 
    adminKey: apiKey.value, 
    playbackKey: playbackKey.value, 
    rememberMe: rememberMe.value 
  })
  
  try {
    const data = await listVideos()
    emit('success', { videos: data.videos || [], apiKey: apiKey.value, playbackKey: playbackKey.value })
  } catch (err) {
    clearCredentials()
    loginError.value = "Invalid Admin API Key or Playback Key. Access Denied."
    emit('notify', { message: "Invalid credentials. Access Denied.", type: "error" })
  } finally {
    isLoggingIn.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 z-[200] bg-slate-50 dark:bg-[#111318] flex items-center justify-center p-4">
    <div class="w-full max-w-md bg-white dark:bg-[#1a1d24] p-8 rounded-3xl shadow-2xl border border-slate-200 dark:border-[#2d3139]">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-rose-500/10 text-rose-500 rounded-2xl flex items-center justify-center mx-auto mb-4">
          <Activity class="w-8 h-8" />
        </div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white mb-2">Selvod Console</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400">Enter your administrative credentials.</p>
      </div>
      
      <div class="space-y-5">
        <div>
          <label class="block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5">
            Admin Secret Key <span class="text-rose-500 font-bold">*</span>
          </label>
          <div class="relative">
            <input 
              v-model="apiKey" 
              @keyup.enter="login" 
              :type="showAdminKey ? 'text' : 'password'" 
              class="w-full pl-4 pr-11 py-3 bg-slate-50 dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors placeholder-slate-400 text-sm" 
              placeholder="••••••••" 
            />
            <button 
              type="button"
              @click="showAdminKey = !showAdminKey" 
              class="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
            >
              <Eye v-if="!showAdminKey" class="w-5 h-5" />
              <EyeOff v-else class="w-5 h-5" />
            </button>
          </div>
        </div>

        <div>
          <label class="block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5">
            Playback Scope Key <span class="text-rose-500 font-bold">*</span>
          </label>
          <div class="relative">
            <input 
              v-model="playbackKey" 
              @keyup.enter="login" 
              :type="showPlaybackKey ? 'text' : 'password'" 
              class="w-full pl-4 pr-11 py-3 bg-slate-50 dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors placeholder-slate-400 text-sm" 
              placeholder="••••••••" 
            />
            <button 
              type="button"
              @click="showPlaybackKey = !showPlaybackKey" 
              class="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
            >
              <Eye v-if="!showPlaybackKey" class="w-5 h-5" />
              <EyeOff v-else class="w-5 h-5" />
            </button>
          </div>
        </div>

        <div v-if="loginError" class="p-3.5 bg-rose-500/10 border border-rose-500/20 rounded-xl text-rose-500 text-sm font-medium flex items-start gap-2.5">
          <AlertTriangle class="w-5 h-5 shrink-0 mt-0.5" />
          <span>{{ loginError }}</span>
        </div>

        <label class="flex items-center gap-3 cursor-pointer group mt-4">
          <div class="relative flex items-center justify-center w-5 h-5 border-2 border-slate-300 dark:border-[#2d3139] rounded bg-white dark:bg-[#111318] group-hover:border-primary transition-colors">
            <input type="checkbox" v-model="rememberMe" class="opacity-0 absolute inset-0 cursor-pointer" />
            <div v-if="rememberMe" class="w-2.5 h-2.5 bg-primary rounded-sm"></div>
          </div>
          <span class="text-sm text-slate-600 dark:text-slate-300 select-none">Remember my credentials</span>
        </label>

        <button 
          @click="login" 
          :disabled="isLoggingIn"
          class="w-full mt-6 py-3 bg-primary text-white font-medium rounded-xl hover:shadow-lg hover:shadow-primary/30 hover:opacity-90 transition-all active:scale-[0.98] cursor-pointer flex justify-center items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Loader2 v-if="isLoggingIn" class="w-5 h-5 animate-spin" />
          {{ isLoggingIn ? 'Verifying Credentials...' : 'Enter Console' }}
        </button>
      </div>
    </div>
  </div>
</template>
