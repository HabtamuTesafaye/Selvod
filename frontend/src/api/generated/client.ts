import { api, healthApi } from '../../lib/api'
import type { components, paths } from './types'

type Video = components['schemas']['Video']
type SignedURL = components['schemas']['SignedURL']
type Library = components['schemas']['Library']
type LibraryKey = components['schemas']['LibraryKey']
type Health = components['schemas']['Health']

function pick<T, K extends keyof T>(obj: T, ...keys: K[]): Pick<T, K> {
  const result = {} as Pick<T, K>
  for (const key of keys) {
    if (obj[key] !== undefined) result[key] = obj[key]
  }
  return result
}

export async function listVideos(libraryId?: string, { signal }: { signal?: AbortSignal } = {}): Promise<{ videos: Video[]; total: number }> {
  const params = libraryId ? { library_id: libraryId } : undefined
  const { data } = await api.get('/videos', { params, signal })
  return data
}

export async function uploadVideo(formData: FormData, config: Record<string, unknown> = {}): Promise<Video> {
  const { data } = await api.post('/videos', formData, config)
  return data as Video
}

export async function getVideo(id: string): Promise<Video> {
  const { data } = await api.get(`/videos/${id}`)
  return data
}

export async function updateVideo(id: string, body: { title?: string; library_id?: string }): Promise<Video> {
  const { data } = await api.patch(`/videos/${id}`, pick(body, 'title', 'library_id'))
  return data
}

export async function deleteVideo(id: string): Promise<void> {
  await api.delete(`/videos/${id}`)
}

export async function getStream(id: string, { signal }: { signal?: AbortSignal } = {}): Promise<SignedURL> {
  const { data } = await api.get(`/videos/${id}/stream`, { signal })
  return data
}

export async function getEmbed(id: string, { signal }: { signal?: AbortSignal } = {}): Promise<{ url: string }> {
  const { data } = await api.get(`/videos/${id}/embed`, { signal })
  return data
}

export async function getHealth({ signal }: { signal?: AbortSignal } = {}): Promise<Health> {
  const { data } = await healthApi.get('/health', { signal })
  return data
}

export async function listLibraries({ signal }: { signal?: AbortSignal } = {}): Promise<Library[]> {
  const { data } = await api.get('/libraries', { signal })
  return data
}

export async function createLibrary(name: string): Promise<{ library: Library; default_key: LibraryKey | null }> {
  const { data } = await api.post('/libraries', { name })
  return data
}

export async function updateLibrary(id: string, name: string): Promise<Library> {
  const { data } = await api.patch(`/libraries/${id}`, { name })
  return data
}

export async function listLibraryKeys(libraryId: string, { signal }: { signal?: AbortSignal } = {}): Promise<LibraryKey[]> {
  const { data } = await api.get(`/libraries/${libraryId}/keys`, { signal })
  return data
}

export async function createLibraryKey(libraryId: string, keyName: string): Promise<LibraryKey> {
  const { data } = await api.post(`/libraries/${libraryId}/keys`, { key_name: keyName })
  return data
}

export async function revokeLibraryKey(libraryId: string, keyId: string): Promise<void> {
  await api.post(`/libraries/${libraryId}/keys/${keyId}/revoke`)
}

export async function deleteLibraryKey(libraryId: string, keyId: string): Promise<void> {
  await api.delete(`/libraries/${libraryId}/keys/${keyId}`)
}

export async function regenerateLibraryKey(libraryId: string, keyId: string): Promise<{ playback_secret: string }> {
  const { data } = await api.post(`/libraries/${libraryId}/keys/${keyId}/regenerate`)
  return data
}

export async function getGlobalPlayerConfig({ signal }: { signal?: AbortSignal } = {}): Promise<Record<string, unknown>> {
  const { data } = await api.get('/config/player', { signal })
  return data
}

export async function updateGlobalPlayerConfig(config: Record<string, unknown>): Promise<Record<string, unknown>> {
  const { data } = await api.post('/config/player', config)
  return data
}
