# 🛠 Selvod Go Backend Engine

The Selvod backend is a high-concurrency Go service responsible for video ingestion, ABR transcoding orchestration, database persistence, and security signing.

---

## 🏗 Modular Architecture

We follow a clean, flat modular structure to keep components decoupled and testable:

*   **`config/`**: Configuration loading and strict environment variable validations.
*   **`api/` & `handler/`**: Routes setup and REST HTTP handlers. Uses dependency injection to wire database stores, background queues, and cryptosigners.
*   **`store/`**: Data persistence layer using SQLite in WAL mode, handling videos, libraries, and library keys.
*   **`queue/`**: Background worker pool managing concurrency and recovering unfinished transcode tasks upon restarts.
*   **`transcoder/`**: Robust FFmpeg wrapper performing raw input verification (`ffprobe`) and multi-bitrate HLS conversions.
*   **`signer/`**: MD5-based cryptographic signing for master manifests and segment validation cookies.
*   **`hooks/`**: Event-driven webhook broadcaster to notify LMS/CMS integrations when media state changes.

---

## 🔒 Security Model: Cookie-Based Segment Verification

Selvod uses a two-phase auth verification to balance security and playback performance:

1.  **Manifest Authentication (Query String):**
    The client fetches the master manifest using signed query parameters:
    `GET /hls/{library_id}/{video_id}/master.m3u8?token={token}&expires={expiry}`
    This request triggers a subrequest authentication to the Go backend (`/api/v1/auth/manifest`), verifying the token against the library's active playback key.
2.  **Segment Authentication (Session Cookie):**
    If the manifest token is valid, the Go backend sets an `sv_session_{video_id}` cookie.
    Subsequent HLS requests for variant manifests and segment `.ts` chunks:
    `GET /hls/{library_id}/{video_id}/{variant}/index.m3u8`
    `GET /hls/{library_id}/{video_id}/{variant}/{segment}.ts`
    are validated directly by Nginx checking the cookie signature against the global `SV_STREAM_SECRET`. This removes database lookups for each video chunk.

---

## ⚙️ Configuration Variables

The backend loads config from environment variables:

| Environment Variable | Description | Default |
| :--- | :--- | :--- |
| `SV_PORT` | Port the backend listens on | `8080` |
| `SV_DB_PATH` | Path to the SQLite WAL database | `data/selvod.db` |
| `SV_STORAGE_PATH` | Storage directory for uploads and HLS output | `data` |
| `SV_MAX_WORKERS` | Max concurrent background transcode workers | `2` |
| `SV_STREAM_SECRET` | Global secret key for segment cookie validation | *Required* |
| `SV_API_KEY` | Master administrative API key | *Required* |
| `SV_PLAYBACK_KEY` | Playback key for signing stream URLs | *Required* |
| `SV_BASE_URL` | Base URL of the edge streaming server | `http://localhost:18080` |
| `SV_WEBHOOK_URL` | Endpoint to post video lifecycle events to | *Optional* |

---

## 🧪 Testing & Execution

### Run Backend Directly:
```bash
go run ./cmd/selvod
```

### Run Unit Tests:
```bash
go test ./...
```

### Run Security Logic Audit:
```bash
go run ./cmd/security-audit/main.go
```
