<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import {
  Palette,
  Play,
  Volume2,
  Maximize,
  Settings2,
  Save,
  Loader2,
  CheckCircle2,
  RotateCcw,
  Eye
} from 'lucide-vue-next'
import { getGlobalPlayerConfig, updateGlobalPlayerConfig } from '../api/videos'

const emit = defineEmits(['notify'])

const isLoading = ref(true)
const isSaving = ref(false)
const saveSuccess = ref(false)
const hasChanges = ref(false)

// Player Config state
const config = ref({
  accentColor: '#e11d48',
  controls: {
    playLarge: true,
    play: true,
    progress: true,
    currentTime: true,
    mute: true,
    volume: true,
    fullscreen: true,
    settings: true,
    pip: false,
    speed: true,
    quality: true,
    captions: false,
    airplay: false
  },
  behavior: {
    autoplay: false,
    loop: false,
    clickToPlay: true,
    hideControls: true,
    resetOnEnd: false,
    invertTime: true,
    seekTime: 10
  },
  branding: {
    showWatermark: false,
    watermarkText: '',
    watermarkPosition: 'bottom-right'
  }
})

// Copy for change detection
const savedConfig = ref(null)

watch(config, () => {
  if (savedConfig.value) {
    hasChanges.value = JSON.stringify(config.value) !== JSON.stringify(savedConfig.value)
  }
}, { deep: true })

onMounted(async () => {
  try {
    const data = await getGlobalPlayerConfig()
    if (data && Object.keys(data).length > 0) {
      // Merge saved config with defaults
      if (data.accentColor) config.value.accentColor = data.accentColor
      if (data.controls) config.value.controls = { ...config.value.controls, ...data.controls }
      if (data.behavior) config.value.behavior = { ...config.value.behavior, ...data.behavior }
      if (data.branding) config.value.branding = { ...config.value.branding, ...data.branding }
    }
    savedConfig.value = JSON.parse(JSON.stringify(config.value))
  } catch (err) {
    console.error('Failed to load player config', err)
  } finally {
    isLoading.value = false
  }
})

const saveConfig = async () => {
  isSaving.value = true
  try {
    await updateGlobalPlayerConfig(config.value)
    savedConfig.value = JSON.parse(JSON.stringify(config.value))
    hasChanges.value = false
    saveSuccess.value = true
    emit('notify', { message: 'Player configuration saved successfully.', type: 'success' })
    setTimeout(() => { saveSuccess.value = false }, 2500)
  } catch (err) {
    emit('notify', { message: 'Failed to save player configuration.', type: 'error' })
  } finally {
    isSaving.value = false
  }
}

const resetDefaults = () => {
  config.value = {
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
  }
}

const controlLabels = {
  playLarge: { label: 'Large Play Button', desc: 'Centered play overlay on video' },
  play: { label: 'Play/Pause', desc: 'Play and pause toggle in controls' },
  progress: { label: 'Progress Bar', desc: 'Seekable timeline progress bar' },
  currentTime: { label: 'Time Display', desc: 'Current time and duration readout' },
  mute: { label: 'Mute Toggle', desc: 'Mute/unmute audio button' },
  volume: { label: 'Volume Slider', desc: 'Adjustable volume slider control' },
  fullscreen: { label: 'Fullscreen', desc: 'Enter/exit fullscreen button' },
  settings: { label: 'Settings Menu', desc: 'Gear icon for quality and speed' },
  pip: { label: 'Picture-in-Picture', desc: 'Float video in a mini-window' },
  speed: { label: 'Playback Speed', desc: 'Speed options in settings menu' },
  quality: { label: 'Quality Selector', desc: 'Resolution options in settings' },
  captions: { label: 'Captions/Subtitles', desc: 'Toggle captions on the player' },
  airplay: { label: 'AirPlay', desc: 'Cast to Apple AirPlay devices' }
}

const behaviorLabels = {
  autoplay: { label: 'Autoplay', desc: 'Start playing on load (if browser allows)' },
  loop: { label: 'Loop Video', desc: 'Restart video when it ends' },
  clickToPlay: { label: 'Click to Play', desc: 'Click video area to play/pause' },
  hideControls: { label: 'Auto-hide Controls', desc: 'Hide controls after inactivity' },
  resetOnEnd: { label: 'Reset on End', desc: 'Reset to beginning when finished' },
  invertTime: { label: 'Invert Time', desc: 'Show remaining instead of elapsed' }
}

import { PRESET_COLORS } from '../constants/colors'

