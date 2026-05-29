<script setup>
import { onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useDebouncedRef } from '../composables/useDebouncedRef'
import { useUploadStore } from '../stores/upload'
import {
  deleteVideo as deleteVideoRequest,
  getHealth,
  listVideos,
  listLibraries,
  listLibraryKeys,
  updateVideo
} from '../api/videos'
import { getAdminKey, getPlaybackKey, saveCredentials, clearCredentials } from '../lib/credentials'
import VideoCard from '../components/VideoCard.vue'
import Documentation from '../components/Documentation.vue'
import UnderConstruction from '../components/UnderConstruction.vue'
import PlayerConfig from './PlayerConfig.vue'
import ConfirmationModal from '../components/ConfirmationModal.vue'
import AppLayout from '../components/layout/AppLayout.vue'
import BaseModal from '../components/modals/BaseModal.vue'
import SettingsModal from '../components/modals/SettingsModal.vue'
import PlayerModal from '../components/modals/PlayerModal.vue'
import StatsCard from '../components/video/StatsCard.vue'
import VideoFilters from '../components/video/VideoFilters.vue'
import VideoPagination from '../components/video/VideoPagination.vue'
import KeyManagement from '../components/library/KeyManagement.vue'
import UploadModal from '../components/library/UploadModal.vue'
import {
  Activity,
  Library,
  HardDrive,
  Pencil,
  Copy,
  Check,
  Code
} from 'lucide-vue-next'

const emit = defineEmits(['logout', 'notify'])

const uploadStore = useUploadStore()
const videos = ref([])

const libraries = ref([])
const activeLibraryId = ref('')
const libraryKeys = ref([])
const unmaskedKeys = ref({})

const toggleKeyMask = (keyId) => {
  unmaskedKeys.value[keyId] = !unmaskedKeys.value[keyId]
}

const showUploadModal = ref(false)
const showEmbedModal = ref(false)
const embedCode = ref('')
const embedVideoId = ref(null)
const embedKeyParam = ref('')
const copySuccess = ref(false)
const activeVideoId = ref(null)
const showEditVideoModal = ref(false)
const editingVideo = ref(null)
const editingVideoTitle = ref('')
const editingVideoLibraryId = ref('')
const isLoading = ref(true)
const health = ref({ status: 'ok', storage_available_bytes: 0 })

const activeFilter = ref('all')
const currentPage = ref(1)
const itemsPerPage = ref(12)
const { raw: searchQuery, debounced: debouncedSearchQuery } = useDebouncedRef('')

watch(debouncedSearchQuery, () => {
  currentPage.value = 1
})

const filteredVideos = computed(() => {
  let vids = videos.value

  if (debouncedSearchQuery.value) {
    const q = debouncedSearchQuery.value.toLowerCase()
    vids = vids.filter(v => v.title.toLowerCase().includes(q) || v.id.toLowerCase().includes(q))
  }

  if (activeFilter.value === 'published') {
    return vids.filter(v => v.status === 'completed')
  }
  if (activeFilter.value === 'processing') {
    return vids.filter(v => v.status === 'transcoding' || v.status === 'pending')
  }
  return vids
})

const paginatedVideos = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredVideos.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(filteredVideos.value.length / itemsPerPage.value) || 1
})

const activeTab = ref('library')
const showSettings = ref(false)

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
      title: 'Sign Out',
      description: 'Are you sure you want to sign out of the Selvod console? You will need to re-enter your API keys to regain access.',
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

const saveSettings = ({ apiKey, playbackKey }) => {
  saveCredentials({ adminKey: apiKey, playbackKey, rememberMe: true })
  showSettings.value = false
  emit('notify', { message: "Credentials updated successfully.", type: "success" })
  fetchVideos()
  fetchHealth()
}

const fetchLibraries = async () => {
  if (librariesController) librariesController.abort()
  librariesController = new AbortController()
  try {
    const data = await listLibraries({ signal: librariesController.signal })
    libraries.value = data || []
    if (libraries.value.length > 0 && !activeLibraryId.value) {
      activeLibraryId.value = libraries.value[0].id
    }
  } catch (err) {
    if (err.name === 'CanceledError') return
    console.error("Failed to fetch libraries", err)
  }
}

