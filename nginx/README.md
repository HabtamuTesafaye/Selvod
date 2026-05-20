# 🕸️ Selvod Nginx Adaptive Edge

Nginx acts as the secure entry point (reverse proxy, SSL terminator, and CDN/delivery gate) for all video playback requests and API routing.

---

## ⚙️ Configuration & Environment Variables

Nginx utilizes the standard `nginx:alpine` image with built-in runtime template parsing. At startup, environment variables are injected into `default.conf.template` to yield the final active configuration:

*   `SV_SERVER_NAME`: The domains Nginx listens to (e.g., `localhost` or `vod.example.com`).
*   `SV_FRONTEND_URL`: CORS configuration, allowing credentials and cross-origin resource sharing from the frontend SPA.
*   `SV_STREAM_SECRET`: Cryptographic secret key used by `secure_link_md5` to validate cookie sessions.

---

## 🔒 Security Architectures

Nginx protects video content using two layers of path-based token and cookie checks:

### 1. Master Manifest Request (`auth_request`)
When a client requests the master manifest playlist:
`GET /hls/{library_id}/{video_id}/master.m3u8?token={token}&expires={expiry}`

*   Nginx intercepts this request and initiates an internal subrequest to the Go backend:
    `POST /api/v1/auth/manifest`
*   The Go backend verifies the signed token against the library's active playback key.
*   If valid, Go returns `200 OK` and a `Set-Cookie` header containing the HLS session cookie (`sv_session_{video_id}`).
*   Nginx appends this cookie to the response, rewrites the request path to serve the static file on disk, and passes it to the client.

### 2. Segment and Variant Requests (`secure_link`)
Subsequent variant manifest playlists and `.ts` chunk requests bypass Go database lookups entirely:
`GET /hls/{library_id}/{video_id}/{variant}/index.m3u8`
`GET /hls/{library_id}/{video_id}/{variant}/{segment_index}.ts`

*   Nginx parses the `sv_session_{video_id}` cookie to extract the signature and expiry.
*   It computes the MD5 checksum using:
    `secure_link_md5 "$SV_STREAM_SECRET/$vid_id/$secure_link_expires";`
*   If the cookie is valid and not expired, Nginx serves the segment directly from disk storage (read-only mount).
*   If invalid or missing, Nginx returns `403 Forbidden`.
*   If expired, Nginx returns `410 Gone`.

---

## 📂 Certificates

To support HTTPS for secure cookies in development and production:
*   **Production:** Mount your verified certificates folder to `/etc/nginx/certs` containing `server.crt` and `server.key`.
*   **Local Development:** Run `./setup_certs.sh` from the project root to generate self-signed certificates in `./data/certs/` which are mounted automatically.
