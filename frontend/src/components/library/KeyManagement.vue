<script setup>
import { ref } from 'vue'
import { createLibrary, updateLibrary, createLibraryKey, revokeLibraryKey, deleteLibraryKey, regenerateLibraryKey } from '../../api/videos'
import BaseModal from '../modals/BaseModal.vue'
import ConfirmationModal from '../ConfirmationModal.vue'
import { Key, Plus, Copy, Pencil, Eye, EyeOff, RefreshCw, Ban, Trash2, Lock, Library } from 'lucide-vue-next'

const props = defineProps({
  libraries: { type: Array, required: true },
  activeLibraryId: { type: String, required: true },
  libraryKeys: { type: Array, required: true },
  unmaskedKeys: { type: Object, default: () => ({}) }
})

const emit = defineEmits(['update:active-library-id', 'notify', 'fetch-keys', 'fetch-libraries'])

const showCreateKeyModal = ref(false)
const showCreateLibraryModal = ref(false)
const showEditLibraryModal = ref(false)
const newKeyName = ref('')
const newLibraryName = ref('')
const editingLibraryName = ref('')
const generatedKeySecret = ref('')

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

const formatDate = (dateStr) => {
  if (!dateStr) return 'N/A'
  return new Intl.DateTimeFormat('en-US', {
    month: 'short', day: 'numeric', year: 'numeric'
  }).format(new Date(dateStr))
}

const copyToClipboard = async (text, successMsg) => {
  try {
    await navigator.clipboard.writeText(text)
    emit('notify', { message: successMsg || 'Copied to clipboard.', type: 'success' })
  } catch {
    emit('notify', { message: 'Failed to copy to clipboard.', type: 'error' })
  }
}

const handleCreateLibrary = async () => {
  if (!newLibraryName.value.trim()) return
  try {
    const result = await createLibrary(newLibraryName.value.trim())
    const lib = result.library || result
    emit('notify', { message: "Library created. Default key generated.", type: "success" })
    newLibraryName.value = ''
    showCreateLibraryModal.value = false
    emit('fetch-libraries')
    if (result.default_key?.playback_secret) {
      generatedKeySecret.value = result.default_key.playback_secret
    }
  } catch {
    emit('notify', { message: "Failed to create library.", type: "error" })
  }
}

const triggerEditLibrary = () => {
  const lib = props.libraries.find(l => l.id === props.activeLibraryId)
  if (!lib) return
  editingLibraryName.value = lib.name
  showEditLibraryModal.value = true
}

const handleUpdateLibrary = async () => {
  if (!editingLibraryName.value.trim()) return
  try {
    await updateLibrary(props.activeLibraryId, editingLibraryName.value.trim())
    emit('notify', { message: "Library name updated successfully.", type: "success" })
    showEditLibraryModal.value = false
    emit('fetch-libraries')
  } catch {
    emit('notify', { message: "Failed to update library name.", type: "error" })
  }
}

const handleCreateKey = async () => {
  if (!newKeyName.value.trim() || !props.activeLibraryId) return
  try {
    const result = await createLibraryKey(props.activeLibraryId, newKeyName.value.trim())
    generatedKeySecret.value = result.playback_secret
    newKeyName.value = ''
    showCreateKeyModal.value = false
    emit('fetch-keys')
  } catch {
    emit('notify', { message: "Failed to create key.", type: "error" })
  }
}

const handleRevokeKey = (keyId) => {
  triggerConfirmation({
    title: 'Revoke Playback Key',
    description: 'Are you sure you want to revoke this key? Any external client using this key will immediately lose access to the library.',
    confirmText: 'Revoke Key',
    cancelText: 'Cancel',
    type: 'danger',
    onConfirm: async () => {
      try {
        await revokeLibraryKey(props.activeLibraryId, keyId)
        emit('notify', { message: "Key revoked successfully.", type: "success" })
        emit('fetch-keys')
      } catch {
        emit('notify', { message: "Failed to revoke key.", type: "error" })
      }
    }
  })
}

const handleDeleteKey = (keyId) => {
  triggerConfirmation({
    title: 'Delete Playback Key',
    description: 'Are you sure you want to permanently delete this key? This action is irreversible and any external client using this key will lose access.',
    confirmText: 'Delete Key',
    cancelText: 'Cancel',
    type: 'danger',
    onConfirm: async () => {
      try {
        await deleteLibraryKey(props.activeLibraryId, keyId)
        emit('notify', { message: "Key permanently deleted.", type: "success" })
        emit('fetch-keys')
      } catch {
        emit('notify', { message: "Failed to delete key.", type: "error" })
      }
    }
  })
}

