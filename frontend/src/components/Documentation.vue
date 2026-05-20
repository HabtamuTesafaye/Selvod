<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Key, Lock } from 'lucide-vue-next'

const props = defineProps({
  apiKey: String,
  playbackKey: String,
  libraryId: String
})

const activeSection = ref('overview')
const backendTab = ref('node')
const frontendTab = ref('vue')

const sections = [
  { id: 'overview', title: 'Overview' },
  { id: 'auth', title: 'Authentication' },
  { id: 'embed', title: 'Embed Player (Lightweight)' },
  { id: 'backend', title: 'Server-Side Signing' },
  { id: 'frontend', title: 'Frontend Integration' },
  { id: 'api', title: 'REST API Reference' }
]

const backendTabs = [
  { id: 'node', label: 'Node.js' },
  { id: 'python', label: 'Python' },
  { id: 'go', label: 'Go' },
]

const frontendTabs = [
  { id: 'vue', label: 'Vue 3' },
  { id: 'nuxt', label: 'Nuxt 3' },
  { id: 'next', label: 'Next.js' },
  { id: 'react', label: 'React' },
]

const contentRef = ref(null)
let observer = null

const scrollTo = (id) => {
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

onMounted(() => {
  observer = new IntersectionObserver((entries) => {
    // find entry closest to top with the highest intersectionRatio
    const visible = entries.filter(e => e.isIntersecting)
    if (visible.length) {
      // pick the one highest on screen
      const top = visible.reduce((prev, curr) =>
        curr.boundingClientRect.top < prev.boundingClientRect.top ? curr : prev
      )
      activeSection.value = top.target.id
    }
  }, {
    root: contentRef.value,
    threshold: [0, 0.25, 0.5],
    rootMargin: '-10% 0px -50% 0px'
  })

  sections.forEach(s => {
    const el = document.getElementById(s.id)
    if (el) observer.observe(el)
  })
})

onUnmounted(() => {
  if (observer) observer.disconnect()
})

// Code snippets
const backendCode = {
  node: `const crypto = require('crypto');

// Secret shared with Nginx (Your Playback Key)
const SECRET = '${props.playbackKey || "YOUR_PLAYBACK_KEY"}';
const LIBRARY_ID = '${props.libraryId || "YOUR_LIBRARY_ID"}';

function generateSignedUrl(videoId, ttlSeconds = 3600) {
  const expires = Math.floor(Date.now() / 1000) + ttlSeconds;

  // Format: "SECRET/videoId/expires"
  const input = \`\${SECRET}/\${videoId}/\${expires}\`;

  // MD5 → Base64URL (no padding, + → -, / → _)
  const token = crypto.createHash('md5')
    .update(input).digest('base64')
    .replace(/=/g, '').replace(/\\+/g, '-').replace(/\\//g, '_');

  return \`https://vod.yourdomain.com/hls/\${LIBRARY_ID}/\${videoId}/master.m3u8?token=\${token}&expires=\${expires}\`;
}`,
  python: `import hashlib, base64, time

SECRET = "${props.playbackKey || 'YOUR_PLAYBACK_KEY'}"
LIBRARY_ID = "${props.libraryId || 'YOUR_LIBRARY_ID'}"

def generate_signed_url(video_id: str, ttl_seconds: int = 3600) -> str:
    expires = int(time.time()) + ttl_seconds

    # Format: "SECRET/videoId/expires"
    raw = f"{SECRET}/{video_id}/{expires}"

    # MD5 → Base64URL
    digest = hashlib.md5(raw.encode()).digest()
    token = base64.b64encode(digest).decode() \\
        .rstrip('=').replace('+', '-').replace('/', '_')

    return (
        f"https://vod.yourdomain.com/hls/{LIBRARY_ID}/{video_id}"
        f"/master.m3u8?token={token}&expires={expires}"
    )`,
  go: `package main

import (
    "crypto/md5"
    "encoding/base64"
    "fmt"
    "strings"
    "time"
)

const secret = "${props.playbackKey || 'YOUR_PLAYBACK_KEY'}"
const libraryID = "${props.libraryId || 'YOUR_LIBRARY_ID'}"

func generateSignedURL(videoID string, ttl time.Duration) string {
    expires := time.Now().Add(ttl).Unix()

    // Format: "SECRET/videoId/expires"
    raw := fmt.Sprintf("%s/%s/%d", secret, videoID, expires)

    // MD5 → Base64URL
    sum := md5.Sum([]byte(raw))
    token := base64.StdEncoding.EncodeToString(sum[:])
    token = strings.TrimRight(token, "=")
    token = strings.ReplaceAll(token, "+", "-")
    token = strings.ReplaceAll(token, "/", "_")

    return fmt.Sprintf(
        "https://vod.yourdomain.com/hls/%s/%s/master.m3u8?token=%s&expires=%d",
        libraryID, videoID, token, expires,
    )
}`,
}

const frontendCode = {
  vue: `<!-- VideoPlayer.vue -->
<template>
  <video ref="videoRef" controls class="w-full rounded-xl bg-black"></video>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import Hls from 'hls.js'

const props = defineProps({ signedUrl: String })
const videoRef = ref(null)
let hls = null

onMounted(() => {
  if (Hls.isSupported()) {
    hls = new Hls({ withCredentials: true })
    hls.loadSource(props.signedUrl)
    hls.attachMedia(videoRef.value)
  } else if (videoRef.value.canPlayType('application/vnd.apple.mpegurl')) {
    videoRef.value.src = props.signedUrl // Safari native HLS
  }
})

onUnmounted(() => hls?.destroy())
<\/script>`,

  nuxt: `<!-- components/VideoPlayer.vue -->
<template>
  <video ref="videoRef" controls class="w-full rounded-xl bg-black"></video>
</template>

<script setup>
import Hls from 'hls.js'

const props = defineProps({ signedUrl: String })
const videoRef = ref(null)
let hls = null

// useNuxtApp() ensures this runs only client-side
onMounted(() => {
  if (!process.client) return
  if (Hls.isSupported()) {
    hls = new Hls({ withCredentials: true })
    hls.loadSource(props.signedUrl)
    hls.attachMedia(videoRef.value)
  } else if (videoRef.value?.canPlayType('application/vnd.apple.mpegurl')) {
    videoRef.value.src = props.signedUrl
  }
})

onUnmounted(() => hls?.destroy())
<\/script>`,

  next: `// components/VideoPlayer.tsx
'use client'
import { useEffect, useRef } from 'react'
import Hls from 'hls.js'

export default function VideoPlayer({ signedUrl }: { signedUrl: string }) {
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    let hls: Hls | null = null

    if (Hls.isSupported()) {
      hls = new Hls({ withCredentials: true })
      hls.loadSource(signedUrl)
      hls.attachMedia(video)
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = signedUrl // Safari
    }

    return () => hls?.destroy()
  }, [signedUrl])

  return <video ref={videoRef} controls className="w-full rounded-xl bg-black" />
}`,

  react: `// VideoPlayer.jsx
import { useEffect, useRef } from 'react'
import Hls from 'hls.js'

export default function VideoPlayer({ signedUrl }) {
  const videoRef = useRef(null)

  useEffect(() => {
    const video = videoRef.current
    if (!video) return
    let hls = null

    if (Hls.isSupported()) {
      hls = new Hls({ withCredentials: true })
      hls.loadSource(signedUrl)
      hls.attachMedia(video)
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = signedUrl // Safari native HLS
    }

    return () => hls?.destroy()
  }, [signedUrl])

  return <video ref={videoRef} controls className="w-full rounded-xl bg-black" />
}`,
}
</script>

<template>
  <div class="flex h-full bg-white dark:bg-gray-900 overflow-hidden rounded-xl border border-gray-200 dark:border-gray-800">

    <!-- Sticky Doc Sidebar with scroll-spy -->
    <div class="w-60 shrink-0 border-r border-gray-200 dark:border-gray-800 flex flex-col h-full hidden md:flex">
      <div class="px-6 pt-6 pb-4 border-b border-gray-100 dark:border-gray-800">
        <h3 class="text-xs font-bold text-gray-400 uppercase tracking-wider">Documentation</h3>
      </div>
      <nav class="flex-1 overflow-y-auto px-3 py-4 space-y-0.5">
        <button
          v-for="section in sections"
          :key="section.id"
          @click="scrollTo(section.id)"
          class="w-full text-left px-3 py-2 rounded-lg transition-all text-sm font-medium flex items-center gap-2"
          :class="activeSection === section.id
            ? 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400 font-semibold'
            : 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800'"
        >
          <span
            class="w-1.5 h-1.5 rounded-full shrink-0 transition-all"
            :class="activeSection === section.id ? 'bg-indigo-500' : 'bg-transparent'"
          ></span>
          {{ section.title }}
        </button>
      </nav>
    </div>

    <!-- Doc Content — independently scrollable -->
    <div ref="contentRef" class="flex-1 overflow-y-auto p-8 lg:p-12 scroll-smooth">
      <div class="max-w-3xl mx-auto space-y-16">

        <!-- Overview -->
        <section id="overview" class="space-y-6 scroll-mt-8">
          <div>
            <h1 class="text-3xl font-extrabold text-gray-900 dark:text-white mb-4">Developer Integration Guide</h1>
            <p class="text-lg text-gray-600 dark:text-gray-400 leading-relaxed">
              Learn how to authenticate with the API, securely generate signed HLS playback URLs on your backend, and embed the player in any frontend.
            </p>
          </div>
          <div class="bg-indigo-50 dark:bg-indigo-900/20 rounded-xl p-6 border border-indigo-100 dark:border-indigo-800/50">
            <h4 class="font-semibold text-indigo-900 dark:text-indigo-300 mb-2">Zero-Trust Architecture</h4>
            <p class="text-indigo-800 dark:text-indigo-400/80 text-sm">
              Selvod uses a stateless HMAC-MD5 secure link system. Video URLs must be dynamically signed by your backend using your Playback Key — they're never exposed to the browser.
            </p>
          </div>
        </section>

        <!-- Authentication -->
        <section id="auth" class="space-y-6 scroll-mt-8">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white border-b border-gray-200 dark:border-gray-800 pb-2">Authentication</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
              <div class="flex items-center gap-3 mb-4">
                <div class="p-2 bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400 rounded-lg">
                  <Key class="w-5 h-5" />
                </div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Admin API Key</h3>
              </div>
              <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">Used for CRUD operations on videos and libraries. Keep strictly on your server.</p>
              <code class="block bg-gray-900 text-gray-200 p-3 rounded-lg text-xs break-all border border-gray-700 font-mono">{{ apiKey || 'Not configured' }}</code>
            </div>

            <div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
              <div class="flex items-center gap-3 mb-4">
                <div class="p-2 bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 rounded-lg">
                  <Lock class="w-5 h-5" />
                </div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Playback Secret</h3>
              </div>
              <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">Used by your backend to sign video URLs. Never expose this to the frontend.</p>
            </div>
          </div>
        </section>

        <!-- Standalone Embed Player -->
        <section id="embed" class="space-y-6 scroll-mt-8">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white border-b border-gray-200 dark:border-gray-800 pb-2">Embed Player (Lightweight & Standalone)</h2>
          <p class="text-gray-600 dark:text-gray-400">
            Selvod includes a high-performance standalone player that allows clients to play videos with just their <strong>Library Playback Key</strong> and <strong>Video ID</strong>.
          </p>
          <div class="bg-gray-50 dark:bg-gray-850 rounded-xl p-6 border border-gray-250 dark:border-gray-800 space-y-4">
            <h4 class="font-bold text-gray-800 dark:text-gray-200">How to Embed with Iframe:</h4>
            <p class="text-sm text-gray-600 dark:text-gray-400">
              Inject the player straight into your website or CMS as an iframe. All token authentication, playback validation, HLS stream buffering, and token renewals are handled automatically.
            </p>
            <pre class="bg-gray-900 text-emerald-400 p-4 rounded-xl text-xs overflow-x-auto border border-gray-800 font-mono"><code>&lt;iframe
  src="/embed.html?videoId=YOUR_VIDEO_ID&amp;key=YOUR_LIBRARY_PLAYBACK_KEY"
  width="100%"
  height="100%"
  frameborder="0"
  allow="autoplay; fullscreen"
  allowfullscreen
&gt;&lt;/iframe&gt;</code></pre>
            <div class="p-4 bg-indigo-50 dark:bg-indigo-900/10 rounded-lg border border-indigo-100 dark:border-indigo-800/40 text-xs text-indigo-700 dark:text-indigo-300 space-y-2">
              <strong class="block text-indigo-900 dark:text-indigo-200">🔒 Library Scoping Isolation Guarantees:</strong>
              <p>Clients can only access videos in libraries where they have an active key. A key scoped to Library A will return <code class="bg-indigo-100 dark:bg-indigo-900/45 px-1.5 py-0.5 rounded font-bold">403 Forbidden</code> if attempting to play a video from Library B, even if the video ID is valid.</p>
            </div>
          </div>
        </section>

        <!-- Server-Side Signing -->
        <section id="backend" class="space-y-6 scroll-mt-8">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white border-b border-gray-200 dark:border-gray-800 pb-2">Server-Side Signing</h2>
          <p class="text-gray-600 dark:text-gray-400">
            When a user requests playback, your backend generates a signed URL and returns it to the frontend. Choose your backend language:
          </p>

          <!-- Backend language tabs -->
          <div class="bg-[#1e1e1e] rounded-xl overflow-hidden border border-gray-800 shadow-xl">
            <div class="flex items-center gap-0 border-b border-gray-800 bg-[#252526] overflow-x-auto">
              <div class="flex gap-2 px-4 py-3 shrink-0">
                <div class="w-3 h-3 rounded-full bg-red-500"></div>
                <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
                <div class="w-3 h-3 rounded-full bg-green-500"></div>
              </div>
              <div class="flex border-l border-gray-800 overflow-x-auto">
                <button
                  v-for="tab in backendTabs"
                  :key="tab.id"
                  @click="backendTab = tab.id"
                  class="px-4 py-3 text-xs font-medium transition-colors whitespace-nowrap border-r border-gray-800"
                  :class="backendTab === tab.id
                    ? 'text-white bg-[#1e1e1e] border-t-2 border-t-indigo-500 -mt-px'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-[#2d2d2d]'"
                >
                  {{ tab.label }}
                </button>
              </div>
            </div>
            <pre class="p-6 text-sm font-mono text-gray-300 overflow-x-auto leading-relaxed max-h-[420px] overflow-y-auto"><code>{{ backendCode[backendTab] }}</code></pre>
          </div>
        </section>

        <!-- Frontend Integration -->
        <section id="frontend" class="space-y-6 scroll-mt-8">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white border-b border-gray-200 dark:border-gray-800 pb-2">Frontend Integration</h2>
          <p class="text-gray-600 dark:text-gray-400">
            Once your backend returns the signed URL, pass it to any HLS-capable player. Here are ready-to-use components for popular frameworks:
          </p>

          <!-- Frontend language tabs -->
          <div class="bg-[#1e1e1e] rounded-xl overflow-hidden border border-gray-800 shadow-xl">
            <div class="flex items-center gap-0 border-b border-gray-800 bg-[#252526] overflow-x-auto">
              <div class="flex gap-2 px-4 py-3 shrink-0">
                <div class="w-3 h-3 rounded-full bg-red-500"></div>
                <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
                <div class="w-3 h-3 rounded-full bg-green-500"></div>
              </div>
              <div class="flex border-l border-gray-800 overflow-x-auto">
                <button
                  v-for="tab in frontendTabs"
                  :key="tab.id"
                  @click="frontendTab = tab.id"
                  class="px-4 py-3 text-xs font-medium transition-colors whitespace-nowrap border-r border-gray-800"
                  :class="frontendTab === tab.id
                    ? 'text-white bg-[#1e1e1e] border-t-2 border-t-indigo-500 -mt-px'
                    : 'text-gray-400 hover:text-gray-200 hover:bg-[#2d2d2d]'"
                >
                  {{ tab.label }}
                </button>
              </div>
            </div>
            <pre class="p-6 text-sm font-mono text-gray-300 overflow-x-auto leading-relaxed max-h-[480px] overflow-y-auto"><code>{{ frontendCode[frontendTab] }}</code></pre>
          </div>

          <div class="bg-amber-50 dark:bg-amber-900/10 rounded-xl p-4 border border-amber-200 dark:border-amber-800/40 text-sm text-amber-800 dark:text-amber-300">
            <strong>Install HLS.js:</strong>
            <code class="ml-2 bg-amber-100 dark:bg-amber-900/30 px-2 py-0.5 rounded font-mono text-xs">npm install hls.js</code>
          </div>
        </section>

        <!-- API Reference -->
        <section id="api" class="space-y-6 scroll-mt-8 pb-16">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white border-b border-gray-200 dark:border-gray-800 pb-2">REST API Reference</h2>
          <div class="space-y-3">

            <div v-for="endpoint in [
              { method: 'GET', color: 'blue', path: '/api/v1/videos', desc: 'Returns a paginated list of all videos in the active library. Requires Admin key.' },
              { method: 'POST', color: 'green', path: '/api/v1/videos', desc: 'Upload a new video. Expects multipart/form-data with a file field and optional title and library_id.' },
              { method: 'GET', color: 'blue', path: '/api/v1/videos/:id/stream', desc: 'Returns a signed HLS manifest URL. Accepts both Admin and Playback keys.' },
              { method: 'PATCH', color: 'yellow', path: '/api/v1/videos/:id', desc: 'Update video title or move it to another library. Requires Admin key.' },
              { method: 'DELETE', color: 'red', path: '/api/v1/videos/:id', desc: 'Deletes a video and removes all transcoded HLS segments from storage.' },
              { method: 'GET', color: 'blue', path: '/api/v1/libraries', desc: 'Lists all libraries. Requires Admin key.' },
              { method: 'POST', color: 'green', path: '/api/v1/libraries', desc: 'Create a new library. Body: { name: string }.' },
              { method: 'GET', color: 'blue', path: '/api/v1/libraries/:id/keys', desc: 'List all playback keys for a library.' },
              { method: 'POST', color: 'green', path: '/api/v1/libraries/:id/keys', desc: 'Generate a new access key for this library. Secret is returned once.' },
              { method: 'POST', color: 'green', path: '/api/v1/libraries/:id/keys/:key_id/regenerate', desc: 'Rotate the secret for an existing key. Old secret is immediately invalidated.' },
              { method: 'POST', color: 'yellow', path: '/api/v1/libraries/:id/keys/:key_id/revoke', desc: 'Deactivate a key without deleting it.' },
              { method: 'DELETE', color: 'red', path: '/api/v1/libraries/:id/keys/:key_id', desc: 'Permanently delete a playback key.' },
            ]" :key="endpoint.path + endpoint.method" class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden">
              <div class="bg-gray-50 dark:bg-gray-800/50 p-3 border-b border-gray-200 dark:border-gray-800 flex items-center gap-3 flex-wrap">
                <span :class="[
                  'px-2 py-0.5 rounded text-xs font-bold font-mono',
                  endpoint.color === 'blue' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-400' :
                  endpoint.color === 'green' ? 'bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-400' :
                  endpoint.color === 'yellow' ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/50 dark:text-amber-400' :
                  'bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-400'
                ]">{{ endpoint.method }}</span>
                <code class="text-gray-900 dark:text-gray-200 font-mono text-sm">{{ endpoint.path }}</code>
              </div>
              <div class="p-3 text-sm text-gray-600 dark:text-gray-400">{{ endpoint.desc }}</div>
            </div>

          </div>
        </section>

      </div>
    </div>
  </div>
</template>
