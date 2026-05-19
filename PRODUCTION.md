# Selvod Production Lockdown Guide

Selvod uses a dual-compose architecture to separate Local Development from Production Hardening.

## 1. Choosing your Stack

### Local Development
Run this for a zero-config experience (HTTP only, no certs):
```bash
docker-compose -f docker-compose.local.yml up --build
```

### Production Deployment
Run this for a TITAN-Hardened internet-facing deployment (HTTPS mandatory):
```bash
docker-compose up --build
```

## 2. Production Lockdown Steps

1. **Obtain SSL Certificates**:
   Place `server.crt` and `server.key` into `./nginx/certs/`. The production `docker-compose.yml` will fail to start Nginx if these are missing.

2. **Configure Production Secrets**:
   Create a real `.env` file. Do NOT use the default values. You must provide:
   - `SV_STREAM_SECRET`
   - `SV_API_KEY`
   - `SV_PLAYBACK_KEY`
   - `SV_BASE_URL` (Must start with `https://`)
   - `SV_FRONTEND_URL` (Exact domain of your dashboard)

3. **Firewall Lockdown**:
   Only Port **80** (for redirect) and **443** (for HTTPS) should be open to the public internet on your host.

4. **Run Integrity Check**:
   Before deploying, always run:
   ```bash
   ./check.sh
   ```
   This ensures no "changeme" secrets or insecure code has slipped into the build.
