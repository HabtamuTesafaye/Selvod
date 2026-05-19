import { api, healthApi } from '../lib/api'

export async function listVideos() {
  const { data } = await api.get('/videos')
  return data
}

export async function uploadVideo(formData, config = {}) {
  const { data } = await api.post('/videos', formData, config)
  return data
}

export async function deleteVideo(id) {
  await api.delete(`/videos/${id}`)
}

export async function getStream(id) {
  const { data } = await api.get(`/videos/${id}/stream`)
  return data
}

export async function getHealth() {
  const { data } = await healthApi.get('/health')
  return data
}
