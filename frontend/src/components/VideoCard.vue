<script setup>
import { ref } from 'vue'
import { CheckCircle2, Clock, AlertCircle, Play, Trash2, HardDrive, Share2, Pencil } from 'lucide-vue-next'
import ForgePlayer from './ForgePlayer.vue'

const props = defineProps({
  video: { type: Object, required: true }
})

const emit = defineEmits(['play', 'delete', 'embed', 'edit'])

const isHovered = ref(false)

const formatDate = (date) => new Date(date).toLocaleDateString()
const formatSize = (bytes) => (bytes / (1024 * 1024)).toFixed(1) + ' MB'

const statusConfig = {
  pending: { icon: Clock, color: 'text-amber-500', bg: 'bg-amber-50', label: 'Queued' },
  transcoding: { icon: Clock, color: 'text-blue-500', bg: 'bg-blue-50', label: 'Processing', pulse: true },
  completed: { icon: CheckCircle2, color: 'text-emerald-500', bg: 'bg-emerald-50', label: 'Ready' },
  failed: { icon: AlertCircle, color: 'text-rose-500', bg: 'bg-rose-50', label: 'Failed' }
}

const config = statusConfig[props.video.status] || statusConfig.pending
</script>

<template>
  <div class="bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] rounded-xl overflow-hidden hover:shadow-md dark:hover:shadow-black/50 transition-shadow">
    <div 
      class="aspect-video bg-slate-100 dark:bg-[#111318] relative group cursor-pointer overflow-hidden" 
      @click="emit('play', video.id)"
      @mouseenter="isHovered = true"
      @mouseleave="isHovered = false"
    >
      <!-- Preview Player (always mounted so the browser can load and render the first frame as a thumbnail) -->
      <ForgePlayer v-if="video.status === 'completed'" :video-id="video.id" preview :is-hovered="isHovered" class="absolute inset-0 z-0 object-cover scale-105 pointer-events-none" />
      
      <!-- Play Button Overlay on Hover -->
      <div v-if="video.status === 'completed'" class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/40 z-10">
        <div class="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg transform group-hover:scale-110 transition-transform backdrop-blur-sm cursor-pointer">
          <Play class="w-6 h-6 text-primary fill-current ml-1" />
        </div>
      </div>
      <div v-else-if="video.status !== 'completed'" class="absolute inset-0 flex items-center justify-center bg-slate-50 dark:bg-[#111318] z-10">
        <Clock :class="['w-10 h-10', config.color, config.pulse ? 'animate-pulse' : '']" />
      </div>
      
      <!-- Duration Badge Placeholder (If we added duration logic) -->
      <div class="absolute bottom-2 right-2 bg-black/70 backdrop-blur-md text-white text-[10px] font-bold px-1.5 py-0.5 rounded z-10 cursor-default">
        {{ video.duration ? new Date(video.duration * 1000).toISOString().substr(14, 5) : '--:--' }}
      </div>
    </div>
    
    <div class="p-4">
      <div class="flex justify-between items-start mb-2">
        <h3 class="font-semibold text-slate-900 dark:text-slate-100 truncate flex-1 cursor-pointer hover:text-primary transition-colors" @click="emit('play', video.id)">{{ video.title }}</h3>
        <span :class="['px-2 py-1 rounded-full text-xs font-medium flex items-center gap-1 dark:bg-opacity-10', config.bg, config.color]">
          <component :is="config.icon" class="w-3 h-3" />
          {{ config.label }}
        </span>
      </div>

      <p class="text-xs text-slate-500 dark:text-slate-400 mb-2 truncate">ID: {{ video.id.split('-')[0] }}-{{ video.id.split('-')[1].substring(0,4) }}</p>

      <div class="flex items-center gap-3 text-xs text-slate-500 dark:text-slate-400 mt-3">
        <span class="flex items-center gap-1">
          <HardDrive class="w-3 h-3" />
          {{ formatSize(video.total_size_bytes || video.upload_size_bytes) }}
        </span>
        <span>{{ formatDate(video.created_at) }}</span>
      </div>

      <div v-if="video.error_message" class="mt-2 p-2 bg-rose-50 dark:bg-rose-900/20 rounded border border-rose-100 dark:border-rose-900/50 text-[10px] text-rose-600 dark:text-rose-400 line-clamp-2">
        {{ video.error_message }}
      </div>

      <div class="mt-4 pt-4 border-t border-slate-100 dark:border-[#2d3139] flex justify-between items-center">
        <div class="flex items-center gap-2">
          <button 
            v-if="video.status === 'completed'"
            @click="emit('play', video.id)"
            class="text-sm font-medium text-primary hover:text-rose-600 transition-colors cursor-pointer"
          >
            Watch
          </button>
          <span v-else class="text-sm text-slate-400 dark:text-slate-500">Processing...</span>

          <button 
            v-if="video.status === 'completed'"
            @click="emit('embed', video.id)"
            class="p-1.5 text-slate-400 dark:text-slate-500 hover:text-primary rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            title="Embed Code"
          >
            <Share2 class="w-4 h-4" />
          </button>
        </div>
        
        <div class="flex items-center gap-1">
          <button 
            @click="emit('edit', video)" 
            class="p-1.5 text-slate-400 dark:text-slate-500 hover:text-primary rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors cursor-pointer"
            title="Edit Video"
          >
            <Pencil class="w-4 h-4" />
          </button>
          
          <button 
            @click="emit('delete', video.id)" 
            class="p-1.5 text-slate-400 dark:text-slate-500 hover:text-rose-600 dark:hover:text-rose-500 rounded-lg hover:bg-rose-50 dark:hover:bg-rose-900/20 transition-colors cursor-pointer"
            title="Delete Video"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
