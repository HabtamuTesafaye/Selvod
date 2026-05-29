const ADMIN_KEY = 'SV_API_KEY'
const PLAYBACK_KEY = 'SV_PLAYBACK_KEY'

export function getAdminKey() {
  return sessionStorage.getItem(ADMIN_KEY) || ''
}

export function getPlaybackKey() {
  return sessionStorage.getItem(PLAYBACK_KEY) || ''
}

// Security: Always use sessionStorage only — keys clear on tab close.
// The rememberMe parameter is accepted for API compatibility but has no effect.
export function saveCredentials({ adminKey, playbackKey, rememberMe }) {
  sessionStorage.setItem(ADMIN_KEY, adminKey)
  sessionStorage.setItem(PLAYBACK_KEY, playbackKey)
  // Migrate: clear any legacy localStorage entries
  localStorage.removeItem(ADMIN_KEY)
  localStorage.removeItem(PLAYBACK_KEY)
}

export function clearCredentials() {
  sessionStorage.removeItem(ADMIN_KEY)
  sessionStorage.removeItem(PLAYBACK_KEY)
  localStorage.removeItem(ADMIN_KEY)
  localStorage.removeItem(PLAYBACK_KEY)
}
