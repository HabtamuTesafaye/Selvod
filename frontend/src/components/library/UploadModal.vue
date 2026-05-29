<script setup>
import { ref } from 'vue'
import { useUploadStore } from '../../stores/upload'
import BaseModal from '../modals/BaseModal.vue'
import { Upload, Activity, Check } from 'lucide-vue-next'

const props = defineProps({
  isOpen: { type: Boolean, required: true },
  activeLibraryId: { type: String, required: true },
  libraries: { type: Array, required: true }
})

const emit = defineEmits(['close', 'notify', 'upload-complete'])

const uploadStore = useUploadStore()
const fileInputRef = ref(null)
const isDragging = ref(false)

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
      await uploadStore.uploadFile(file, null, props.activeLibraryId)
      emit('notify', { message: `${file.name} uploaded. Transcoding queued.`, type: 'success' })
      emit('upload-complete')
    } catch {
      emit('notify', { message: `${file.name} upload failed.`, type: 'error' })
    }
  }
}

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files.length) return
  for (const file of files) {
    try {
      emit('notify', { message: `Uploading ${file.name}...`, type: "info" })
      await uploadStore.uploadFile(file, null, props.activeLibraryId)
      emit('notify', { message: `${file.name} uploaded. Transcoding queued.`, type: "success" })
      emit('upload-complete')
    } catch {
      emit('notify', { message: `${file.name} upload failed.`, type: "error" })
    }
  }
  if (event.target) event.target.value = ''
}
</script>

<template>
  <BaseModal
    :is-open="isOpen"
    title="Upload Media"
    description="Add videos or audio files to your library."
    max-width="max-w-lg"
    @close="$emit('close')"
  >
    <template #icon>
      <Upload class="w-5 h-5 text-primary" />
    </template>
    <div
      :class="[
        'border-2 border-dashed rounded-xl p-8 text-center transition-all duration-200 cursor-pointer group',
        isDragging
          ? 'border-primary bg-primary/5'
          : 'border-slate-200 dark:border-[#2d3139] hover:border-primary/50 hover:bg-slate-50 dark:hover:bg-[#111318]'
      ]"
      @click="fileInputRef?.click()"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
    >
      <input
        ref="fileInputRef"
        type="file"
        class="hidden"
        accept="video/*,audio/*"
        multiple
        @change="handleFileUpload"
      >
      <div class="flex flex-col items-center gap-3">
        <div :class="['w-14 h-14 rounded-2xl flex items-center justify-center transition-all', isDragging ? 'bg-primary/10 text-primary scale-110' : 'bg-slate-100 dark:bg-[#2d3139] text-slate-400 group-hover:bg-primary/10 group-hover:text-primary']">
          <Upload class="w-7 h-7" />
        </div>
        <div>
          <p class="font-semibold text-slate-800 dark:text-white text-sm">
            Drop files here or <span class="text-primary">browse</span>
          </p>
          <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">
            Uploading to <span class="font-medium text-slate-600 dark:text-slate-300">{{ libraries.find(l => l.id === activeLibraryId)?.name || 'selected library' }}</span>
          </p>
          <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
            MP4, MOV, MKV, WebM &middot; MP3, AAC, WAV, OGG
          </p>
        </div>
      </div>
    </div>
    <div
      v-if="Object.keys(uploadStore.uploads).length"
      class="space-y-3 mt-6"
    >
      <div
        v-for="(upload, id) in uploadStore.uploads"
        :key="id"
        class="bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] rounded-xl p-4 flex items-center gap-4"
      >
        <div class="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
          <component
            :is="upload.status === 'completed' ? Check : Activity"
            :class="['w-5 h-5', { 'animate-pulse': upload.status === 'uploading' }]"
          />
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex justify-between items-baseline mb-1.5">
            <span class="font-medium text-slate-900 dark:text-white text-sm truncate pr-2">{{ upload.name }}</span>
            <span class="text-xs font-semibold text-primary">{{ upload.progress }}%</span>
          </div>
          <div class="w-full bg-slate-100 dark:bg-[#2d3139] h-1.5 rounded-full overflow-hidden">
            <div
              class="bg-primary h-full transition-all duration-300"
              :style="{ width: `${upload.progress}%` }"
            />
          </div>
          <div class="mt-1.5 text-[10px] uppercase tracking-wider font-semibold text-slate-500 flex items-center gap-1.5">
            <span v-if="upload.status === 'completed'" class="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            <span v-else-if="upload.status === 'uploading'" class="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
            <span v-else-if="upload.status === 'error'" class="w-1.5 h-1.5 rounded-full bg-rose-500" />
            {{ upload.status }}
          </div>
        </div>
      </div>
    </div>
    <div class="flex justify-end mt-6">
      <button
        v-if="Object.keys(uploadStore.uploads).length === 0"
        class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer"
        @click="$emit('close')"
      >
        Cancel
      </button>
      <button
        v-else
        class="px-5 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
        @click="$emit('close')"
      >
        Done
      </button>
    </div>
  </BaseModal>
</template>
