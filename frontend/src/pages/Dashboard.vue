<script setup>
import { onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useUploadStore } from '../stores/upload'
import {
  deleteVideo as deleteVideoRequest,
  getHealth,
  listVideos,
  listLibraries,
  createLibrary,
  updateLibrary,
  listLibraryKeys,
  createLibraryKey,
  revokeLibraryKey,
  deleteLibraryKey,
  regenerateLibraryKey,
  getEmbed,
  updateVideo
} from '../api/videos'
import { getAdminKey, getPlaybackKey, saveCredentials, clearCredentials } from '../lib/credentials'
import VideoCard from '../components/VideoCard.vue'
import Documentation from '../components/Documentation.vue'
import UnderConstruction from '../components/UnderConstruction.vue'
import ConfirmationModal from '../components/ConfirmationModal.vue'
import AppLayout from '../components/layout/AppLayout.vue'
import BaseModal from '../components/modals/BaseModal.vue'
import SettingsModal from '../components/modals/SettingsModal.vue'
import PlayerModal from '../components/modals/PlayerModal.vue'
import StatsCard from '../components/video/StatsCard.vue'
import VideoFilters from '../components/video/VideoFilters.vue'
import VideoPagination from '../components/video/VideoPagination.vue'
import {
  Activity,
  Library,
  HardDrive,
  Key,
  Copy,
  Plus,
  Trash2,
  Lock,
  Check,
  Code,
  Pencil,
  Eye,
  EyeOff,
  RefreshCw,
  Ban,
  Upload
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

const showCreateLibraryModal = ref(false)
const newLibraryName = ref('')
const showEditLibraryModal = ref(false)
const editingLibraryName = ref('')
const showCreateKeyModal = ref(false)
const newKeyName = ref('')
const generatedKeySecret = ref('')
const showUploadModal = ref(false)
const showEmbedModal = ref(false)
const embedCode = ref('')
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
  try {
    const data = await listLibraries()
    libraries.value = data || []
    if (libraries.value.length > 0 && !activeLibraryId.value) {
      activeLibraryId.value = libraries.value[0].id
    }
  } catch (err) {
    console.error("Failed to fetch libraries", err)
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Intl.DateTimeFormat('en-US', {
    month: 'short', day: 'numeric', year: 'numeric'
  }).format(new Date(dateStr))
}

const fetchLibraryKeys = async () => {
  if (!activeLibraryId.value) return
  try {
    const data = await listLibraryKeys(activeLibraryId.value)
    libraryKeys.value = data || []
  } catch (err) {
    console.error("Failed to fetch library keys", err)
  }
}

watch(activeLibraryId, () => {
  fetchLibraryKeys()
  fetchVideos()
})

const handleCreateLibrary = async () => {
  if (!newLibraryName.value.trim()) return
  try {
    const result = await createLibrary(newLibraryName.value.trim())
    // Backend returns { library, default_key } — the secret is only shown once
    const lib = result.library || result
    emit('notify', { message: "Library created. Default key generated.", type: "success" })
    newLibraryName.value = ''
    showCreateLibraryModal.value = false
    await fetchLibraries()
    activeLibraryId.value = lib.id
    // Show the auto-generated default key secret immediately
    if (result.default_key?.playback_secret) {
      generatedKeySecret.value = result.default_key.playback_secret
    }
  } catch (err) {
    emit('notify', { message: "Failed to create library.", type: "error" })
  }
}

const triggerEditLibrary = () => {
  const lib = libraries.value.find(l => l.id === activeLibraryId.value)
  if (!lib) return
  editingLibraryName.value = lib.name
  showEditLibraryModal.value = true
}

const handleUpdateLibrary = async () => {
  if (!editingLibraryName.value.trim()) return
  try {
    await updateLibrary(activeLibraryId.value, editingLibraryName.value.trim())
    emit('notify', { message: "Library name updated successfully.", type: "success" })
    showEditLibraryModal.value = false
    await fetchLibraries()
  } catch (err) {
    emit('notify', { message: "Failed to update library name.", type: "error" })
  }
}

const handleCreateKey = async () => {
  if (!newKeyName.value.trim() || !activeLibraryId.value) return
  try {
    const result = await createLibraryKey(activeLibraryId.value, newKeyName.value.trim())
    generatedKeySecret.value = result.playback_secret
    newKeyName.value = ''
    showCreateKeyModal.value = false
    await fetchLibraryKeys()
  } catch (err) {
    emit('notify', { message: "Failed to create key.", type: "error" })
  }
}

const handleRevokeKey = async (keyId) => {
  triggerConfirmation({
    title: 'Revoke Playback Key',
    description: 'Are you sure you want to revoke this key? Any external client using this key will immediately lose access to the library.',
    confirmText: 'Revoke Key',
    cancelText: 'Cancel',
    type: 'danger',
    onConfirm: async () => {
      try {
        await revokeLibraryKey(activeLibraryId.value, keyId)
        emit('notify', { message: "Key revoked successfully.", type: "success" })
        await fetchLibraryKeys()
      } catch (err) {
        emit('notify', { message: "Failed to revoke key.", type: "error" })
      }
    }
  })
}

const handleDeleteKey = async (keyId) => {
  triggerConfirmation({
    title: 'Delete Playback Key',
    description: 'Are you sure you want to permanently delete this key? This action is irreversible and any external client using this key will lose access.',
    confirmText: 'Delete Key',
    cancelText: 'Cancel',
    type: 'danger',
    onConfirm: async () => {
      try {
        await deleteLibraryKey(activeLibraryId.value, keyId)
        emit('notify', { message: "Key permanently deleted.", type: "success" })
        await fetchLibraryKeys()
      } catch (err) {
        emit('notify', { message: "Failed to delete key.", type: "error" })
      }
    }
  })
}

const handleRegenerateKey = async (keyId) => {
  triggerConfirmation({
    title: 'Regenerate Playback Key',
    description: 'Are you sure you want to regenerate this key? The existing key secret will be invalidated immediately, and a new one will be generated.',
    confirmText: 'Regenerate',
    cancelText: 'Cancel',
    type: 'warning',
    onConfirm: async () => {
      try {
        const result = await regenerateLibraryKey(activeLibraryId.value, keyId)
        generatedKeySecret.value = result.playback_secret
        emit('notify', { message: "Key regenerated successfully.", type: "success" })
        await fetchLibraryKeys()
      } catch (err) {
        emit('notify', { message: "Failed to regenerate key.", type: "error" })
      }
    }
  })
}

const handleEmbedRequest = async (videoId) => {
  try {
    const data = await getEmbed(videoId)
    embedCode.value = `<iframe src="${data.url}" width="640" height="360" frameborder="0" allow="autoplay; encrypted-media; picture-in-picture" allowfullscreen></iframe>`
    showEmbedModal.value = true
  } catch (err) {
    emit('notify', { message: "Failed to generate embed URL.", type: "error" })
  }
}

const copyEmbedCode = () => {
  navigator.clipboard.writeText(embedCode.value)
  copySuccess.value = true
  setTimeout(() => {
    copySuccess.value = false
  }, 2000)
}

const fetchVideos = async () => {
  try {
    const data = await listVideos(activeLibraryId.value)
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

const fileInputRef = ref(null)
const isDragging = ref(false)

const triggerUpload = () => {
  // Navigate to library tab first so the user sees what's happening
  if (activeTab.value !== 'library') activeTab.value = 'library'
  showUploadModal.value = true
}

const handleDrop = async (event) => {
  isDragging.value = false
  const files = event.dataTransfer?.files
  if (!files?.length) return
  for (const file of files) {
    if (!file.type.startsWith('video/') && !file.type.startsWith('audio/')) {
      emit('notify', { message: `${file.name} is not a video or audio file.`, type: 'error' })
      continue
    }
    try {
      emit('notify', { message: `Uploading ${file.name}...`, type: 'info' })
      await uploadStore.uploadFile(file, null, activeLibraryId.value)
      emit('notify', { message: `${file.name} uploaded. Transcoding queued.`, type: 'success' })
      fetchVideos()
    } catch (err) {
      emit('notify', { message: `${file.name} upload failed.`, type: 'error' })
    }
  }
}

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files.length) return

  showUploadModal.value = false

  for (const file of files) {
    try {
      emit('notify', { message: `Uploading ${file.name}...`, type: "info" })
      await uploadStore.uploadFile(file, null, activeLibraryId.value)
      emit('notify', { message: `${file.name} uploaded. Transcoding queued.`, type: "success" })
      fetchVideos()
    } catch (err) {
      emit('notify', { message: `${file.name} upload failed.`, type: "error" })
    }
  }

  if (event.target) event.target.value = ''
}

let videosInterval = null
let healthInterval = null

onMounted(async () => {
  const savedColor = localStorage.getItem('SV_PRIMARY_COLOR')
  if (savedColor) {
    document.documentElement.style.setProperty('--color-primary', savedColor)
  }

  await fetchLibraries()
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
  <AppLayout
    :active-tab="activeTab"
    @navigate="handleNavigate"
    @open-settings="showSettings = true"
  >
    <input ref="fileInputRef" type="file" class="hidden" accept="video/*,audio/*" multiple @change="handleFileUpload" />

    <div v-if="activeTab === 'library'" class="max-w-7xl mx-auto space-y-8">

      <!-- Page Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold text-slate-900 dark:text-white mb-1">Content Library</h2>
          <p class="text-slate-500 dark:text-slate-400 text-sm">Manage and monitor your VOD assets.</p>
        </div>

        <div class="flex items-center gap-3">
          <div class="flex items-center gap-1.5 relative min-w-[200px]">
            <select
              v-model="activeLibraryId"
              class="w-full px-3 py-2 bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors cursor-pointer"
            >
              <option v-for="lib in libraries" :key="lib.id" :value="lib.id">
                {{ lib.name }}
              </option>
            </select>

            <button
              v-if="activeLibraryId"
              @click="triggerEditLibrary"
              class="p-2 border border-slate-200 dark:border-[#2d3139] bg-white dark:bg-[#1a1d24] hover:bg-slate-50 dark:hover:bg-[#2d3139] rounded-lg transition-colors cursor-pointer"
              title="Rename Library"
            >
              <Pencil class="w-4 h-4 text-slate-500 dark:text-slate-400 hover:text-primary" />
            </button>
          </div>

          <button
            @click="showCreateLibraryModal = true"
            class="px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-[#2d3139] dark:hover:bg-[#3d424e] text-slate-700 dark:text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-1.5 cursor-pointer"
          >
            <Plus class="w-4 h-4" />
            Library
          </button>
        </div>
      </div>


      <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-xl p-6 shadow-sm">
        <div class="flex justify-between items-center mb-4">
          <div>
            <h3 class="font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Key class="w-4 h-4 text-primary" />
              Library Access Keys
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">Keys generated here let external clients stream and sign URLs for this library.</p>
          </div>
          <button
            @click="showCreateKeyModal = true"
            class="px-3 py-1.5 bg-primary hover:bg-rose-600 text-white rounded-lg text-xs font-semibold transition-colors flex items-center gap-1 cursor-pointer"
          >
            <Plus class="w-3.5 h-3.5" />
            New Key
          </button>
        </div>

        <div v-if="libraryKeys.length === 0" class="text-center py-6 text-sm text-slate-400 dark:text-slate-500">
          No playback keys generated for this library.
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="border-b border-slate-100 dark:border-[#2d3139] text-slate-400 text-xs font-semibold uppercase tracking-wider">
                <th class="py-2.5">Key Name</th>
                <th class="py-2.5">Key Secret</th>
                <th class="py-2.5">Status</th>
                <th class="py-2.5">Created At</th>
                <th class="py-2.5 text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="k in libraryKeys" :key="k.id" class="border-b border-slate-100 dark:border-[#2d3139]/50 last:border-0 text-slate-700 dark:text-slate-300">
                <td class="py-3 font-medium text-slate-900 dark:text-white">{{ k.key_name }}</td>
                <td class="py-3">
                  <div class="flex items-center gap-2">
                    <span class="font-mono text-xs px-2.5 py-1 bg-slate-50 dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] rounded-lg select-all">
                      {{ unmaskedKeys[k.id] ? k.playback_secret : '••••••••••••••••••••••••••••••••' }}
                    </span>
                    <button
                      @click="toggleKeyMask(k.id)"
                      class="p-1 hover:bg-slate-100 dark:hover:bg-[#2d3139] rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
                      :title="unmaskedKeys[k.id] ? 'Mask Secret' : 'Unmask Secret'"
                    >
                      <component :is="unmaskedKeys[k.id] ? EyeOff : Eye" class="w-4 h-4" />
                    </button>
                    <button
                      @click="navigator.clipboard.writeText(k.playback_secret); emit('notify', { message: 'Access Key secret copied.', type: 'success' })"
                      class="p-1 hover:bg-slate-100 dark:hover:bg-[#2d3139] rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
                      title="Copy Secret"
                    >
                      <Copy class="w-4 h-4" />
                    </button>
                  </div>
                </td>
                <td class="py-3">
                  <span :class="['px-2 py-0.5 rounded-full text-[10px] font-semibold', k.is_active ? 'bg-emerald-50 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400' : 'bg-slate-100 dark:bg-[#2d3139] text-slate-500']">
                    {{ k.is_active ? 'Active' : 'Revoked' }}
                  </span>
                </td>
                <td class="py-3 text-xs">{{ formatDate(k.created_at) }}</td>
                <td class="py-3 text-right">
                  <div class="flex items-center justify-end gap-2">
                    <button
                      @click="handleRegenerateKey(k.id)"
                      class="p-1.5 hover:bg-amber-500/10 hover:text-amber-500 dark:hover:bg-amber-500/20 text-slate-400 dark:text-slate-500 rounded transition-colors cursor-pointer"
                      title="Regenerate Key Secret"
                    >
                      <RefreshCw class="w-4 h-4" />
                    </button>
                    <button
                      v-if="k.is_active"
                      @click="handleRevokeKey(k.id)"
                      class="p-1.5 hover:bg-rose-500/10 hover:text-rose-500 dark:hover:bg-rose-500/20 text-slate-400 dark:text-slate-500 rounded transition-colors cursor-pointer"
                      title="Revoke (Deactivate) Key"
                    >
                      <Ban class="w-4 h-4" />
                    </button>
                    <button
                      @click="handleDeleteKey(k.id)"
                      class="p-1.5 hover:bg-rose-500/10 hover:text-rose-500 dark:hover:bg-rose-500/20 text-slate-400 dark:text-slate-500 rounded transition-colors cursor-pointer"
                      title="Delete Key Completely"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

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
              <span class="font-medium text-slate-900 dark:text-white truncate">{{ upload.name }}</span>
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

      <VideoFilters :active-filter="activeFilter" @update:active-filter="activeFilter = $event; currentPage = 1" @upload="triggerUpload" />

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

    <div v-else-if="activeTab === 'docs'" class="h-full w-full">
      <Documentation
        :api-key="getAdminKey()"
        :playback-key="getPlaybackKey()"
        :library-id="activeLibraryId"
      />
    </div>

    <UnderConstruction v-else :title="activeTab.charAt(0).toUpperCase() + activeTab.slice(1)" />

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

    <BaseModal :is-open="showCreateLibraryModal" title="Create Library" description="Create a logical grouping for related media assets, each with custom credentials." @close="showCreateLibraryModal = false">
      <template #icon>
        <Library class="w-5 h-5 text-primary" />
      </template>
      <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Library Name</label>
      <input v-model="newLibraryName" placeholder="e.g. Production Team" type="text" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
      <div class="flex gap-3 justify-end mt-6">
        <button @click="showCreateLibraryModal = false" class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer">Cancel</button>
        <button @click="handleCreateLibrary" class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Create</button>
      </div>
    </BaseModal>

    <BaseModal :is-open="showEditLibraryModal" title="Rename Library" description="Modify the name of this library group." @close="showEditLibraryModal = false">
      <template #icon>
        <Library class="w-5 h-5 text-primary" />
      </template>
      <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Library Name</label>
      <input v-model="editingLibraryName" placeholder="e.g. Production Team" type="text" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
      <div class="flex gap-3 justify-end mt-6">
        <button @click="showEditLibraryModal = false" class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer">Cancel</button>
        <button @click="handleUpdateLibrary" class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Save</button>
      </div>
    </BaseModal>

    <BaseModal :is-open="showCreateKeyModal" title="Generate Access Key" description="Create a playback signing token for this library." @close="showCreateKeyModal = false">
      <template #icon>
        <Key class="w-5 h-5 text-primary" />
      </template>
      <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Key Name</label>
      <input v-model="newKeyName" placeholder="e.g. Website Key" type="text" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
      <div class="flex gap-3 justify-end mt-6">
        <button @click="showCreateKeyModal = false" class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer">Cancel</button>
        <button @click="handleCreateKey" class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Generate</button>
      </div>
    </BaseModal>

    <div v-if="generatedKeySecret" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/15 backdrop-blur-[2px]">
      <div class="bg-white dark:bg-[#1a1d24] rounded-2xl p-6 w-full max-w-md border border-slate-200 dark:border-[#2d3139] shadow-2xl">
        <h3 class="text-lg font-bold text-emerald-600 dark:text-emerald-400 flex items-center gap-2 mb-2">
          <Lock class="w-5 h-5" />
          Access Key Generated
        </h3>
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-6">Copy this key secret now. For security, it will not be shown again.</p>

        <div class="flex items-center gap-2 bg-slate-50 dark:bg-[#111318] p-3 rounded-xl border border-slate-200 dark:border-[#2d3139] font-mono text-sm break-all relative">
          <span class="flex-1 select-all pr-10 text-slate-800 dark:text-slate-200">{{ generatedKeySecret }}</span>
          <button
            @click="navigator.clipboard.writeText(generatedKeySecret); emit('notify', { message: 'Secret copied to clipboard.', type: 'success' })"
            class="absolute right-3 p-1.5 hover:bg-slate-200 dark:hover:bg-[#2d3139] rounded transition-colors text-slate-500 cursor-pointer"
            title="Copy Secret"
          >
            <Copy class="w-4 h-4" />
          </button>
        </div>

        <div class="flex justify-end mt-6">
          <button @click="generatedKeySecret = ''" class="px-5 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Done</button>
        </div>
      </div>
    </div>

    <BaseModal :is-open="showEmbedModal" title="Embed Video Player" description="Use the iframe below to embed the ForgePlayer on external sites." max-width="max-w-lg" @close="showEmbedModal = false">
      <template #icon>
        <Code class="w-5 h-5 text-primary" />
      </template>
      <div class="relative bg-slate-950 p-4 rounded-xl font-mono text-xs text-emerald-400 border border-slate-800 break-all pr-12">
        {{ embedCode }}
        <button
          @click="copyEmbedCode"
          class="absolute right-3 top-3 p-1.5 bg-slate-900 hover:bg-slate-800 text-slate-300 rounded transition-colors cursor-pointer border border-slate-800"
          title="Copy Embed Code"
        >
          <component :is="copySuccess ? Check : Copy" class="w-4 h-4" />
        </button>
      </div>
      <div class="flex justify-end mt-6">
        <button @click="showEmbedModal = false" class="px-5 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Done</button>
      </div>
    </BaseModal>

    <BaseModal :is-open="showEditVideoModal" title="Edit Video Details" description="Update the title or move the video to another library." max-width="max-w-md" @close="showEditVideoModal = false">
      <template #icon>
        <Pencil class="w-5 h-5 text-primary" />
      </template>
      <div class="space-y-4">
        <div>
          <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Video Title</label>
          <input v-model="editingVideoTitle" placeholder="e.g. Tutorial Video" type="text" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
        </div>
        <div>
          <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Library</label>
          <select v-model="editingVideoLibraryId" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors cursor-pointer">
            <option v-for="lib in libraries" :key="lib.id" :value="lib.id">
              {{ lib.name }}
            </option>
          </select>
        </div>
      </div>
      <div class="flex gap-3 justify-end mt-8">
        <button @click="showEditVideoModal = false" class="px-5 py-2.5 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer">Cancel</button>
        <button @click="handleUpdateVideo" class="px-5 py-2.5 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Save Changes</button>
      </div>
    </BaseModal>

    <BaseModal :is-open="showUploadModal" title="Upload Media" description="Add videos or audio files to your library." max-width="max-w-lg" @close="showUploadModal = false">
      <template #icon>
        <component :is="Upload" class="w-5 h-5 text-primary" />
      </template>
      <div
        @click="fileInputRef?.click()"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="handleDrop"
        :class="[
          'border-2 border-dashed rounded-xl p-8 text-center transition-all duration-200 cursor-pointer group',
          isDragging
            ? 'border-primary bg-primary/5'
            : 'border-slate-200 dark:border-[#2d3139] hover:border-primary/50 hover:bg-slate-50 dark:hover:bg-[#111318]'
        ]"
      >
        <div class="flex flex-col items-center gap-3">
          <div :class="['w-14 h-14 rounded-2xl flex items-center justify-center transition-all', isDragging ? 'bg-primary/10 text-primary scale-110' : 'bg-slate-100 dark:bg-[#2d3139] text-slate-400 group-hover:bg-primary/10 group-hover:text-primary']">
            <component :is="Upload" class="w-7 h-7" />
          </div>
          <div>
            <p class="font-semibold text-slate-800 dark:text-white text-sm">Drop files here or <span class="text-primary">browse</span></p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">
              Uploading to <span class="font-medium text-slate-600 dark:text-slate-300">{{ libraries.find(l => l.id === activeLibraryId)?.name || 'selected library' }}</span>
            </p>
            <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">MP4, MOV, MKV, WebM &middot; MP3, AAC, WAV, OGG</p>
          </div>
        </div>
      </div>
      <div class="flex justify-end mt-4">
        <button @click="showUploadModal = false" class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer">Cancel</button>
      </div>
    </BaseModal>

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
