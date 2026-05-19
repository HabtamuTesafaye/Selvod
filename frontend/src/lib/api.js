import axios from 'axios'
import { getAdminKey, getPlaybackKey } from './credentials'

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
