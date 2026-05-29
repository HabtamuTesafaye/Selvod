import { api, healthApi } from '../lib/api'

export async function listVideos(libraryId, { signal } = {}) {
  const url = libraryId ? `/videos?library_id=${libraryId}` : '/videos'
  const { data } = await api.get(url, { signal })
  return data
}

export async function uploadVideo(formData, config = {}) {
  const { data } = await api.post('/videos', formData, config)
  return data
}

export async function deleteVideo(id) {
  await api.delete(`/videos/${id}`)
}

export async function getStream(id, { signal } = {}) {
  const { data } = await api.get(`/videos/${id}/stream`, { signal })
  return data
}

export async function getEmbed(id, { signal } = {}) {
  const { data } = await api.get(`/videos/${id}/embed`, { signal })
  return data
}

export async function getHealth({ signal } = {}) {
  const { data } = await healthApi.get('/health', { signal })
  return data
}

// Library Management
export async function listLibraries({ signal } = {}) {
  const { data } = await api.get('/libraries', { signal })
  return data
}

export async function createLibrary(name) {
  const { data } = await api.post('/libraries', { name })
  return data
}

export async function listLibraryKeys(libraryId, { signal } = {}) {
  const { data } = await api.get(`/libraries/${libraryId}/keys`, { signal })
  return data
}

export async function createLibraryKey(libraryId, keyName) {
  const { data } = await api.post(`/libraries/${libraryId}/keys`, { key_name: keyName })
  return data
}

export async function revokeLibraryKey(libraryId, keyId) {
  await api.post(`/libraries/${libraryId}/keys/${keyId}/revoke`)
}

export async function deleteLibraryKey(libraryId, keyId) {
  await api.delete(`/libraries/${libraryId}/keys/${keyId}`)
}

export async function regenerateLibraryKey(libraryId, keyId) {
  const { data } = await api.post(`/libraries/${libraryId}/keys/${keyId}/regenerate`)
  return data
}

export async function updateVideo(id, { title, library_id }) {
  const { data } = await api.patch(`/videos/${id}`, { title, library_id })
  return data
}

export async function updateLibrary(id, name) {
  const { data } = await api.patch(`/libraries/${id}`, { name })
  return data
}

export async function getGlobalPlayerConfig({ signal } = {}) {
  const { data } = await api.get('/config/player', { signal })
  return data
}

export async function updateGlobalPlayerConfig(config) {
  const { data } = await api.post('/config/player', config)
  return data
}
