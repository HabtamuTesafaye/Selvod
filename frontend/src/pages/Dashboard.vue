<script setup>
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { useUploadStore } from '../stores/upload'
import { deleteVideo as deleteVideoRequest, getHealth, listVideos } from '../api/videos'
import { getAdminKey, getPlaybackKey, saveCredentials, clearCredentials } from '../lib/credentials'
import VideoCard from '../components/VideoCard.vue'
import ForgePlayer from '../components/ForgePlayer.vue'
import Sidebar from '../components/Sidebar.vue'
import ThemeToggle from '../components/ThemeToggle.vue'
import UnderConstruction from '../components/UnderConstruction.vue'
import ConfirmationModal from '../components/ConfirmationModal.vue'
import { 
  Activity, 
  Settings, 
  X, 
  ChevronLeft,
  ChevronRight,
  Library,
  HardDrive
} from 'lucide-vue-next'

const emit = defineEmits(['logout', 'notify'])

const uploadStore = useUploadStore()
const videos = ref([])
const health = ref({ status: 'ok', storage_available_bytes: 0 })
const activeVideoId = ref(null)
const isLoading = ref(true)

const activeFilter = ref('all')
const currentPage = ref(1)
const itemsPerPage = ref(12)

const filteredVideos = computed(() => {
  if (activeFilter.value === 'published') {
    return videos.value.filter(v => v.status === 'completed')
  }
  if (activeFilter.value === 'processing') {
    return videos.value.filter(v => v.status === 'transcoding' || v.status === 'pending')
  }
  return videos.value
})

const paginatedVideos = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredVideos.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(filteredVideos.value.length / itemsPerPage.value) || 1
})

const changePage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

const activeTab = ref('library')
const showSettings = ref(false)

const apiKey = ref(getAdminKey())
const playbackKey = ref(getPlaybackKey())
const showAdminKey = ref(false)
const showPlaybackKey = ref(false)

const confirmModalConfig = ref({
  isOpen: false,
  title: '',
  description: '',
  confirmText: 'Confirm',
  cancelText: 'Cancel',
  type: 'danger',
  onConfirm: () => {}
})

const triggerConfirmation = (config) => {
  confirmModalConfig.value = {
    isOpen: true,
    title: config.title,
    description: config.description,
    confirmText: config.confirmText || 'Confirm',
    cancelText: config.cancelText || 'Cancel',
    type: config.type || 'danger',
    onConfirm: config.onConfirm
  }
}

const handleConfirmAction = () => {
  confirmModalConfig.value.isOpen = false
  confirmModalConfig.value.onConfirm()
}

const handleCancelAction = () => {
  confirmModalConfig.value.isOpen = false
}

const handleNavigate = (tab) => {
  if (tab === 'logout') {
    triggerConfirmation({
      title: 'Exit Console',
      description: 'Are you sure you want to log out of the Selvod console? You will need to re-enter your API keys to regain access.',
      confirmText: 'Sign Out',
      cancelText: 'Stay',
      type: 'warning',
      onConfirm: () => {
        logout()
      }
    })
  } else {
    activeTab.value = tab
  }
}

const logout = () => {
  clearCredentials()
  if (videosInterval) {
    clearInterval(videosInterval)
    videosInterval = null
  }
  if (healthInterval) {
    clearInterval(healthInterval)
    healthInterval = null
  }
  emit('logout')
}

const saveSettings = () => {
  saveCredentials({ adminKey: apiKey.value, playbackKey: playbackKey.value, rememberMe: true })
  showSettings.value = false
  emit('notify', { message: "Credentials updated successfully!", type: "success" })
  fetchVideos()
  fetchHealth()
}

const fetchVideos = async () => {
  try {
    const data = await listVideos()
    videos.value = data.videos || []
  } catch (err) {
    console.error("Failed to fetch videos", err)
    emit('notify', { message: "Failed to fetch video library.", type: "error" })
  } finally {
    isLoading.value = false
  }
}

const fetchHealth = async () => {
  try {
    health.value = await getHealth()
  } catch (err) {
    console.error("Health check failed", err)
  }
}