const fetchLibraryKeys = async () => {
  if (!activeLibraryId.value) return
  if (librariesController) librariesController.abort()
  librariesController = new AbortController()
  try {
    const data = await listLibraryKeys(activeLibraryId.value, { signal: librariesController.signal })
    libraryKeys.value = data || []
  } catch (err) {
    if (err.name === 'CanceledError') return
    console.error("Failed to fetch library keys", err)
  }
}

watch(activeLibraryId, () => {
  fetchLibraryKeys()
  fetchVideos()
})

const handleEmbedRequest = async (videoId) => {
  try {
    embedVideoId.value = videoId
    const activeKey = libraryKeys.value.find(k => k.is_active)
    embedKeyParam.value = activeKey ? activeKey.playback_secret : 'YOUR_LIBRARY_KEY'
    updateEmbedCode()
    showEmbedModal.value = true
  } catch (err) {
    emit('notify', { message: "Failed to generate embed code.", type: "error" })
  }
}

const updateEmbedCode = () => {
  if (!embedVideoId.value) return
  const baseUrl = window.location.origin
  embedCode.value = `<iframe src="${baseUrl}/embed.html?videoId=${embedVideoId.value}&key=${embedKeyParam.value}" width="640" height="360" frameborder="0" allow="autoplay; encrypted-media; picture-in-picture" allowfullscreen></iframe>`
}

const copyToClipboard = async (text, successMsg) => {
  try {
    await navigator.clipboard.writeText(text)
    emit('notify', { message: successMsg || 'Copied to clipboard.', type: 'success' })
  } catch {
    emit('notify', { message: 'Failed to copy to clipboard.', type: 'error' })
  }
}

const copyEmbedCode = () => {
  copyToClipboard(embedCode.value, 'Embed code copied to clipboard.')
  copySuccess.value = true
  setTimeout(() => {
    copySuccess.value = false
  }, 2000)
}

const fetchVideos = async () => {
  if (videosController) videosController.abort()
  videosController = new AbortController()
  try {
    const data = await listVideos(activeLibraryId.value, { signal: videosController.signal })
    videos.value = data.videos || []
  } catch (err) {
    if (err.name === 'CanceledError') return
    console.error("Failed to fetch videos", err)
    emit('notify', { message: "Failed to fetch video library.", type: "error" })
  } finally {
    isLoading.value = false
  }
}

const fetchHealth = async () => {
  if (healthController) healthController.abort()
  healthController = new AbortController()
  try {
    health.value = await getHealth({ signal: healthController.signal })
  } catch (err) {
    if (err.name === 'CanceledError') return
    console.error("Health check failed", err)
  }
}

