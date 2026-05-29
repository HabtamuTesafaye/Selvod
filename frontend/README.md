# 🎨 Selvod Dashboard SPA

A Vue 3 Single Page Application that serves as an admin console and client demo. It manages the video catalog, library structures, dynamic streaming keys, and showcases adaptive bitrate playbacks.

---

## 🚀 Key Technologies

- **Vue 3 (Composition API):** Modern reactive structure.
- **Pinia:** Global state management for persistent background upload queues.
- **Tailwind CSS v4:** Layout, animations, and typography.
- **Hls.js & ForgePlayer:** Handles ABR segments and custom token renewal.
- **Axios:** Handles API calls with request/response authorization interceptors.

---

## 🔐 Security Architecture

### Key Types

| Key                      | Scope                                        | Used By                         | Storage                             |
| ------------------------ | -------------------------------------------- | ------------------------------- | ----------------------------------- |
| **Admin API Key**        | Full CRUD (videos, libraries, keys)          | Dashboard only                  | `sessionStorage` (browser)          |
| **Playback Key**         | Stream URL signing (`/stream` endpoint)      | Dashboard player                | `sessionStorage` (browser)          |
| **Library Playback Key** | Scoped stream signing for a specific library | External clients / embed player | Passed via URL param or server-side |

> **The Admin API Key should never be shared or exposed publicly.** The dashboard stores it in `sessionStorage` (cleared on tab close) as a self-hosted admin tool. If exposing the dashboard publicly, consider switching to an `httpOnly` cookie session.

### Token Lifecycle

1. **Stream Signing (2 hours):** `GET /api/v1/videos/:id/stream` returns a signed HLS manifest URL with a 2-hour expiry.
2. **Embed Signing (client-side):** The embed player receives the library key via URL parameter and requests stream signing directly.
3. **75% Refresh:** `ForgePlayer.vue` schedules a token refresh at 75% of `expires_in`, seamlessly reloading the HLS source without interrupting playback.
4. **Segment Cookies:** On manifest auth, the backend sets a `sv_session_{video_id}` cookie that Nginx validates on each `.ts` segment request via `secure_link_md5`.

### Token Format (Nginx `secure_link`)

```
MD5(playback_secret / library_id / video_id / expires) → Base64URL
```

This is an inherent constraint of Nginx's `secure_link` module (MD5 only). Tokens are short-lived and segment-level cookies provide a second auth layer.

### Credential Handling

- Credentials are stored in **`sessionStorage` only** — they clear when the browser tab is closed.
- An **idle timeout** (30 min) auto-clears credentials on inactivity.
- `localStorage` is **not** used for credential storage.

### Library-Scoped Access

All video queries are scoped to `activeLibraryId`. The API interceptor in `lib/api.js` automatically attaches the correct Bearer token:

- **Admin key** for CRUD operations
- **Playback key** for stream signing (`/stream` endpoint)

Library keys (generated via the dashboard) allow external clients to access only videos within their designated library.

---

## 🛠 Features

### 1. Persistent Background Uploads

Background uploads are managed by `src/stores/upload.js`. You can queue multiple files, navigate across dashboard views, and the upload progress persists in the background.

### 2. ForgePlayer (Seamless HLS Token Renewal)

Standard video players stutter or fail when security tokens expire mid-stream. **ForgePlayer** (`src/components/ForgePlayer.vue`) intercepts expiration, requests a new signed token, updates the source, and resumes playback in milliseconds.

### 3. Library & Keys Management

Enables creating multi-tenant libraries and generating, revoking, or regenerating keys.

---

## 💻 Local Development Setup

### 1. Configure Dev Proxies

The frontend relies on Vite's local dev server proxy configured in `vite.config.js`. It directs requests as follows:

- `/api` & `/health` backend API container/server.
- `/hls` Nginx HTTPS container (enforcing secure links).

Create `frontend/.env` (defaults are loaded automatically) or `frontend/.env.local` to customize target hosts:

```env
# If Go backend runs inside Docker-Compose Local (exposed host port: 8081)
VITE_API_TARGET=http://localhost:8081

# Target for HLS media streams served by Nginx
VITE_HLS_TARGET=https://localhost:18443
```

### 2. Run the App

```bash
# Install dependencies
yarn install

# Spin up Vite Dev Server (served at http://localhost:5173)
yarn dev
```

---

## 🏗 Build for Production

```bash
yarn build
```

This outputs static assets to `frontend/dist/`, ready to be served by Nginx or your production web server.