const requestDelete = (id) => {
  triggerConfirmation({
    title: 'Delete Video',
    description: 'Are you absolutely sure you want to permanently delete this video? All transcoded HLS segments and database records will be destroyed.',
    confirmText: 'Yes, delete it',
    cancelText: 'Cancel',
    type: 'danger',
    onConfirm: async () => {
      try {
        await deleteVideoRequest(id)
        videos.value = videos.value.filter(v => v.id !== id)
        if (activeVideoId.value === id) activeVideoId.value = null
        emit('notify', { message: "Video permanently deleted.", type: "success" })
      } catch (err) {
        emit('notify', { message: "Failed to delete video.", type: "error" })
      }
    }
  })
}

const fileInputRef = ref(null)

const triggerUpload = () => {
  if (fileInputRef.value) {
    fileInputRef.value.click()
  }
}

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files.length) return
  
  for (const file of files) {
    try {
      emit('notify', { message: `Starting upload of ${file.name}...`, type: "info" })
      await uploadStore.uploadFile(file)
      emit('notify', { message: `Upload complete for ${file.name}! Transcoding queued.`, type: "success" })
      fetchVideos() // Refresh list
    } catch (err) {
      emit('notify', { message: `Upload failed for ${file.name}.`, type: "error" })
    }
  }
  
  if (event.target) event.target.value = ''
}

const themeColors = [
  { name: 'Rose', hex: '#f43f5e', class: 'bg-[#f43f5e]' },
  { name: 'Blue', hex: '#3b82f6', class: 'bg-[#3b82f6]' },
  { name: 'Indigo', hex: '#6366f1', class: 'bg-[#6366f1]' },
  { name: 'Emerald', hex: '#10b981', class: 'bg-[#10b981]' },
  { name: 'Amber', hex: '#f59e0b', class: 'bg-[#f59e0b]' }
]

const changePrimaryColor = (colorHex) => {
  document.documentElement.style.setProperty('--color-primary', colorHex)
  localStorage.setItem('SV_PRIMARY_COLOR', colorHex)
}

let videosInterval = null
let healthInterval = null

onMounted(() => {
  const savedColor = localStorage.getItem('SV_PRIMARY_COLOR')
  if (savedColor) {
    document.documentElement.style.setProperty('--color-primary', savedColor)
  }

  fetchVideos()
  fetchHealth()
  videosInterval = setInterval(fetchVideos, 5000)
  healthInterval = setInterval(fetchHealth, 30000)
})

onUnmounted(() => {
  if (videosInterval) clearInterval(videosInterval)
  if (healthInterval) clearInterval(healthInterval)
})
</script>