const requestDelete = (id) => {
  triggerConfirmation({
    title: 'Delete Video',
    description: 'Are you sure you want to permanently delete this video? All transcoded HLS segments and database records will be destroyed.',
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

const triggerEditVideo = (video) => {
  editingVideo.value = video
  editingVideoTitle.value = video.title
  editingVideoLibraryId.value = video.library_id
  showEditVideoModal.value = true
}

const handleUpdateVideo = async () => {
  if (!editingVideoTitle.value.trim() || !editingVideoLibraryId.value) return
  try {
    await updateVideo(editingVideo.value.id, {
      title: editingVideoTitle.value.trim(),
      library_id: editingVideoLibraryId.value
    })
    emit('notify', { message: "Video details updated.", type: "success" })
    showEditVideoModal.value = false
    await fetchVideos()
  } catch (err) {
    emit('notify', { message: "Failed to update video.", type: "error" })
  }
}

const triggerUpload = () => {
  if (activeTab.value !== 'library') activeTab.value = 'library'
  showUploadModal.value = true
}

let videosInterval = null
let healthInterval = null
let videosController = null
let healthController = null
let librariesController = null

const POLL_FAST = 5000   // 5s when videos are processing
const POLL_SLOW = 30000  // 30s when all videos are idle

const getPollingInterval = () => {
  const hasProcessing = videos.value.some(v => v.status === 'transcoding' || v.status === 'pending')
  return hasProcessing ? POLL_FAST : POLL_SLOW
}

const schedulePoll = () => {
  if (videosInterval) clearTimeout(videosInterval)
  videosInterval = setTimeout(async () => {
    await fetchVideos()
    schedulePoll()
  }, getPollingInterval())
}

onMounted(async () => {
  const savedColor = localStorage.getItem('SV_PRIMARY_COLOR')
  if (savedColor) {
    document.documentElement.style.setProperty('--color-primary', savedColor)
  }

  await fetchLibraries()
  fetchHealth()
  schedulePoll()
  healthInterval = setInterval(fetchHealth, 30000)
})

onUnmounted(() => {
  if (videosInterval) clearTimeout(videosInterval)
  if (healthInterval) clearInterval(healthInterval)
  if (videosController) videosController.abort()
  if (healthController) healthController.abort()
  if (librariesController) librariesController.abort()
})
</script>

<template>
  <AppLayout
    :active-tab="activeTab"
    @navigate="handleNavigate"
    @open-settings="showSettings = true"
  >
    <div
      v-if="activeTab === 'library'"
      class="h-full overflow-y-auto"
    >
    <div class="max-w-7xl mx-auto space-y-8 pb-8">
      <KeyManagement
        :libraries="libraries"
        :active-library-id="activeLibraryId"
        :library-keys="libraryKeys"
        :unmasked-keys="unmaskedKeys"
        @update:active-library-id="activeLibraryId = $event"
        @toggle-key-mask="toggleKeyMask"
        @fetch-keys="fetchLibraryKeys"
        @fetch-libraries="fetchLibraries"
        @notify="(n) => emit('notify', n)"
      />

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatsCard
          :icon="Activity"
          label="Total Views"
          value="Coming Soon"
          badge="Analytics TBD"
          badge-color="text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-[#2d3139]"
          icon-bg="bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400"
        />
        <StatsCard
          :icon="Library"
          label="Total Assets"
          :value="`${videos.length} Files`"
          badge="Videos & Audio"
          icon-bg="bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400"
          badge-color="text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/30"
        />
        <StatsCard
          :icon="HardDrive"
          label="Storage Available"
          :value="`${(health.storage_available_bytes / (1024**3)).toFixed(1)} GB Free`"
          badge="Local Server"
          badge-color="text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-[#2d3139]"
          icon-bg="bg-primary/10 text-primary"
          :progress="35"
          progress-label="35% Used"
        />
      </div>


      <VideoFilters
        :active-filter="activeFilter"
        :search-query="searchQuery"
        @update:active-filter="activeFilter = $event; currentPage = 1"
        @update:search-query="searchQuery = $event"
        @upload="triggerUpload"
      />

      <div
        v-if="isLoading"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
      >
        <div
          v-for="i in 8"
          :key="i"
          class="bg-white dark:bg-[#1a1d24] rounded-xl aspect-video animate-pulse border border-slate-100 dark:border-[#2d3139]"
        />
      </div>

      <div
        v-else-if="filteredVideos.length === 0"
        class="text-center py-20 bg-white dark:bg-[#1a1d24] rounded-2xl border border-slate-200 dark:border-[#2d3139] border-dashed"
      >
        <div class="w-16 h-16 bg-slate-50 dark:bg-[#2d3139] rounded-full flex items-center justify-center mx-auto mb-4">
          <Library class="w-8 h-8 text-slate-400" />
        </div>
        <h3 class="text-slate-900 dark:text-white font-semibold text-lg mb-1">
          No assets found
        </h3>
        <p class="text-slate-500 dark:text-slate-400 text-sm max-w-sm mx-auto mb-6">
          {{ activeFilter === 'all' ? 'Your content library is empty. Upload your first video to see it appear here.' : 'No videos match the selected filter.' }}
        </p>
        <button
          v-if="activeFilter === 'all'"
          class="bg-primary hover:bg-rose-600 text-white px-6 py-2.5 rounded-lg font-medium transition-colors cursor-pointer"
          @click="triggerUpload"
        >
          Upload Video
        </button>
      </div>

      <div
        v-else
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
      >
        <VideoCard
          v-for="v in paginatedVideos"
          :key="v.id"
          :video="v"
          @play="activeVideoId = $event"
          @delete="requestDelete"
          @embed="handleEmbedRequest"
          @edit="triggerEditVideo"
        />
      </div>

      <VideoPagination
        v-if="filteredVideos.length"
        :current-page="currentPage"
        :total-pages="totalPages"
        :total-items="filteredVideos.length"
        :items-per-page="itemsPerPage"
        @update:current-page="currentPage = $event"
        @update:items-per-page="itemsPerPage = $event; currentPage = 1"
      />
    </div>
    </div>

    <div
      v-else-if="activeTab === 'docs'"
      class="h-full w-full overflow-hidden"
    >
      <Documentation
        :api-key="getAdminKey()"
        :playback-key="getPlaybackKey()"
        :library-id="activeLibraryId"
      />
    </div>

    <PlayerConfig
      v-else-if="activeTab === 'player-ui'"
      @notify="(n) => emit('notify', n)"
    />

    <UnderConstruction
      v-else
      :title="activeTab.charAt(0).toUpperCase() + activeTab.slice(1)"
    />

    <PlayerModal
      :is-open="!!activeVideoId"
      :video-id="activeVideoId"
      :videos="videos"
      @close="activeVideoId = null"
    />

    <SettingsModal
      :is-open="showSettings"
      :api-key="getAdminKey()"
      :playback-key="getPlaybackKey()"
      @close="showSettings = false"
      @save="saveSettings"
    />



    <BaseModal
      :is-open="showEmbedModal"
      title="Embed Video Player"
      description="Use the iframe below to embed the ForgePlayer on external sites."
      max-width="max-w-lg"
      @close="showEmbedModal = false"
    >
      <template #icon>
        <Code class="w-5 h-5 text-primary" />
      </template>

      <div class="relative bg-slate-950 p-4 rounded-xl font-mono text-xs text-emerald-400 border border-slate-800 break-all pr-12">
        {{ embedCode }}
        <button
          class="absolute right-3 top-3 p-1.5 bg-slate-900 hover:bg-slate-800 text-slate-300 rounded transition-colors cursor-pointer border border-slate-800"
          aria-label="Copy Embed Code"
          @click="copyEmbedCode"
        >
          <component
            :is="copySuccess ? Check : Copy"
            class="w-4 h-4"
          />
        </button>
      </div>
      <div class="flex justify-end mt-6">
        <button
          class="px-5 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
          @click="showEmbedModal = false"
        >
          Done
        </button>
      </div>
    </BaseModal>

    <BaseModal
      :is-open="showEditVideoModal"
      title="Edit Video Details"
      description="Update the title or move the video to another library."
      max-width="max-w-md"
      @close="showEditVideoModal = false"
    >
      <template #icon>
        <Pencil class="w-5 h-5 text-primary" />
      </template>
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Video Title</label>
          <input
            v-model="editingVideoTitle"
            placeholder="e.g. Tutorial Video"
            type="text"
            class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
          >
        </div>
        <div>
          <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Library</label>
          <select
            v-model="editingVideoLibraryId"
            class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors cursor-pointer"
          >
            <option
              v-for="lib in libraries"
              :key="lib.id"
              :value="lib.id"
            >
              {{ lib.name }}
            </option>
          </select>
        </div>
      </div>
      <div class="flex gap-3 justify-end mt-8">
        <button
          class="px-5 py-2.5 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer"
          @click="showEditVideoModal = false"
        >
          Cancel
        </button>
        <button
          class="px-5 py-2.5 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
          @click="handleUpdateVideo"
        >
          Save Changes
        </button>
      </div>
    </BaseModal>

    <UploadModal
      :is-open="showUploadModal"
      :active-library-id="activeLibraryId"
      :libraries="libraries"
      @close="showUploadModal = false"
      @notify="(n) => emit('notify', n)"
      @upload-complete="fetchVideos"
    />

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
  </AppLayout>
</template>
