# 🛠 Selvod Backend Engine

The Selvod backend is a high-concurrency Go service responsible for video ingestion, orchestration, and security tokenization.

## 🏗 Modular Architecture

We follow a **clean, flattened architecture** to ensure maintainability and testability.

### Core Modules:
- **`api/` & `handler/`**: Separates server lifecycle from request logic. Uses dependency injection to provide handlers with stores and signers.
- **`store/`**: A thread-safe metadata layer using **SQLite in WAL mode**. Optimized for simultaneous transcode status updates.
- **`queue/`**: A configurable worker pool that manages the heavy lifting of background ffmpeg tasks.
- **`signer/`**: The security hub. Generates path-based Hardened Secure Link (MD5) signatures that are RFC 8216 compliant.
- **`transcoder/`**: A robust FFmpeg wrapper that builds Adaptive Bitrate (ABR) HLS streams with optimized GOP settings.
- **`hooks/`**: An event-driven registry that dispatches platform events (Upload, Transcode, Delete) to external webhooks.

## 🔒 Security Model

Selvod implements **Path-Based Signing**. Unlike query-string tokens, our tokens are part of the URL path:
`GET /hls/{token}/{expiry}/{video_id}/index.m3u8`

This ensures that relative paths for HLS segments (`.ts` files) remain valid and signed throughout the playback session without requiring additional client-side logic.

## ⚡ Performance

- **Non-blocking IO:** Uploads are written directly to disk using buffered streams.
- **Transcode Efficiency:** The system tracks the exact start/finish times of every transcode job, allowing you to calculate your hardware's "Real-time Factor."

## 🧪 Development

Run the backend locally:
```bash
go run ./cmd/selvod
```

Run unit tests:
```bash
go test ./...
```