const enabledControlsList = computed(() => {
  return Object.entries(config.value.controls)
    .filter(([, v]) => v)
    .map(([k]) => controlLabels[k]?.label || k)
})
</script>

<template>
  <div class="max-w-5xl mx-auto space-y-8">
    <!-- Page Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h2 class="text-2xl font-bold text-slate-900 dark:text-white mb-1 flex items-center gap-3">
          <div class="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center">
            <Palette class="w-5 h-5 text-primary" />
          </div>
          Player UI Configuration
        </h2>
        <p class="text-slate-500 dark:text-slate-400 text-sm">
          Configure the look and behavior of your embedded video player. All embeds share this global configuration.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <button
          class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] bg-white dark:bg-[#1a1d24] hover:bg-slate-50 dark:hover:bg-[#2d3139] rounded-xl text-sm font-medium text-slate-600 dark:text-slate-300 transition-colors flex items-center gap-2 cursor-pointer"
          @click="resetDefaults"
        >
          <RotateCcw class="w-4 h-4" />
          Reset Defaults
        </button>
        <button
          :disabled="isSaving || !hasChanges"
          :class="[
            'px-5 py-2 rounded-xl text-sm font-semibold transition-all flex items-center gap-2 cursor-pointer shadow-sm',
            hasChanges
              ? 'bg-primary hover:bg-rose-600 text-white'
              : 'bg-slate-100 dark:bg-[#2d3139] text-slate-400 dark:text-slate-500 cursor-not-allowed'
          ]"
          @click="saveConfig"
        >
          <Loader2
            v-if="isSaving"
            class="w-4 h-4 animate-spin"
          />
          <CheckCircle2
            v-else-if="saveSuccess"
            class="w-4 h-4"
          />
          <Save
            v-else
            class="w-4 h-4"
          />
          {{ isSaving ? 'Saving...' : saveSuccess ? 'Saved!' : 'Save Configuration' }}
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div
      v-if="isLoading"
      class="flex items-center justify-center py-32"
    >
      <Loader2 class="w-8 h-8 text-primary animate-spin" />
    </div>

    <template v-else>
      <!-- Live Preview -->
      <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-2xl p-6 shadow-sm">
        <div class="flex items-center gap-2 mb-4">
          <Eye class="w-4 h-4 text-primary" />
          <h3 class="font-semibold text-slate-900 dark:text-white">
            Live Preview
          </h3>
          <span class="text-xs text-slate-400 dark:text-slate-500 ml-2">How your player will look</span>
        </div>

        <div
          class="relative rounded-xl overflow-hidden bg-slate-950 aspect-video max-w-2xl mx-auto"
          :style="{ '--plyr-color-main': config.accentColor }"
        >
          <!-- Simulated video area -->
          <div class="absolute inset-0 bg-gradient-to-br from-slate-800 via-slate-900 to-slate-950 flex items-center justify-center">
            <div class="text-center">
              <div
                v-if="config.controls.playLarge"
                class="w-16 h-16 rounded-full flex items-center justify-center mx-auto mb-3 transition-colors"
                :style="{ backgroundColor: config.accentColor + '22', border: `2px solid ${config.accentColor}44` }"
              >
                <Play
                  class="w-7 h-7 ml-1"
                  :style="{ color: config.accentColor }"
                />
              </div>
              <p class="text-slate-500 text-xs">
                Player Preview
              </p>
            </div>
          </div>

          <!-- Branding watermark preview -->
          <div
            v-if="config.branding.showWatermark && config.branding.watermarkText"
            :class="[
              'absolute text-[10px] font-bold text-white/40 px-2 py-0.5 select-none z-20',
              config.branding.watermarkPosition === 'top-left' ? 'top-3 left-3' : '',
              config.branding.watermarkPosition === 'top-right' ? 'top-3 right-3' : '',
              config.branding.watermarkPosition === 'bottom-left' ? 'bottom-12 left-3' : '',
              config.branding.watermarkPosition === 'bottom-right' ? 'bottom-12 right-3' : ''
            ]"
          >
            {{ config.branding.watermarkText }}
          </div>

          <!-- Simulated controls bar -->
          <div class="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-3 pt-8">
            <!-- Progress bar -->
            <div
              v-if="config.controls.progress"
              class="w-full h-1 bg-white/20 rounded-full mb-3 relative cursor-pointer group"
            >
              <div
                class="h-full rounded-full transition-all relative"
                style="width: 35%;"
                :style="{ backgroundColor: config.accentColor }"
              >
                <div
                  class="absolute right-0 top-1/2 -translate-y-1/2 w-3 h-3 rounded-full shadow-md opacity-0 group-hover:opacity-100 transition-opacity"
                  :style="{ backgroundColor: config.accentColor }"
                />
              </div>
            </div>

            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <button
                  v-if="config.controls.play"
                  class="text-white/90 hover:text-white transition-colors"
                >
                  <Play class="w-5 h-5 fill-current" />
                </button>

                <div
                  v-if="config.controls.volume"
                  class="flex items-center gap-1.5"
                >
                  <button
                    v-if="config.controls.mute"
                    class="text-white/90 hover:text-white transition-colors"
                  >
                    <Volume2 class="w-4 h-4" />
                  </button>
                  <div class="w-16 h-1 bg-white/20 rounded-full">
                    <div
                      class="h-full rounded-full w-3/4"
                      :style="{ backgroundColor: config.accentColor }"
                    />
                  </div>
                </div>

                <span
                  v-if="config.controls.currentTime"
                  class="text-white/70 text-xs font-mono"
                >
                  {{ config.behavior.invertTime ? '-2:15' : '1:45' }} / 4:00
                </span>
              </div>

              <div class="flex items-center gap-2">
                <button
                  v-if="config.controls.settings"
                  class="text-white/70 hover:text-white transition-colors"
                >
                  <Settings2 class="w-4 h-4" />
                </button>
                <button
                  v-if="config.controls.fullscreen"
                  class="text-white/70 hover:text-white transition-colors"
                >
                  <Maximize class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Active controls summary -->
        <div class="mt-4 flex flex-wrap gap-1.5">
          <span
            v-for="ctrl in enabledControlsList"
            :key="ctrl"
            class="px-2 py-0.5 bg-slate-100 dark:bg-[#2d3139] text-slate-500 dark:text-slate-400 rounded text-[10px] font-medium"
          >
            {{ ctrl }}
          </span>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Accent Color -->
        <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-2xl p-6 shadow-sm">
          <div class="flex items-center gap-2 mb-4">
            <Palette class="w-4 h-4 text-primary" />
            <h3 class="font-semibold text-slate-900 dark:text-white">
              Accent Color
            </h3>
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mb-4">
            This color applies to the progress bar, volume slider, play buttons, and all interactive highlights.
          </p>

          <!-- Preset swatches -->
          <div class="flex flex-wrap gap-2 mb-4">
            <button
              v-for="preset in PRESET_COLORS"
              :key="preset.hex"
              :title="preset.name"
              :class="[
                'w-8 h-8 rounded-lg transition-all cursor-pointer border-2',
                config.accentColor === preset.hex
                  ? 'border-slate-900 dark:border-white scale-110 shadow-lg'
                  : 'border-transparent hover:scale-105'
              ]"
              :style="{ backgroundColor: preset.hex }"
              @click="config.accentColor = preset.hex"
            />
          </div>

          <!-- Custom color picker -->
          <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-[#111318] rounded-xl border border-slate-100 dark:border-[#2d3139]">
            <input
              v-model="config.accentColor"
              type="color"
              class="w-10 h-10 rounded-lg cursor-pointer bg-transparent border-0 p-0"
            >
            <div>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                Custom Hex
              </p>
              <input
                v-model="config.accentColor"
                type="text"
                maxlength="7"
                class="text-sm font-mono bg-transparent text-slate-900 dark:text-white focus:outline-none w-24"
              >
            </div>
          </div>
        </div>

        <!-- Branding -->
        <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-2xl p-6 shadow-sm">
          <div class="flex items-center gap-2 mb-4">
            <Eye class="w-4 h-4 text-primary" />
            <h3 class="font-semibold text-slate-900 dark:text-white">
              Branding
            </h3>
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mb-4">
            Add subtle branding overlays on the player for visual ownership.
          </p>

          <div class="space-y-4">
            <label class="flex items-center justify-between cursor-pointer">
              <div>
                <p class="text-sm font-medium text-slate-700 dark:text-slate-200">Show Watermark</p>
                <p class="text-xs text-slate-400 dark:text-slate-500">Overlay text on the video</p>
              </div>
              <button
                :class="[
                  'relative inline-flex h-6 w-11 items-center rounded-full transition-colors',
                  config.branding.showWatermark ? 'bg-primary' : 'bg-slate-200 dark:bg-[#2d3139]'
                ]"
                @click="config.branding.showWatermark = !config.branding.showWatermark"
              >
                <span
                  :class="[
                    'inline-block h-4 w-4 transform rounded-full bg-white transition-transform shadow',
                    config.branding.showWatermark ? 'translate-x-6' : 'translate-x-1'
                  ]"
                />
              </button>
            </label>

            <div
              v-if="config.branding.showWatermark"
              class="space-y-3 pl-1"
            >
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Watermark Text</label>
                <input
                  v-model="config.branding.watermarkText"
                  type="text"
                  placeholder="e.g. YourBrand"
                  class="w-full px-3 py-2 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                >
              </div>
              <div>
                <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Position</label>
                <select
                  v-model="config.branding.watermarkPosition"
                  class="w-full px-3 py-2 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 cursor-pointer"
                >
                  <option value="top-left">
                    Top Left
                  </option>
                  <option value="top-right">
                    Top Right
                  </option>
                  <option value="bottom-left">
                    Bottom Left
                  </option>
                  <option value="bottom-right">
                    Bottom Right
                  </option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Controls Grid -->
      <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-2xl p-6 shadow-sm">
        <div class="flex items-center gap-2 mb-1">
          <Play class="w-4 h-4 text-primary" />
          <h3 class="font-semibold text-slate-900 dark:text-white">
            Player Controls
          </h3>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-5">
          Toggle which controls appear in the player UI. Disabled controls are completely removed from the player.
        </p>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <label
            v-for="(info, key) in controlLabels"
            :key="key"
            :class="[
              'flex items-center justify-between p-3 rounded-xl border cursor-pointer transition-all',
              config.controls[key]
                ? 'border-primary/30 bg-primary/5 dark:bg-primary/5'
                : 'border-slate-100 dark:border-[#2d3139] bg-slate-50 dark:bg-[#111318]'
            ]"
          >
            <div class="mr-3">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ info.label }}</p>
              <p class="text-[10px] text-slate-400 dark:text-slate-500 leading-tight">{{ info.desc }}</p>
            </div>
            <button
              :class="[
                'relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0',
                config.controls[key] ? 'bg-primary' : 'bg-slate-200 dark:bg-[#2d3139]'
              ]"
              @click="config.controls[key] = !config.controls[key]"
            >
              <span
                :class="[
                  'inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform shadow',
                  config.controls[key] ? 'translate-x-[18px]' : 'translate-x-1'
                ]"
              />
            </button>
          </label>
        </div>
      </div>

      <!-- Behavior Settings -->
      <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-2xl p-6 shadow-sm">
        <div class="flex items-center gap-2 mb-1">
          <Settings2 class="w-4 h-4 text-primary" />
          <h3 class="font-semibold text-slate-900 dark:text-white">
            Playback Behavior
          </h3>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-5">
          Configure how the player behaves by default when embedded on external websites.
        </p>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label
            v-for="(info, key) in behaviorLabels"
            :key="key"
            :class="[
              'flex items-center justify-between p-3 rounded-xl border cursor-pointer transition-all',
              config.behavior[key]
                ? 'border-primary/30 bg-primary/5 dark:bg-primary/5'
                : 'border-slate-100 dark:border-[#2d3139] bg-slate-50 dark:bg-[#111318]'
            ]"
          >
            <div class="mr-3">
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ info.label }}</p>
              <p class="text-[10px] text-slate-400 dark:text-slate-500 leading-tight">{{ info.desc }}</p>
            </div>
            <button
              :class="[
                'relative inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0',
                config.behavior[key] ? 'bg-primary' : 'bg-slate-200 dark:bg-[#2d3139]'
              ]"
              @click="config.behavior[key] = !config.behavior[key]"
            >
              <span
                :class="[
                  'inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform shadow',
                  config.behavior[key] ? 'translate-x-[18px]' : 'translate-x-1'
                ]"
              />
            </button>
          </label>

          <!-- Seek time - special numeric control -->
          <div class="flex items-center justify-between p-3 rounded-xl border border-slate-100 dark:border-[#2d3139] bg-slate-50 dark:bg-[#111318]">
            <div>
              <p class="text-sm font-medium text-slate-800 dark:text-slate-200">
                Seek Time
              </p>
              <p class="text-[10px] text-slate-400 dark:text-slate-500 leading-tight">
                Seconds to skip with arrows
              </p>
            </div>
            <select
              v-model.number="config.behavior.seekTime"
              class="px-2 py-1 bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-lg text-sm cursor-pointer focus:outline-none"
            >
              <option :value="5">
                5s
              </option>
              <option :value="10">
                10s
              </option>
              <option :value="15">
                15s
              </option>
              <option :value="30">
                30s
              </option>
            </select>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
