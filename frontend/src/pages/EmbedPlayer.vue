<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import Hls from 'hls.js'
import axios from 'axios'
import Plyr from 'plyr'
import 'plyr/dist/plyr.css'

// Read params from URL: /embed.html?videoId=xxx&key=yyy
const params = new URLSearchParams(window.location.search)
const videoId = params.get('videoId') || params.get('v')
const libraryKey = params.get('key') || params.get('k')

const playerConfig = ref({
  accentColor: '#e11d48',
  controls: {
    playLarge: true, play: true, progress: true, currentTime: true,
    mute: true, volume: true, fullscreen: true, settings: true,
    pip: false, speed: true, quality: true, captions: false, airplay: false
  },
  behavior: {
    autoplay: false, loop: false, clickToPlay: true, hideControls: true,
    resetOnEnd: false, invertTime: true, seekTime: 10
  },
  branding: {
    showWatermark: false, watermarkText: '', watermarkPosition: 'bottom-right'
  }
})

const videoRef = ref(null)
const error = ref(null)
const errorCode = ref(null)
const isLoading = ref(true)
const isRefreshing = ref(false)
const isAudioOnly = ref(false)

let hls = null
let refreshTimer = null
let player = null

const clientApi = axios.create({ baseURL: '/api/v1' })

// Attach the library playback key as Bearer token
clientApi.interceptors.request.use((config) => {
  if (libraryKey) {
    config.headers.Authorization = `Bearer ${libraryKey}`
  }
  return config
})

async function getStream() {
  const { data } = await clientApi.get(`/videos/${videoId}/stream`)
  return data
}

function resolveStreamUrl(url) {
  if (url && url.includes('/hls/')) {
    return '/hls/' + url.split('/hls/').slice(1).join('/hls/')
  }
  return url
}

async function initPlayer() {
  // Validate required params
  if (!videoId || !libraryKey) {
    errorCode.value = 400
    error.value = !videoId
      ? 'Missing required parameter: videoId'
      : 'Missing required parameter: key'
    isLoading.value = false
    return
  }

  try {
    resetPlayer()
    isLoading.value = true

    // Fetch global player config
    try {
      const { data: cfgData } = await axios.get('/api/v1/config/player')
      if (cfgData && Object.keys(cfgData).length > 0) {
        if (cfgData.accentColor) playerConfig.value.accentColor = cfgData.accentColor
        if (cfgData.controls) playerConfig.value.controls = { ...playerConfig.value.controls, ...cfgData.controls }
        if (cfgData.behavior) playerConfig.value.behavior = { ...playerConfig.value.behavior, ...cfgData.behavior }
        if (cfgData.branding) playerConfig.value.branding = { ...playerConfig.value.branding, ...cfgData.branding }
      }
    } catch (e) {
      console.warn('Could not fetch player config, using defaults', e)
    }

    // Build Plyr controls list from config
    const plyrControls = []
    const cfg = playerConfig.value.controls
    if (cfg.playLarge) plyrControls.push('play-large')
    if (cfg.play) plyrControls.push('play')
    if (cfg.progress) plyrControls.push('progress')
    if (cfg.currentTime) plyrControls.push('current-time')
    if (cfg.mute) plyrControls.push('mute')
    if (cfg.volume) plyrControls.push('volume')
    if (cfg.captions) plyrControls.push('captions')
    if (cfg.settings) plyrControls.push('settings')
    if (cfg.pip) plyrControls.push('pip')
    if (cfg.airplay) plyrControls.push('airplay')
    if (cfg.fullscreen) plyrControls.push('fullscreen')

    const plyrSettings = []
    if (cfg.quality) plyrSettings.push('quality')
    if (cfg.speed) plyrSettings.push('speed')

    // Initialize Plyr
    if (!player && videoRef.value) {
      player = new Plyr(videoRef.value, {
        controls: plyrControls,
        settings: plyrSettings,
        invertTime: playerConfig.value.behavior.invertTime,
        seekTime: playerConfig.value.behavior.seekTime,
        clickToPlay: playerConfig.value.behavior.clickToPlay,
        hideControls: playerConfig.value.behavior.hideControls,
        resetOnEnd: playerConfig.value.behavior.resetOnEnd,
        loop: { active: playerConfig.value.behavior.loop },
        autoplay: playerConfig.value.behavior.autoplay
      })
    }

    const data = await getStream()
    const streamUrl = resolveStreamUrl(data.url)

    if (Hls.isSupported()) {
      hls = new Hls({
        xhrSetup: (xhr) => {
          xhr.withCredentials = true
        }
      })
      hls.loadSource(streamUrl)
      hls.attachMedia(videoRef.value)

      hls.on(Hls.Events.MANIFEST_PARSED, (event, manifestData) => {
        isLoading.value = false
        if (hls.levels && hls.levels.length > 0) {
          const firstLevel = hls.levels[0]
          isAudioOnly.value = !firstLevel.width && !firstLevel.height
        }
      })

      hls.on(Hls.Events.ERROR, (event, data) => {
        if (data.fatal) {
          if (data.response && data.response.code) {
            errorCode.value = data.response.code
            error.value = getErrorMessage(data.response.code)
          }

          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              if (data.response?.code === 410) {
                refreshStream()
              } else if (data.response?.code === 403 || data.response?.code === 404) {
                isLoading.value = false
                hls.destroy()
              } else {
                hls.startLoad()
              }
              break
            case Hls.ErrorTypes.MEDIA_ERROR:
              hls.recoverMediaError()
              break
            default:
              isLoading.value = false
              hls.destroy()
              if (!errorCode.value) {
                error.value = 'Failed to load stream'
              }
              break
          }
        }
      })
    } else if (videoRef.value?.canPlayType('application/vnd.apple.mpegurl')) {
      // Safari native HLS
      videoRef.value.src = streamUrl
      videoRef.value.addEventListener('loadedmetadata', () => {
        isLoading.value = false
        if (videoRef.value.videoWidth === 0 && videoRef.value.videoHeight === 0) {
          isAudioOnly.value = true
        }
      }, { once: true })
      videoRef.value.addEventListener('error', () => {
        isLoading.value = false
        error.value = 'Failed to load stream'
        errorCode.value = 500
      })
    }

    // Schedule token refresh at 75% of lifetime
    if (data.expires_in) {
      scheduleRefresh((data.expires_in * 0.75) * 1000)
    }
  } catch (err) {
    isLoading.value = false
    const status = err.response?.status
    errorCode.value = status || 500
    error.value = getErrorMessage(status, err.response?.data?.error)
    console.error('Embed player init failed:', err)
  }
}

