import axios from 'axios'
import { getAdminKey, getPlaybackKey } from './credentials'

const controllers = new Map()

export function cancelRequest(key) {
  const ctrl = controllers.get(key)
  if (ctrl) {
    ctrl.abort()
    controllers.delete(key)
  }
}

export const api = axios.create({
  baseURL: '/api/v1',
})

export const healthApi = axios.create({
  baseURL: '/',
})

api.interceptors.request.use((config) => {
  const isStreamSigning = config.url?.endsWith('/stream')
  const token = isStreamSigning ? getPlaybackKey() : getAdminKey()
  config.headers.Authorization = `Bearer ${token}`
  if (config.signal) {
    const controller = new AbortController()
    config.signal.addEventListener('abort', () => controller.abort())
    config.signal = controller.signal
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && (error.response.status === 401 || error.response.status === 403)) {
      window.dispatchEvent(new CustomEvent('unauthorized'))
    }
    return Promise.reject(error)
  }
)
