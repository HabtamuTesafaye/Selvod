# 🎨 Selvod Dashboard

A premium, Vue 3 reference implementation for managing your VOD library. Built with performance and user experience as top priorities.

## 🚀 Key Technologies

- **Vue 3 (Composition API):** For reactive and maintainable components.
- **Pinia:** Global state management used to track background uploads across routes.
- **Tailwind CSS v4:** Cutting-edge styling with glassmorphism and modern color palettes.
- **Yarn:** For stable, deterministic dependency management.

## 🛠 Features

### 1. Persistent Background Uploads
Uploads are managed by a global Pinia store. You can start an upload, navigate to your library, check other videos, and your upload progress will remain active and visible in the UI.

### 2. ForgePlayer (Seamless HLS)
Standard HLS players often stutter or stop when a playback token expires. **ForgePlayer** solves this by implementing a seamless swap logic:
1.  Detects token expiration.
2.  Fetches a fresh signed URL.
3.  Pauses -> Swaps Source -> Seeks to last position -> Resumes.
This happens in milliseconds, ensuring the viewer never loses their place.

### 3. Real-time Status & Analytics
The dashboard provides visual feedback for:
- **Transcode Efficiency:** How fast your hardware is processing.
- **Storage Availability:** Live disk space monitoring.
- **ABR Readiness:** Indicators for when 1080p/720p variants are ready.

## 🛠 Development

```bash
yarn install
yarn dev
```

## 🏗 Build for Production

```bash
yarn build
```