function getErrorMessage(code, fallback) {
  switch (code) {
    case 401: return 'Unauthorized — invalid or missing playback key'
    case 403: return 'Access denied — this key cannot access this video'
    case 404: return 'Video not found'
    case 423: return 'Video is still being processed'
    default: return fallback || 'An unexpected error occurred'
  }
}

async function refreshStream() {
  if (isRefreshing.value || !videoRef.value) return
  isRefreshing.value = true

  const video = videoRef.value
  const currentTime = video.currentTime
  const wasPlaying = !video.paused

  try {
    const data = await getStream()
    const streamUrl = resolveStreamUrl(data.url)

    video.pause()

    if (hls) {
      hls.loadSource(streamUrl)
      hls.once(Hls.Events.MANIFEST_PARSED, () => {
        video.currentTime = currentTime
        if (wasPlaying) video.play()
        isRefreshing.value = false
      })
    } else {
      video.src = streamUrl
      video.onloadedmetadata = () => {
        video.currentTime = currentTime
        if (wasPlaying) video.play()
        isRefreshing.value = false
      }
    }

    if (data.expires_in) {
      scheduleRefresh((data.expires_in * 0.75) * 1000)
    }
  } catch (err) {
    console.error('Token refresh failed:', err)
    isRefreshing.value = false
  }
}

function scheduleRefresh(delay) {
  if (!Number.isFinite(delay) || delay <= 0) return
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = setTimeout(refreshStream, delay)
}

function resetPlayer() {
  isAudioOnly.value = false
  errorCode.value = null
  error.value = null
  if (hls) { hls.destroy(); hls = null }
  if (refreshTimer) { clearTimeout(refreshTimer); refreshTimer = null }
  if (player) { player.destroy(); player = null }
}

onMounted(() => { initPlayer() })
onUnmounted(() => { resetPlayer() })
</script>

<template>
  <div
    class="embed-root"
    :style="{ '--plyr-color-main': playerConfig.accentColor }"
  >
    <!-- Video element -->
    <video
      v-show="!error"
      ref="videoRef"
      :class="['embed-video', isAudioOnly ? 'audio-hidden' : '']"
      preload="metadata"
      playsinline
      @canplay="isLoading = false"
      @waiting="isLoading = true"
      @playing="isLoading = false"
    />

    <!-- Branding Watermark -->
    <div
      v-if="playerConfig.branding.showWatermark && playerConfig.branding.watermarkText && !error"
      :class="[
        'absolute z-10 pointer-events-none select-none text-white/50 font-bold text-sm px-3 py-1',
        playerConfig.branding.watermarkPosition === 'top-left' ? 'top-4 left-4' : '',
        playerConfig.branding.watermarkPosition === 'top-right' ? 'top-4 right-4' : '',
        playerConfig.branding.watermarkPosition === 'bottom-left' ? 'bottom-16 left-4' : '',
        playerConfig.branding.watermarkPosition === 'bottom-right' ? 'bottom-16 right-4' : ''
      ]"
    >
      {{ playerConfig.branding.watermarkText }}
    </div>

    <!-- Audio-only visualizer -->
    <div
      v-if="isAudioOnly && !error"
      class="audio-overlay"
    >
      <div class="audio-icon-ring">
        <svg
          class="audio-icon"
          viewBox="0 0 24 24"
          fill="none"
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
      <span class="audio-label">Audio Stream</span>
      <div class="audio-bars">
        <span
          class="bar"
          style="animation-duration: 0.8s;"
        />
        <span
          class="bar"
          style="animation-duration: 1.1s;"
        />
        <span
          class="bar"
          style="animation-duration: 0.7s;"
        />
        <span
          class="bar"
          style="animation-duration: 0.9s;"
        />
        <span
          class="bar"
          style="animation-duration: 1.0s;"
        />
      </div>
    </div>

    <!-- Loading spinner -->
    <div
      v-if="isLoading && !error"
      class="loading-overlay"
    >
      <div class="spinner" />
      <span class="loading-text">Loading stream...</span>
    </div>

    <!-- Error state -->
    <div
      v-if="error"
      class="error-overlay"
    >
      <div class="error-content">
        <div
          v-if="errorCode"
          class="error-code-bg"
        >
          {{ errorCode }}
          <span class="error-code-fg">{{ errorCode }}</span>
        </div>
        <div
          v-else
          class="error-icon-wrap"
        >
          <svg
            class="error-icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle
              cx="12"
              cy="12"
              r="10"
            />
            <line
              x1="12"
              y1="8"
              x2="12"
              y2="12"
            />
            <line
              x1="12"
              y1="16"
              x2="12.01"
              y2="16"
            />
          </svg>
        </div>
        <h3 class="error-title">
          Playback Failed
        </h3>
        <p class="error-message">
          {{ error }}
        </p>
        <button
          class="retry-btn"
          @click="initPlayer"
        >
          Retry
        </button>
      </div>
    </div>

    <!-- Refresh indicator -->
    <div
      v-if="isRefreshing && !error"
      class="refresh-indicator"
    />

    <!-- Watermark -->
    <div
      v-if="!error && !isLoading"
      class="watermark"
    >
      SELVOD
    </div>
  </div>
