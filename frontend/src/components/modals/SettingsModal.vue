<script setup>
import { ref, watch } from 'vue'
import { Settings } from 'lucide-vue-next'
import BaseModal from './BaseModal.vue'

const props = defineProps({
  isOpen: Boolean,
  apiKey: String,
  playbackKey: String
})

const emit = defineEmits(['close', 'save'])

const localApiKey = ref('')
const localPlaybackKey = ref('')

watch(() => props.isOpen, (open) => {
  if (open) {
    localApiKey.value = props.apiKey
    localPlaybackKey.value = props.playbackKey
  }
})

const themeColors = [
  { name: 'Rose', hex: '#f43f5e' },
  { name: 'Blue', hex: '#3b82f6' },
  { name: 'Indigo', hex: '#6366f1' },
  { name: 'Emerald', hex: '#10b981' },
  { name: 'Amber', hex: '#f59e0b' }
]

const changePrimaryColor = (colorHex) => {
  document.documentElement.style.setProperty('--color-primary', colorHex)
  localStorage.setItem('SV_PRIMARY_COLOR', colorHex)
}
</script>

<template>
  <BaseModal :is-open="isOpen" title="System & Credentials" description="Configure the primary dashboard theme and authentication keys used to query the secure API endpoints." max-width="max-w-md" @close="$emit('close')">
    <div class="space-y-4">
      <div>
        <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-2">Accent Color</label>
        <div class="flex gap-3">
          <button
            v-for="c in themeColors"
            :key="c.name"
            :title="c.name"
            @click="changePrimaryColor(c.hex)"
            :style="{ backgroundColor: c.hex }"
            class="w-7 h-7 rounded-full shadow-sm hover:scale-110 transition-transform cursor-pointer border-2 border-white dark:border-[#1a1d24] focus:outline-none"
          />
        </div>
      </div>

      <hr class="border-slate-100 dark:border-[#2d3139] my-4" />

      <div>
        <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Admin API Key</label>
        <input v-model="localApiKey" type="password" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
      </div>

      <div>
        <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Playback Scope Key</label>
        <input v-model="localPlaybackKey" type="password" class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors" />
      </div>
    </div>

    <div class="flex gap-3 justify-end mt-8">
      <button @click="$emit('close')" class="px-5 py-2.5 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer">Cancel</button>
      <button @click="$emit('save', { apiKey: localApiKey, playbackKey: localPlaybackKey })" class="px-5 py-2.5 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer">Save Keys</button>
    </div>
  </BaseModal>
</template>
