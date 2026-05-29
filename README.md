# 🎥 Selvod - Modern Headless VOD Infrastructure

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](./check.sh)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](backend/go.mod)

**Selvod** is a professional, high-performance Video-on-Demand (VOD) engine built for developers who need secure, automated, and scalable video delivery. It handles the entire lifecycle from **raw upload** and **ABR transcoding** to **secure path-based streaming** through a dedicated Nginx sidecar.

---

## 🏗 System Architecture

Selvod uses a decoupled architecture separating heavy media processing from secure client delivery.

<div align="center">

### 📸 Demo Gallery

<div style="overflow-x: auto; white-space: nowrap; padding: 12px 0; border-radius: 12px; -webkit-overflow-scrolling: touch;">
  <a href="demos/demo1.png"><img src="demos/demo1.png" alt="Demo 1" style="height: 280px; display: inline-block; margin: 0 8px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);"></a>
  <a href="demos/demo2.png"><img src="demos/demo2.png" alt="Demo 2" style="height: 280px; display: inline-block; margin: 0 8px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);"></a>
  <a href="demos/demo3.png"><img src="demos/demo3.png" alt="Demo 3" style="height: 280px; display: inline-block; margin: 0 8px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);"></a>
  <a href="demos/demo4.png"><img src="demos/demo4.png" alt="Demo 4" style="height: 280px; display: inline-block; margin: 0 8px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);"></a>
  <a href="demos/demo5.png"><img src="demos/demo5.png" alt="Demo 5" style="height: 280px; display: inline-block; margin: 0 8px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.15);"></a>
</div>
<p style="font-size: 14px; color: #888; margin-top: 8px;">← Scroll or swipe to browse →</p>

</div>

```mermaid
graph TD
    subgraph "Client Layer"
        Client[Vue 3 SPA / Iframe Embed]
    end

    subgraph "Delivery Layer (Nginx Edge)"
        Nginx[Nginx Sidecar]
    end

    subgraph "Application Layer (Go Backend)"
        API[Chi API Router]
        Queue[Worker Queue]
        Transcoder[FFmpeg Engine]
    end

    subgraph "Storage Layer"
        DB[(SQLite WAL Store)]
        Disk[(Shared Disk Volume)]
    end

    %% Upload & Transcoding Flow
    Client -->|1. Upload Raw Media| API
    API -->|2. Register Job| DB
    API -->|3. Dispatch Task| Queue
    Queue -->|4. Process Video| Transcoder
    Transcoder -->|5. Read / Write metadata| DB
    Transcoder -->|6. Write HLS Playlists & Segments| Disk

    %% Playback Authorization & Delivery Flow
    Client -->|7. Request Playback Token| API
    API -->|8. Verify Library Key & Return Signed URL| Client
    Client -->|9. GET master.m3u8 with token| Nginx
    Nginx -->|10. Subrequest validation /auth/manifest| API
    API -->|11. Verify HMAC Signature & Set Session Cookie| Nginx
    Nginx -->|12. Return Manifest & Cookie| Client
    Client -->|13. Request segments with Cookie| Nginx
    Nginx -->|14. Local secure_link verification bypasses Go| Disk
    Nginx -->|15. Stream TS Chunks| Client
```

---

## 🚀 Quick Start (Production/Dockerized Stack)

Deploy with absolute verification and automated production key generation:

```bash
./deploy_production.sh
```

This script:

1. Verifies that high-entropy secrets exist (and auto-generates them if defaults are found).
2. Runs the system integrity and verification checks (`./check.sh`).
3. Launches the production container stack.

---

## 💻 Local Development & Testing

Local testing requires coordination between the frontend dev server, the Go backend API, and Nginx (handling the cookie-based HLS secure links).

### 1. Prerequisites

Ensure you have the following installed locally:

- **Go** (1.23+)
- **Node.js & Yarn**
- **Docker & Docker Compose**
- **FFmpeg / ffprobe** (if running bare-metal)

### 2. Run via Automated Local Script (Recommended)

You can automate certificate generation and stack launch in one command:

```bash
./deploy_local.sh
```

Alternatively, you can perform these steps manually:

- **Set Up Local SSL Certificates:**
  ```bash
  ./setup_certs.sh
  ```
- **Run via Docker Compose:**
  `bash
    docker compose -f docker-compose.local.yml up --build
    `
  Under this setup:
- **Backend API** is exposed on `http://localhost:8081`.
- **Nginx Edge** is exposed on `http://localhost:18080` (HTTP) and `https://localhost:18443` (HTTPS).
- **Database & Media assets** are mounted locally in `./data`.

Now run the frontend development server:

```bash
cd frontend
yarn install
yarn dev
```

- The frontend dev server runs at `http://localhost:5173`.
- Vite config proxies `/api` and `/health` requests to `http://localhost:8081` (backend).
- Vite config proxies `/hls` requests to `https://localhost:18443` (Nginx Edge) while allowing self-signed certificates (`secure: false`).

### 4. Option B: Run Go Backend Bare-Metal

If you prefer running the Go compiler locally for instant updates:

```bash
# Generate self-signed certs first
./setup_certs.sh

# Run backend directly (default port: 8080)
cd backend
go run ./cmd/selvod
```

> [!NOTE]
> Running Go bare-metal binds it to port `8080` instead of the Docker mapped port `8081`.
> You must create `frontend/.env.local` to tell Vite where to proxy requests:
>
> ```env
> VITE_API_TARGET=http://localhost:8080
> VITE_HLS_TARGET=https://localhost:18443
> ```
>
> And run Nginx locally or map the port so the frontend can retrieve HLS segments.

---

