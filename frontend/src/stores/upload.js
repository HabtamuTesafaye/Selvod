import { defineStore } from 'pinia'
import { uploadVideo } from '../api/videos'

export const useUploadStore = defineStore('upload', {
  state: () => ({
    uploads: {}, // id -> { file, progress, status, error, video }
  }),

  actions: {
    async uploadFile(file, title, libraryId) {
      const tempId = Math.random().toString(36).substring(7)
      this.uploads[tempId] = {
        name: title || file.name,
        progress: 0,
        status: 'uploading',
        error: null
      }

      const formData = new FormData()
      formData.append('file', file)
      if (title) formData.append('title', title)
      if (libraryId) formData.append('library_id', libraryId)

      try {
        const video = await uploadVideo(formData, {
          onUploadProgress: (progressEvent) => {
            const total = progressEvent.total || file.size || 1
            const percentCompleted = Math.round((progressEvent.loaded * 100) / total)
            this.uploads[tempId].progress = percentCompleted
          }
        })

        this.uploads[tempId].status = 'processing'
        this.uploads[tempId].video = video
        this.uploads[tempId].progress = 100
        
        return video
      } catch (err) {
        this.uploads[tempId].status = 'failed'
        this.uploads[tempId].error = err.response?.data || err.message
        throw err
      }
    },

    removeUpload(id) {
      delete this.uploads[id]
    }
  }
})