const handleRegenerateKey = (keyId) => {
  triggerConfirmation({
    title: 'Regenerate Playback Key',
    description: 'Are you sure you want to regenerate this key? The existing key secret will be invalidated immediately, and a new one will be generated.',
    confirmText: 'Regenerate',
    cancelText: 'Cancel',
    type: 'warning',
    onConfirm: async () => {
      try {
        const result = await regenerateLibraryKey(props.activeLibraryId, keyId)
        generatedKeySecret.value = result.playback_secret
        emit('notify', { message: "Key regenerated successfully.", type: "success" })
        emit('fetch-keys')
      } catch {
        emit('notify', { message: "Failed to regenerate key.", type: "error" })
      }
    }
  })
}
</script>

<template>
  <div>
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
      <div>
        <h2 class="text-2xl font-bold text-slate-900 dark:text-white mb-1">
          Content Library
        </h2>
        <p class="text-slate-500 dark:text-slate-400 text-sm">
          Manage and monitor your VOD assets.
        </p>
      </div>

      <div class="flex items-center gap-3">
        <div class="flex items-center gap-1.5 relative min-w-[200px]">
          <select
            :value="activeLibraryId"
            class="w-full px-3 py-2 bg-white dark:bg-[#1a1d24] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors cursor-pointer"
            @change="$emit('update:active-library-id', $event.target.value)"
          >
            <option
              v-for="lib in libraries"
              :key="lib.id"
              :value="lib.id"
            >
              {{ lib.name }}
            </option>
          </select>

          <button
            v-if="activeLibraryId"
            class="p-2 border border-slate-200 dark:border-[#2d3139] bg-white dark:bg-[#1a1d24] hover:bg-slate-50 dark:hover:bg-[#2d3139] rounded-lg transition-colors cursor-pointer"
            aria-label="Rename Library"
            @click="showEditLibraryModal = true; editingLibraryName = libraries.find(l => l.id === activeLibraryId)?.name || ''"
          >
            <Pencil class="w-4 h-4 text-slate-500 dark:text-slate-400 hover:text-primary" />
          </button>
        </div>

        <button
          class="px-3 py-2 bg-slate-100 hover:bg-slate-200 dark:bg-[#2d3139] dark:hover:bg-[#3d424e] text-slate-700 dark:text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-1.5 cursor-pointer"
          @click="showCreateLibraryModal = true"
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
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Keys generated here let external clients stream and sign URLs for this library.
          </p>
        </div>
        <button
          class="px-3 py-1.5 bg-primary hover:bg-rose-600 text-white rounded-lg text-xs font-semibold transition-colors flex items-center gap-1 cursor-pointer"
          @click="showCreateKeyModal = true"
        >
          <Plus class="w-3.5 h-3.5" />
          New Key
        </button>
      </div>

      <div
        v-if="libraryKeys.length === 0"
        class="text-center py-6 text-sm text-slate-400 dark:text-slate-500"
      >
        No playback keys generated for this library.
      </div>
      <div
        v-else
        class="overflow-x-auto"
      >
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
            <tr
              v-for="k in libraryKeys"
              :key="k.id"
              class="border-b border-slate-100 dark:border-[#2d3139]/50 last:border-0 text-slate-700 dark:text-slate-300"
            >
              <td class="py-3 font-medium text-slate-900 dark:text-white">{{ k.key_name }}</td>
              <td class="py-3">
                <div class="flex items-center gap-2">
                  <span class="font-mono text-xs px-2.5 py-1 bg-slate-50 dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] rounded-lg select-all">
                    {{ unmaskedKeys[k.id] ? k.playback_secret : '••••••••••••••••••••••••••••••••' }}
                  </span>
                  <button
                    class="p-1 hover:bg-slate-100 dark:hover:bg-[#2d3139] rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
                    :aria-label="unmaskedKeys[k.id] ? 'Mask Secret' : 'Unmask Secret'"
                    @click="$emit('toggle-key-mask', k.id)"
                  >
                    <component :is="unmaskedKeys[k.id] ? EyeOff : Eye" class="w-4 h-4" />
                  </button>
                  <button
                    class="p-1 hover:bg-slate-100 dark:hover:bg-[#2d3139] rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors cursor-pointer"
                    aria-label="Copy Secret"
                    @click="copyToClipboard(k.playback_secret, 'Access Key secret copied.')"
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
                    class="p-1.5 hover:bg-amber-500/10 hover:text-amber-500 dark:hover:bg-amber-500/20 text-slate-400 dark:text-slate-500 rounded transition-colors cursor-pointer"
                    aria-label="Regenerate Key Secret"
                    @click="handleRegenerateKey(k.id)"
                  >
                    <RefreshCw class="w-4 h-4" />
                  </button>
                  <button
                    v-if="k.is_active"
                    class="p-1.5 hover:bg-rose-500/10 hover:text-rose-500 dark:hover:bg-rose-500/20 text-slate-400 dark:text-slate-500 rounded transition-colors cursor-pointer"
                    aria-label="Revoke (Deactivate) Key"
                    @click="handleRevokeKey(k.id)"
                  >
                    <Ban class="w-4 h-4" />
                  </button>
                  <button
                    class="p-1.5 hover:bg-rose-500/10 hover:text-rose-500 dark:hover:bg-rose-500/20 text-slate-400 dark:text-slate-500 rounded transition-colors cursor-pointer"
                    aria-label="Delete Key Completely"
                    @click="handleDeleteKey(k.id)"
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

    <BaseModal
      :is-open="showCreateLibraryModal"
      title="Create Library"
      description="Create a logical grouping for related media assets, each with custom credentials."
      @close="showCreateLibraryModal = false"
    >
      <template #icon>
        <Library class="w-5 h-5 text-primary" />
      </template>
      <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Library Name</label>
      <input
        v-model="newLibraryName"
        placeholder="e.g. Production Team"
        type="text"
        class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
      >
      <div class="flex gap-3 justify-end mt-6">
        <button
          class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer"
          @click="showCreateLibraryModal = false"
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
          @click="handleCreateLibrary"
        >
          Create
        </button>
      </div>
    </BaseModal>

    <BaseModal
      :is-open="showEditLibraryModal"
      title="Rename Library"
      description="Modify the name of this library group."
      @close="showEditLibraryModal = false"
    >
      <template #icon>
        <Library class="w-5 h-5 text-primary" />
      </template>
      <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Library Name</label>
      <input
        v-model="editingLibraryName"
        placeholder="e.g. Production Team"
        type="text"
        class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
      >
      <div class="flex gap-3 justify-end mt-6">
        <button
          class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer"
          @click="showEditLibraryModal = false"
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
          @click="handleUpdateLibrary"
        >
          Save
        </button>
      </div>
    </BaseModal>

    <BaseModal
      :is-open="showCreateKeyModal"
      title="Generate Access Key"
      description="Create a playback signing token for this library."
      @close="showCreateKeyModal = false"
    >
      <template #icon>
        <Key class="w-5 h-5 text-primary" />
      </template>
      <label class="block text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5">Key Name</label>
      <input
        v-model="newKeyName"
        placeholder="e.g. Website Key"
        type="text"
        class="w-full px-3 py-2.5 bg-white dark:bg-[#111318] border border-slate-200 dark:border-[#2d3139] text-slate-900 dark:text-white rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-colors"
      >
      <div class="flex gap-3 justify-end mt-6">
        <button
          class="px-4 py-2 border border-slate-200 dark:border-[#2d3139] text-slate-600 dark:text-slate-300 rounded-xl hover:bg-slate-50 dark:hover:bg-[#2d3139] text-sm font-medium transition-colors cursor-pointer"
          @click="showCreateKeyModal = false"
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
          @click="handleCreateKey"
        >
          Generate
        </button>
      </div>
    </BaseModal>

    <div
      v-if="generatedKeySecret"
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-900/15 backdrop-blur-[2px]"
    >
      <div class="bg-white dark:bg-[#1a1d24] rounded-2xl p-6 w-full max-w-md border border-slate-200 dark:border-[#2d3139] shadow-2xl">
        <h3 class="text-lg font-bold text-emerald-600 dark:text-emerald-400 flex items-center gap-2 mb-2">
          <Lock class="w-5 h-5" />
          Access Key Generated
        </h3>
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-6">
          Copy this key secret now. For security, it will not be shown again.
        </p>
        <div class="flex items-center gap-2 bg-slate-50 dark:bg-[#111318] p-3 rounded-xl border border-slate-200 dark:border-[#2d3139] font-mono text-sm break-all relative">
          <span class="flex-1 select-all pr-10 text-slate-800 dark:text-slate-200">{{ generatedKeySecret }}</span>
          <button
            class="absolute right-3 p-1.5 hover:bg-slate-200 dark:hover:bg-[#2d3139] rounded transition-colors text-slate-500 cursor-pointer"
            aria-label="Copy Secret"
            @click="copyToClipboard(generatedKeySecret, 'Secret copied to clipboard.')"
          >
            <Copy class="w-4 h-4" />
          </button>
        </div>
        <div class="flex justify-end mt-6">
          <button
            class="px-5 py-2 bg-primary text-white rounded-xl hover:bg-rose-600 text-sm font-medium transition-colors cursor-pointer"
            @click="generatedKeySecret = ''"
          >
            Done
          </button>
        </div>
      </div>
    </div>

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
  </div>
</template>