<template>
  <div class="flex min-h-screen bg-slate-50 dark:bg-[#111318] text-slate-900 dark:text-slate-100 font-sans transition-colors duration-300">
    
    <!-- Hidden File Input for Sidebar Button -->
    <input ref="fileInputRef" type="file" class="hidden" accept="video/*" multiple @change="handleFileUpload" />

    <!-- Sidebar -->
    <Sidebar :active-tab="activeTab" @navigate="handleNavigate" @upload="triggerUpload" />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-y-auto">
      
      <!-- Top Navigation Bar -->
      <header class="h-20 border-b border-slate-200 dark:border-[#2d3139] px-8 flex items-center justify-between sticky top-0 bg-slate-50/90 dark:bg-[#111318]/90 backdrop-blur-md z-30 transition-colors duration-300">
        
        <div class="text-sm font-semibold text-slate-800 dark:text-slate-200">
          Admin Dashboard / <span class="text-primary">{{ activeTab.charAt(0).toUpperCase() + activeTab.slice(1) }}</span>
        </div>

        <!-- Right Actions -->
        <div class="flex items-center gap-4">
          <ThemeToggle />
          <button @click="showSettings = true" class="p-2 text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-[#2d3139]/50 rounded-lg transition-colors cursor-pointer" title="System Settings">
            <Settings class="w-5 h-5" />
          </button>
        </div>
      </header>

      <!-- Main Content -->
      <main class="flex-1 px-8 py-8">
        
        <div v-if="activeTab === 'library'" class="max-w-7xl mx-auto space-y-8">
          
          <!-- Library Header -->
          <div>
            <h2 class="text-2xl font-bold text-slate-900 dark:text-white mb-1">Content Library</h2>
            <p class="text-slate-500 dark:text-slate-400 text-sm">Manage and monitor your VOD assets.</p>
          </div>

          <!-- Stats Grid -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="bg-white dark:bg-[#1a1d24] p-5 rounded-xl border border-slate-200 dark:border-[#2d3139] shadow-sm flex flex-col justify-between">
              <div class="flex justify-between items-start mb-6">
                <div class="w-10 h-10 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
                  <Activity class="w-5 h-5" />
                </div>
                <span class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-900/30 px-2 py-1 rounded">↗ +12%</span>
              </div>
              <div>
                <h3 class="text-slate-500 dark:text-slate-400 text-sm font-medium mb-1">Total Views</h3>
                <div class="text-2xl font-bold text-slate-900 dark:text-white">842,593</div>
              </div>
            </div>

            <div class="bg-white dark:bg-[#1a1d24] p-5 rounded-xl border border-slate-200 dark:border-[#2d3139] shadow-sm flex flex-col justify-between">
              <div class="flex justify-between items-start mb-6">
                <div class="w-10 h-10 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 flex items-center justify-center">
                  <Library class="w-5 h-5" />
                </div>
                <span class="text-xs font-semibold text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30 px-2 py-1 rounded">24 New</span>
              </div>
              <div>
                <h3 class="text-slate-500 dark:text-slate-400 text-sm font-medium mb-1">Total Assets</h3>
                <div class="text-2xl font-bold text-slate-900 dark:text-white">{{ videos.length }} Videos</div>
              </div>
            </div>

            <div class="bg-white dark:bg-[#1a1d24] p-5 rounded-xl border border-slate-200 dark:border-[#2d3139] shadow-sm flex flex-col justify-between">
              <div class="flex justify-between items-start mb-6">
                <div class="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                  <HardDrive class="w-5 h-5" />
                </div>
                <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-[#2d3139] px-2 py-1 rounded">Local Server</span>
              </div>
              <div>
                <div class="flex justify-between items-baseline mb-2">
                  <h3 class="text-slate-500 dark:text-slate-400 text-sm font-medium">Storage Available</h3>
                  <span class="text-sm font-semibold">{{ (health.storage_available_bytes / (1024**3)).toFixed(1) }} GB Free</span>
                </div>
                <div class="w-full bg-slate-100 dark:bg-[#2d3139] h-2 rounded-full overflow-hidden">
                  <div class="bg-primary h-full rounded-full" style="width: 35%"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Active Uploads (Banner Style) -->
          <div v-if="Object.keys(uploadStore.uploads).length" class="space-y-3">
            <div 
              v-for="(upload, id) in uploadStore.uploads" 
              :key="id" 
              class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-xl p-4 shadow-sm flex items-center gap-4"
            >
              <div class="w-12 h-12 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                <Activity class="w-6 h-6 animate-pulse" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex justify-between items-baseline mb-1.5">
                  <span class="font-medium text-slate-900 dark:text-white truncate">{{ upload.file.name }}</span>
                  <span class="text-sm font-semibold text-primary">{{ upload.progress }}%</span>
                </div>
                <div class="w-full bg-slate-100 dark:bg-[#2d3139] h-1.5 rounded-full overflow-hidden">
                  <div class="bg-primary h-full transition-all duration-300" :style="{ width: `${upload.progress}%` }"></div>
                </div>
                <div class="mt-1.5 text-xs text-slate-500 dark:text-slate-400 flex items-center gap-2">
                  <span class="w-2 h-2 rounded-full bg-emerald-500" v-if="upload.status === 'completed'"></span>
                  <span class="w-2 h-2 rounded-full bg-primary animate-pulse" v-else-if="upload.status === 'uploading'"></span>
                  <span class="w-2 h-2 rounded-full bg-rose-500" v-else-if="upload.status === 'error'"></span>
                  {{ upload.status.charAt(0).toUpperCase() + upload.status.slice(1) }}
                </div>
              </div>
            </div>
          </div>

          <!-- Filters & Actions -->
          <div class="mt-8 mb-6 border-b border-slate-200 dark:border-[#2d3139] flex justify-between items-end pb-px">
            <div class="flex gap-6">
              <button 
                @click="activeFilter = 'all'; currentPage = 1"
                :class="[
                  'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
                  activeFilter === 'all' 
                    ? 'text-primary' 
                    : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                ]"
              >
                All Content
                <div v-if="activeFilter === 'all'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
              </button>
              <button 
                @click="activeFilter = 'published'; currentPage = 1"
                :class="[
                  'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
                  activeFilter === 'published' 
                    ? 'text-primary' 
                    : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                ]"
              >
                Published
                <div v-if="activeFilter === 'published'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
              </button>
              <button 
                @click="activeFilter = 'processing'; currentPage = 1"
                :class="[
                  'pb-3 text-sm font-semibold transition-all cursor-pointer relative',
                  activeFilter === 'processing' 
                    ? 'text-primary' 
                    : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
                ]"
              >
                Processing
                <div v-if="activeFilter === 'processing'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary rounded-full"></div>
              </button>
            </div>
          </div>

          <!-- Video Grid -->
          <div v-if="isLoading" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            <div v-for="i in 8" :key="i" class="bg-white dark:bg-[#1a1d24] rounded-xl aspect-video animate-pulse border border-slate-100 dark:border-[#2d3139]"></div>
          </div>

          <div v-else-if="filteredVideos.length === 0" class="text-center py-20 bg-white dark:bg-[#1a1d24] rounded-2xl border border-slate-200 dark:border-[#2d3139] border-dashed">
            <div class="w-16 h-16 bg-slate-50 dark:bg-[#2d3139] rounded-full flex items-center justify-center mx-auto mb-4">
              <Library class="w-8 h-8 text-slate-400" />
            </div>
            <h3 class="text-slate-900 dark:text-white font-semibold text-lg mb-1">No assets found</h3>
            <p class="text-slate-500 dark:text-slate-400 text-sm max-w-sm mx-auto mb-6">
              {{ activeFilter === 'all' ? 'Your content library is empty. Upload your first video to see it appear here.' : 'No videos match the selected filter.' }}
            </p>
            <button v-if="activeFilter === 'all'" @click="triggerUpload" class="bg-primary hover:bg-rose-600 text-white px-6 py-2.5 rounded-lg font-medium transition-colors cursor-pointer">
              Upload Video
            </button>
          </div>

          <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            <VideoCard 
              v-for="v in paginatedVideos" 
              :key="v.id" 
              :video="v" 
              @play="activeVideoId = $event"
              @delete="requestDelete"
            />
          </div>

          <!-- Pagination & Showing customization -->
          <div v-if="filteredVideos.length" class="mt-8 flex flex-col sm:flex-row items-center justify-between gap-4 pt-6 border-t border-slate-200 dark:border-[#2d3139]">
            <div class="flex items-center gap-3 text-sm text-slate-500 dark:text-slate-400">
              <span>Showing {{ (currentPage - 1) * itemsPerPage + 1 }} to {{ Math.min(currentPage * itemsPerPage, filteredVideos.length) }} of {{ filteredVideos.length }}</span>
              <select v-model="itemsPerPage" @change="currentPage = 1" class="px-2 py-1 bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-md text-xs cursor-pointer text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-1 focus:ring-primary">
                <option :value="4">4 per page</option>
                <option :value="8">8 per page</option>
                <option :value="12">12 per page</option>
                <option :value="24">24 per page</option>
              </select>
            </div>

            <div class="flex items-center gap-1">
              <button 
                @click="changePage(currentPage - 1)" 
                :disabled="currentPage === 1"
                class="w-9 h-9 flex items-center justify-center rounded-lg border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-[#2d3139]/50 disabled:text-slate-300 dark:disabled:text-slate-700 disabled:border-slate-100 dark:disabled:border-[#2d3139]/30 disabled:bg-slate-50/50 dark:disabled:bg-transparent disabled:cursor-not-allowed transition-colors cursor-pointer"
              >
                <ChevronLeft class="w-4 h-4" />
              </button>
              
              <button 
                v-for="page in totalPages" 
                :key="page"
                @click="changePage(page)"
                :class="['w-9 h-9 flex items-center justify-center rounded-lg font-medium transition-colors cursor-pointer text-sm',
                  currentPage === page 
                    ? 'bg-primary text-white' 
                    : 'border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-[#2d3139]'
                ]"
              >
                {{ page }}
              </button>

              <button 
                @click="changePage(currentPage + 1)" 
                :disabled="currentPage === totalPages"
                class="w-9 h-9 flex items-center justify-center rounded-lg border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-[#2d3139]/50 disabled:text-slate-300 dark:disabled:text-slate-700 disabled:border-slate-100 dark:disabled:border-[#2d3139]/30 disabled:bg-slate-50/50 dark:disabled:bg-transparent disabled:cursor-not-allowed transition-colors cursor-pointer"
              >
                <ChevronRight class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <UnderConstruction v-else :title="activeTab.charAt(0).toUpperCase() + activeTab.slice(1)" />

      </main>
    </div>

    <!-- Player Overlay Modal -->
    <div v-if="activeVideoId" class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6 lg:p-8 bg-slate-900/15 backdrop-blur-[2px]">
      <div class="w-full max-w-4xl bg-black rounded-2xl overflow-hidden shadow-2xl ring-1 ring-white/10">
        <div class="bg-[#1a1d24] p-4 flex justify-between items-center text-white border-b border-white/10">
          <span class="font-medium truncate">{{ videos.find(v => v.id === activeVideoId)?.title }}</span>
          <button @click="activeVideoId = null" class="p-1 hover:bg-white/10 rounded-lg transition-colors cursor-pointer">
            <X class="w-6 h-6" />
          </button>
        </div>
        <ForgePlayer :video-id="activeVideoId" />
      </div>
    </div>

    <!-- Credentials Settings Modal -->
    <div v-if="showSettings" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/15 backdrop-blur-[2px]">
      <div class="bg-white dark:bg-[#1a1d24] rounded-2xl p-6 w-full max-w-md border border-slate-200 dark:border-[#2d3139] shadow-2xl">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
            <Settings class="w-5 h-5 text-primary animate-spin-slow" />
            System & Credentials
          </h3>
          <button @click="showSettings = false" class="p-1 hover:bg-slate-100 dark:hover:bg-[#2d3139] rounded-lg text-slate-400 transition-colors cursor-pointer">
            <X class="w-5 h-5" />
          </button>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-6">Configure the primary dashboard theme and the authentication keys used to query the secure API endpoints.</p>
        <div class="space-y-4">
          <!-- Theme Color Customization -->
          <div>
            <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-2">Accent Color Theme</label>
            <div class="flex gap-3">
              <button 
                v-for="c in themeColors" 
                :key="c.name" 
                :title="c.name"
                @click="changePrimaryColor(c.hex)"
                :style="{ backgroundColor: c.hex }"
                class="w-7 h-7 rounded-full shadow-sm hover:scale-110 transition-transform cursor-pointer border-2 border-white dark:border-[#1a1d24] focus:outline-none ring-2 ring-transparent focus:ring-slate-400"
                :aria-label="c.name"
              ></button>
            </div>
          </div>
          <hr class="border-slate-100 dark:border-[#2d3139] my-4" />
          <div>
            <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Admin API Key</label>
            <input v-model="apiKey" type="password" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
          </div>
          <div>
            <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Playback Scope Key</label>
            <input v-model="playbackKey" type="password" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
          </div>
        </div>
        <div class="flex gap-3 justify-end mt-8">
          <button @click="showSettings = false" class="px-5 py-2.5 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors">Cancel</button>
          <button @click="saveSettings" class="px-5 py-2.5 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors">Save Keys</button>
        </div>
      </div>
    </div>

    <!-- Reusable Confirmation Modal -->
    <ConfirmationModal 
      :is-open="confirmModalConfig.isOpen"
      :title="confirmModalConfig.title"
      :description="confirmModalConfig.description"
      :confirm-text="confirmModalConfig.confirmText"
      :cancel-text="confirmModalConfig.cancelText"
      :type="confirmModalConfig.type"
      @confirm="handleConfirmAction"
      @cancel="handleCancelAction"
    />

  </div>
</template>