## 🧪 Testing Suites

Selvod implements three levels of validation:

### 1. Backend Unit Tests

Validates logic, stores, and signers.

```bash
cd backend
go test ./...
```

### 2. Cryptographic Security Auditing

Verifies that signatures do not suffer from token drift and maintain strict cross-tenant isolation.

```bash
cd backend
go run ./cmd/security-audit/main.go
```

### 3. E2E Live Perimeter Integration Tests (Absolute Veracity)

Runs a full E2E simulation using a temporary isolated Docker environment:

```bash
./check.sh
```

What `check.sh` does under the hood:

1.  Generates mock certificates and copies them to `test_data/certs`.
2.  Seeds a test SQLite database containing a default library and active playback keys (`seed-audit`).
3.  Launches a test container stack via `docker-compose.test.yml`.
4.  Executes `./test_delivery.sh` which checks:
    - **HTTPS Redirection:** Confirms HTTP redirects to HTTPS.
    - **Auth Scoping:** Verifies Playback Keys can request streams but cannot invoke admin endpoints.
    - **Stream Authorization:** Requests the master manifest, variant playlists, and video segments to ensure cookies are set and validated properly.
    - **Token Enforcement:** Tests that expired, malformed, or missing tokens return `410 Gone` and `403 Forbidden` statuses.
5.  Tears down and deletes `test_data`.

---

## 🔒 Production Hardening & Key Generation

> [!WARNING]
> **CRITICAL SECURITY REQUIREMENT:**
> Under no circumstances should default local environment keys be committed to source control or reused in live production environments. Doing so allows malicious actors to sign arbitrary playbacks, gain administrative upload capabilities, or compromise server directories.

To transition to production safely:

1.  **Generate High-Entropy Secrets:**

    ```bash
    # Stream Signer Secret (used by Nginx and backend to validate cookie segments)
    openssl rand -hex 32

    # Master Admin API Key (full access to uploads, deletions, and library management)
    openssl rand -hex 32

    # Playback Scoped Key (only allowed to generate signed URLs)
    openssl rand -hex 32
    ```

2.  Inject these values securely into your environment variables or secret manager (`SV_STREAM_SECRET`, `SV_API_KEY`, `SV_PLAYBACK_KEY`).
3.  Ensure your firewall blocks external access to the Go backend port (default: `8080`), forcing all client traffic to go through the hardened Nginx delivery container.

---

## 🎬 Client Integration Guide

Selvod uses **library-scoped playback keys** to isolate access. Each library has its own unique key — a client with Library A's key **cannot** play videos from Library B, even if they know the video ID.

### Authentication Model

| Key Type                            | Scope                                       | Use Case                                     |
| :---------------------------------- | :------------------------------------------ | :------------------------------------------- |
| `SV_API_KEY` (Admin)                | Full access to all endpoints                | Server-side admin panel, uploading, deleting |
| `SV_PLAYBACK_KEY` (Global Playback) | `/stream` endpoint only, all libraries      | Trusted internal services                    |
| **Library Playback Key**            | `/stream` endpoint only, **single library** | Client apps, third-party integrations        |

### Option 1: Embed Player (Recommended for Websites)

Selvod ships a built-in embed player at `/embed.html`. Embed it as an iframe:

```html
<iframe
  src="https://vod.example.com/embed.html?videoId=YOUR_VIDEO_ID&key=YOUR_LIBRARY_KEY"
  width="800"
  height="450"
  frameborder="0"
  allow="autoplay; fullscreen"
  allowfullscreen
></iframe>
```

The embed player automatically:

- Authenticates using the library key as a Bearer token
- Loads the HLS stream with adaptive bitrate
- Refreshes tokens silently at 75% of their lifetime
- Shows custom error screens (403 Access Denied, 404 Not Found, 423 Still Processing)

### Option 2: Direct API Integration (Custom Players)

For custom player implementations, use the REST API directly:

**Step 1: Get a signed stream URL**

```bash
curl -H "Authorization: Bearer YOUR_LIBRARY_PLAYBACK_KEY" \
     https://vod.example.com/api/v1/videos/VIDEO_ID/stream
```

**Response:**

```json
{
  "url": "https://vod.example.com/hls/{lib_id}/{vid_id}/master.m3u8?token=xxx&expires=1234567890",
  "token": "xxx",
  "expires": 1234567890,
  "expires_in": 7200
}
```

**Step 2: Load the `url` into any HLS player (hls.js, Video.js, ExoPlayer, AVPlayer)**

> [!IMPORTANT]
> The stream URL expires after `expires_in` seconds. Your player must call `/stream` again before expiry and hot-swap the source to avoid playback interruption.

### Security Guarantees

- A library key can **only** sign streams for videos inside its own library
- Even with a valid video ID, a key from a different library returns `403 Forbidden`
- Revoking or regenerating a library key instantly invalidates all active sessions for that library
- Segment-level cookies are cryptographically bound to both the video ID and global stream secret

---

## 📖 Component Map

| Module       | Description                                                                            | Location                            |
| :----------- | :------------------------------------------------------------------------------------- | :---------------------------------- |
| **Backend**  | Go API Router, Worker Queue, SQLite WAL store, Webhook dispatches, FFmpeg wrapper, and OpenAPI docs at `/docs`. | [`/backend`](./backend/README.md)   |
| **Frontend** | Vue 3 Dashboard, Pinia Upload state, and ForgePlayer.                                  | [`/frontend`](./frontend/README.md) |
| **Nginx**    | Reverse proxy, SSL termination, and Secure Link MD5 cookie check.                      | [`/nginx`](./nginx/README.md)       |

---
