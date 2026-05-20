# 🎨 Selvod Dashboard SPA

A Vue 3 Single Page Application that serves as an admin console and client demo. It manages the video catalog, library structures, dynamic streaming keys, and showcases adaptive bitrate playbacks.

---

## 🚀 Key Technologies

*   **Vue 3 (Composition API):** Modern reactive structure.
*   **Pinia:** Global state management for persistent background upload queues.
*   **Tailwind CSS v4:** Layout, animations, and typography.
*   **Hls.js & ForgePlayer:** Handles ABR segments and custom token renewal.
*   **Axios:** Handles API calls with request/response authorization interceptors.

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
*   `/api` & `/health` ➡️ backend API container/server.
*   `/hls` ➡️ Nginx HTTPS container (enforcing secure links).

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
