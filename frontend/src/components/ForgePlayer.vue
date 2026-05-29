<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'
import Hls from 'hls.js'
import { Loader2, AlertCircle } from 'lucide-vue-next'
import { getStream } from '../api/videos'

const props = defineProps({
  videoId: { type: String, required: true },
  poster: { type: String, default: '' },
  preview: { type: Boolean, default: false },
  isHovered: { type: Boolean, default: true }
})

const videoRef = ref(null)
const error = ref(null)
const errorCode = ref(null)
const isRefreshing = ref(false)
const isVideoLoading = ref(true)
let hls = null
let refreshTimer = null

const isAudioOnly = ref(false)

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
  isAudioOnly.value = false
  errorCode.value = null
  error.value = null
  if (hls) {
    hls.destroy()
    hls = null
  }
  if (refreshTimer) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

const resolveStreamUrl = (url) => {
  if (url && url.includes('/hls/')) {
    return '/hls/' + url.split('/hls/').slice(1).join('/hls/')
  }
  return url
}

const initPlayer = async () => {
  try {
    resetPlayer()
    isVideoLoading.value = true
    const data = await getStream(props.videoId)
    const streamUrl = resolveStreamUrl(data.url)
    
    if (Hls.isSupported()) {
      hls = new Hls({
        xhrSetup: (xhr) => { xhr.withCredentials = true }
      })
      hls.loadSource(streamUrl)
      hls.attachMedia(videoRef.value)
      
      hls.on(Hls.Events.MANIFEST_PARSED, (event, manifestData) => {
        if (hls.levels && hls.levels.length > 0) {
          const firstLevel = hls.levels[0]
          isAudioOnly.value = !firstLevel.width && !firstLevel.height
        } else if (hls.audioTracks && hls.audioTracks.length > 0 && (!hls.levels || hls.levels.length === 0)) {
          isAudioOnly.value = true
        }

        if (props.preview && !props.isHovered && videoRef.value) {
          videoRef.value.currentTime = 0.5
        }
      })

      hls.on(Hls.Events.ERROR, (event, data) => {
        if (data.fatal) {
          if (data.response && data.response.code) {
            errorCode.value = data.response.code
            error.value = data.response.text || 'HTTP Error'
          }
          
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              if (data.response?.code === 410) {
                refreshStream()
              } else if (data.response?.code === 404 || data.response?.code === 403) {
                // Do not recover, it's a hard error
                isVideoLoading.value = false
                hls.destroy()
              } else {
                hls.startLoad()
              }
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              hls.recoverMediaError()
              break;
            default:
              isVideoLoading.value = false
              hls.destroy()
              if (!errorCode.value) {
                error.value = "Failed to load stream"
              }
              break;
          }
        }
      })
    } else if (videoRef.value.canPlayType('application/vnd.apple.mpegurl')) {
      videoRef.value.src = data.url
      videoRef.value.addEventListener('loadedmetadata', () => {
        if (videoRef.value.videoWidth === 0 && videoRef.value.videoHeight === 0) {
          isAudioOnly.value = true
        }
        if (props.preview && !props.isHovered) {
          videoRef.value.currentTime = 0.5
        }
      }, { once: true })
      videoRef.value.addEventListener('error', (e) => {
        isVideoLoading.value = false
        error.value = "Failed to load stream"
        const err = videoRef.value.error
        if (err && err.code) {
           errorCode.value = 400 + err.code // rough fallback
        }
      })
    }

    // Schedule token refresh at 75% of lifetime
    if (data.expires_in) {
      const refreshDelay = (data.expires_in * 0.75) * 1000
      scheduleRefresh(refreshDelay)
    }
  } catch (err) {
    if (!props.preview) {
      error.value = err.response?.data?.error || "Failed to load stream metadata"
      errorCode.value = err.response?.status || 500
    }
    isVideoLoading.value = false
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
    const streamUrl = resolveStreamUrl(data.url)
    
    video.pause()

    if (hls) {
      hls.loadSource(streamUrl)
      hls.once(Hls.Events.MANIFEST_PARSED, () => {
        video.currentTime = currentTime
        if (wasPlaying || (props.preview && props.isHovered)) video.play()
        isRefreshing.value = false
      })
    } else {
      video.src = streamUrl
      video.onloadedmetadata = () => {
        video.currentTime = currentTime
        if (wasPlaying || (props.preview && props.isHovered)) video.play()
        isRefreshing.value = false
      }
    }

    if (data.expires_in) {
      scheduleRefresh((data.expires_in * 0.75) * 1000)
    }
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
  initPlayer()
})

