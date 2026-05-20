<script setup>
import { X } from 'lucide-vue-next'
import ForgePlayer from '../ForgePlayer.vue'

defineProps({
  isOpen: Boolean,
  videoId: String,
  videos: { type: Array, default: () => [] }
})

defineEmits(['close'])
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6 lg:p-8 bg-slate-900/15 backdrop-blur-[2px]">
    <div class="w-full max-w-4xl bg-black rounded-2xl overflow-hidden shadow-2xl ring-1 ring-white/10">
      <div class="bg-[#1a1d24] p-4 flex justify-between items-center text-white border-b border-white/10">
        <span class="font-medium truncate">{{ videos.find(v => v.id === videoId)?.title }}</span>
        <button @click="$emit('close')" class="p-1 hover:bg-white/10 rounded-lg transition-colors cursor-pointer">
          <X class="w-6 h-6" />
        </button>
      </div>
      <ForgePlayer :video-id="videoId" />
    </div>
  </div>
</template>