</template>

<style>
.embed-root {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: #000;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  overflow: hidden;
}

/* Force Plyr to take full height of the container */
.plyr {
  width: 100%;
  height: 100%;
}

.plyr__video-wrapper {
  height: 100% !important;
  padding-bottom: 0 !important;
  margin: 0 !important;
}

.plyr__video-wrapper video {
  height: 100% !important;
  width: 100% !important;
  object-fit: contain !important;
}

.embed-video {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.embed-video.audio-hidden {
  opacity: 0;
  position: absolute;
  z-index: -1;
}

/* Loading */
.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(8px);
  z-index: 10;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: #8b5cf6;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
}

/* Error */
.error-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0a0a0c;
  z-index: 50;
}

.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 32px;
  max-width: 400px;
}

.error-code-bg {
  font-size: clamp(64px, 15vw, 120px);
  font-weight: 900;
  color: rgba(239, 68, 68, 0.12);
  line-height: 1;
  position: relative;
  user-select: none;
  margin-bottom: 8px;
}

.error-code-fg {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: clamp(40px, 10vw, 80px);
  font-weight: 800;
  background: linear-gradient(to bottom, #f87171, #dc2626);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.error-icon-wrap {
  background: rgba(239, 68, 68, 0.1);
  padding: 16px;
  border-radius: 50%;
  margin-bottom: 24px;
}

.error-icon {
  width: 48px;
  height: 48px;
  color: #ef4444;
}

.error-title {
  font-size: 20px;
  font-weight: 700;
  color: #e2e8f0;
  margin: 16px 0 8px;
  letter-spacing: -0.02em;
}

.error-message {
  color: #64748b;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 28px;
}

.retry-btn {
  padding: 10px 24px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #cbd5e1;
  font-size: 13px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

/* Audio */
.audio-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a, #1a103c, #0f172a);
  color: #fff;
  z-index: 0;
  padding: 24px;
}

.audio-icon-ring {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: rgba(139, 92, 246, 0.1);
  border: 1px solid rgba(139, 92, 246, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  animation: pulse-ring 2s ease-in-out infinite;
}

@keyframes pulse-ring {
  0%, 100% { box-shadow: 0 0 0 0 rgba(139, 92, 246, 0.1); }
  50% { box-shadow: 0 0 0 12px rgba(139, 92, 246, 0); }
}

.audio-icon {
  width: 36px;
  height: 36px;
  color: #8b5cf6;
}

.audio-label {
  font-size: 14px;
  font-weight: 600;
  color: #cbd5e1;
  letter-spacing: 0.05em;
}

.audio-bars {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 16px;
  height: 24px;
}

.bar {
  width: 3px;
  background: #8b5cf6;
  border-radius: 4px;
  animation: bounce-bar ease-in-out infinite;
}

.bar:nth-child(1) { height: 60%; }
.bar:nth-child(2) { height: 100%; }
.bar:nth-child(3) { height: 40%; }
.bar:nth-child(4) { height: 80%; }
.bar:nth-child(5) { height: 50%; }

@keyframes bounce-bar {
  0%, 100% { transform: scaleY(0.4); }
  50% { transform: scaleY(1); }
}

/* Refresh indicator */
.refresh-indicator {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 20px;
  height: 20px;
  border: 2px solid #8b5cf6;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* Watermark */
.watermark {
  position: absolute;
  bottom: 52px;
  right: 12px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.15em;
  color: rgba(255, 255, 255, 0.15);
  pointer-events: none;
  user-select: none;
  z-index: 5;
}
</style>