watch(() => props.isHovered, (newVal) => {
  if (props.preview && videoRef.value && !error.value) {
    if (newVal) {
      videoRef.value.play().catch(() => {})
    } else {
      videoRef.value.pause()
    }
  }
})
</script>

<template>
  <div class="relative w-full aspect-video bg-black rounded-lg overflow-hidden group font-sans">
    <!-- Video element hidden if error occurs -->
    <video
      v-show="!error"
      ref="videoRef"
      :class="['w-full h-full object-contain', isAudioOnly ? 'opacity-0 absolute -z-10' : '']"
      :poster="poster"
      :controls="!preview"
      :muted="preview"
      :loop="preview"
      preload="metadata"
      playsinline
      @canplay="handleCanPlay"
      @waiting="handleWaiting"
      @playing="handlePlaying"
    />

    <!-- Audio visualizer overlay -->
    <div
      v-if="isAudioOnly && !error"
      class="absolute inset-0 flex flex-col items-center justify-center bg-gradient-to-br from-slate-900 via-[#1a103c] to-slate-950 text-white z-0 p-6"
    >
      <div class="w-20 h-20 rounded-full bg-primary/10 border border-primary/20 flex items-center justify-center mb-4 shadow-lg shadow-primary/5 animate-pulse">
        <svg
          class="w-10 h-10 text-primary"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3"
          />
        </svg>
      </div>
      <span class="text-sm font-semibold tracking-wide text-slate-200">Audio Broadcast</span>
      <div class="flex items-center gap-1.5 mt-4 h-6">
        <span
          class="w-1 bg-primary rounded-full animate-bounce"
          style="height: 60%; animation-duration: 0.8s;"
        />
        <span
          class="w-1 bg-primary rounded-full animate-bounce"
          style="height: 100%; animation-duration: 1.1s;"
        />
        <span
          class="w-1 bg-primary rounded-full animate-bounce"
          style="height: 40%; animation-duration: 0.7s;"
        />
        <span
          class="w-1 bg-primary rounded-full animate-bounce"
          style="height: 80%; animation-duration: 0.9s;"
        />
        <span
          class="w-1 bg-primary rounded-full animate-bounce"
          style="height: 50%; animation-duration: 1.0s;"
        />
      </div>
    </div>

    <!-- Video Loading Spinner (shown only for active full player) -->
    <div
      v-if="isVideoLoading && !preview && !error"
      class="absolute inset-0 flex flex-col items-center justify-center bg-slate-950/80 backdrop-blur-sm z-10 transition-opacity duration-300"
    >
      <Loader2 class="w-10 h-10 text-primary animate-spin mb-3" />
      <span class="text-sm font-medium text-slate-300">Decrypting & buffering stream...</span>
    </div>

    <!-- Error Overlay with Giant HTTP Code -->
    <div
      v-if="error && !preview"
      class="absolute inset-0 flex flex-col items-center justify-center bg-[#111113] text-white z-50"
    >
      <div class="flex flex-col items-center text-center p-8">
        <div
          v-if="errorCode"
          class="text-8xl md:text-9xl font-black text-rose-600/20 mb-2 leading-none relative select-none"
        >
          {{ errorCode }}
          <span
            class="absolute inset-0 flex items-center justify-center text-rose-500 text-5xl md:text-7xl font-bold bg-clip-text text-transparent bg-gradient-to-b from-rose-400 to-rose-600"
            style="-webkit-text-stroke: 1px rgba(255,255,255,0.1)"
          >
            {{ errorCode }}
          </span>
        </div>
        <div
          v-else
          class="mb-6 bg-rose-500/10 p-4 rounded-full"
        >
          <AlertCircle class="w-12 h-12 text-rose-500" />
        </div>
        
        <h3 class="text-xl font-bold text-slate-200 mb-2 mt-4 tracking-tight">
          Stream Access Denied
        </h3>
        <p class="text-slate-400 text-sm max-w-sm mb-8">
          {{ error }}
        </p>
        
        <button
          class="px-6 py-2.5 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl text-sm font-medium text-slate-300 transition-colors duration-200 flex items-center gap-2"
          @click="initPlayer"
        >
          Retry Connection
        </button>
      </div>
    </div>

    <!-- Refresh Indicator (Subtle) -->
    <div
      v-if="isRefreshing && !preview && !error"
      class="absolute top-4 right-4 animate-spin h-5 w-5 border-2 border-primary border-t-transparent rounded-full"
    />
  </div>
</template>
