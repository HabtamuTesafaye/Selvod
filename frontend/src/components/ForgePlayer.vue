<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'
import Hls from 'hls.js'
import { Loader2 } from 'lucide-vue-next'
import { getStream } from '../api/videos'

const props = defineProps({
  videoId: { type: String, required: true },
  poster: { type: String, default: '' },
  preview: { type: Boolean, default: false },
  isHovered: { type: Boolean, default: true }
})

const videoRef = ref(null)
const error = ref(null)
const isRefreshing = ref(false)
const isVideoLoading = ref(true)
let hls = null
let refreshTimer = null

const handleCanPlay = () => {
  isVideoLoading.value = false
}

const handleWaiting = () => {
  isVideoLoading.value = true
}

const handlePlaying = () => {
  isVideoLoading.value = false
}

const resetPlayer = () => {
  if (hls) {
    hls.destroy()
    hls = null
  }
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

const initPlayer = async () => {
  try {
    error.value = null
    const data = await getStream(props.videoId)
    
    if (Hls.isSupported()) {
      hls = new Hls()
      hls.loadSource(data.url)
      hls.attachMedia(videoRef.value)
      
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        if (props.preview && !props.isHovered && videoRef.value) {
          // Seek forward slightly to load a beautiful frame and avoid any black screen
          videoRef.value.currentTime = 0.5
        }
      })

      hls.on(Hls.Events.ERROR, (event, data) => {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              if (data.response?.code === 410) {
                refreshStream()
              } else {
                hls.startLoad()
              }
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              hls.recoverMediaError()
              break;
            default:
              initPlayer()
              break;
          }
        }
      })
    } else if (videoRef.value.canPlayType('application/vnd.apple.mpegurl')) {
      videoRef.value.src = data.url
      if (props.preview && !props.isHovered) {
        videoRef.value.addEventListener('loadedmetadata', () => {
          videoRef.value.currentTime = 0.5
        }, { once: true })
      }
    }

    // Schedule token refresh at 75% of lifetime
    const refreshDelay = (data.expires_in * 0.75) * 1000
    scheduleRefresh(refreshDelay)
  } catch (err) {
    if (!props.preview) {
      error.value = "Failed to load stream metadata"
    }
    console.error(err)
  }
}

const refreshStream = async () => {
  if (isRefreshing.value || !videoRef.value) return
  isRefreshing.value = true
  
  const video = videoRef.value
  const currentTime = video.currentTime
  const wasPlaying = !video.paused

  try {
    const data = await getStream(props.videoId)
    
    video.pause()

    if (hls) {
      hls.loadSource(data.url)
      hls.once(Hls.Events.MANIFEST_PARSED, () => {
        video.currentTime = currentTime
        if (wasPlaying || (props.preview && props.isHovered)) video.play()
        isRefreshing.value = false
      })
    } else {
      video.src = data.url
      video.onloadedmetadata = () => {
        video.currentTime = currentTime
        if (wasPlaying || (props.preview && props.isHovered)) video.play()
        isRefreshing.value = false
      }
    }

    scheduleRefresh((data.expires_in * 0.75) * 1000)
  } catch (err) {
    console.error("Token refresh failed", err)
    isRefreshing.value = false
  }
}

const scheduleRefresh = (delay) => {
  if (!Number.isFinite(delay) || delay <= 0) return
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = setTimeout(refreshStream, delay)
}

onMounted(() => {
  initPlayer()
})

onUnmounted(() => {
  resetPlayer()
})

watch(() => props.videoId, () => {
  resetPlayer()
  initPlayer()
})

watch(() => props.isHovered, (newVal) => {
  if (props.preview && videoRef.value) {
    if (newVal) {
      videoRef.value.play().catch(() => {})
    } else {
      videoRef.value.pause()
    }
  }
})
</script>

<template>
  <div class="relative w-full aspect-video bg-black rounded-lg overflow-hidden group">
    <video
      ref="videoRef"
      class="w-full h-full object-contain"
      :poster="poster"
      :controls="!preview"
      :muted="preview"
      :loop="preview"
      preload="metadata"
      playsinline
      @canplay="handleCanPlay"
      @waiting="handleWaiting"
      @playing="handlePlaying"
    ></video>

    <!-- Video Loading Spinner (shown only for active full player) -->
    <div v-if="isVideoLoading && !preview" class="absolute inset-0 flex flex-col items-center justify-center bg-slate-950/80 backdrop-blur-sm z-10 transition-opacity duration-300">
      <Loader2 class="w-10 h-10 text-primary animate-spin mb-3" />
      <span class="text-sm font-medium text-slate-300">Decrypting & buffering stream...</span>
    </div>

    <!-- Error Overlay -->
    <div v-if="error && !preview" class="absolute inset-0 flex items-center justify-center bg-black/80 text-white p-4 text-center">
      <div>
        <p class="text-lg font-bold mb-2">{{ error }}</p>
        <button @click="initPlayer" class="px-4 py-2 bg-blue-600 rounded-md hover:bg-blue-500 transition">
          Retry Loading
        </button>
      </div>
    </div>

    <!-- Refresh Indicator (Subtle) -->
    <div v-if="isRefreshing && !preview" class="absolute top-4 right-4 animate-spin h-5 w-5 border-2 border-blue-600 border-t-transparent rounded-full"></div>
  </div>
</template>
