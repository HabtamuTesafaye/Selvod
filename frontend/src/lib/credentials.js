const ADMIN_KEY = 'SV_API_KEY'
const PLAYBACK_KEY = 'SV_PLAYBACK_KEY'

export function getAdminKey() {
  return sessionStorage.getItem(ADMIN_KEY) || localStorage.getItem(ADMIN_KEY) || ''
}

export function getPlaybackKey() {
  return sessionStorage.getItem(PLAYBACK_KEY) || localStorage.getItem(PLAYBACK_KEY) || ''
}

export function saveCredentials({ adminKey, playbackKey, rememberMe }) {
  if (rememberMe) {
    localStorage.setItem(ADMIN_KEY, adminKey)
    localStorage.setItem(PLAYBACK_KEY, playbackKey)
    sessionStorage.removeItem(ADMIN_KEY)
    sessionStorage.removeItem(PLAYBACK_KEY)
  } else {
    sessionStorage.setItem(ADMIN_KEY, adminKey)
    sessionStorage.setItem(PLAYBACK_KEY, playbackKey)
    localStorage.removeItem(ADMIN_KEY)
    localStorage.removeItem(PLAYBACK_KEY)
  }
}

export function clearCredentials() {
  localStorage.removeItem(ADMIN_KEY)
  localStorage.removeItem(PLAYBACK_KEY)
  sessionStorage.removeItem(ADMIN_KEY)
  sessionStorage.removeItem(PLAYBACK_KEY)
}
