#!/bin/bash
# Selvod Full-Spectrum Delivery Integration Test (Mandatory E2E)

# Configuration for Test Stack
TEST_HTTP_PORT="${SV_TEST_HTTP_PORT:-18080}"
TEST_HTTPS_PORT="${SV_TEST_HTTPS_PORT:-18443}"
API_URL="https://localhost:${TEST_HTTPS_PORT}/api/v1"
EDGE_URL="https://localhost:${TEST_HTTPS_PORT}"
HTTP_URL="http://localhost:${TEST_HTTP_PORT}"
ADMIN_KEY="test-admin-key"
PLAYBACK_KEY="test-playback-key"

echo "SELVOD FULL-SPECTRUM PERIMETER TEST"
echo "------------------------------------"

# 1. Fetch Video ID
CURL_OUT=$(curl -k -s -v -H "Authorization: Bearer $ADMIN_KEY" "$API_URL/videos" 2>&1)
echo "DEBUG CURL OUT: $CURL_OUT"
VIDEO_ID=$(echo "$CURL_OUT" | jq -r '.videos[0].id' 2>/dev/null || echo "")
if [ -z "$VIDEO_ID" ]; then
    echo "[FAIL] No videos found. Seed script failed."
    exit 1
fi

# 2. Test HTTPS Behavior (HTTP should redirect)
echo "[TEST] HTTPS Behavior: HTTP redirects to HTTPS..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$HTTP_URL/hls/any/any/any/master.m3u8")
if [ "$STATUS" == "301" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

# 3. Test Auth Scoping (Playback Key Misuse)
echo "[TEST] Auth Scoping: Playback Key Sign..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $PLAYBACK_KEY" "$API_URL/videos/$VIDEO_ID/stream")
if [ "$STATUS" == "200" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

echo "[TEST] Auth Scoping: Playback Key Admin Restriction..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $PLAYBACK_KEY" "$API_URL/videos")
if [ "$STATUS" == "403" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

# 4. Positive Test: Full HLS Graph
# We need to fetch signature token/expires from the stream API
SIGNED_JSON=$(curl -k -s -H "Authorization: Bearer $ADMIN_KEY" "$API_URL/videos/$VIDEO_ID/stream")
TOKEN=$(echo "$SIGNED_JSON" | jq -r '.token')
EXPIRES=$(echo "$SIGNED_JSON" | jq -r '.expires')

# Clear old cookies
rm -f /tmp/cookies.txt

echo "[TEST] Delivery: Master Manifest (200 OK & Cookie Check)..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" -c /tmp/cookies.txt "$EDGE_URL/hls/00000000-0000-0000-0000-000000000001/$VIDEO_ID/master.m3u8?token=$TOKEN&expires=$EXPIRES")
if [ "$STATUS" == "200" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

echo "[TEST] Delivery: Variant Playlist (200 OK)..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" -b /tmp/cookies.txt "$EDGE_URL/hls/00000000-0000-0000-0000-000000000001/$VIDEO_ID/0/index.m3u8?token=$TOKEN&expires=$EXPIRES")
if [ "$STATUS" == "200" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

echo "[TEST] Delivery: Video Segment (200 OK via Cookie)..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" -b /tmp/cookies.txt "$EDGE_URL/hls/00000000-0000-0000-0000-000000000001/$VIDEO_ID/0/001.ts")
if [ "$STATUS" == "200" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

# 5. Negative Test: 410 Gone (Expired Token)
PAST_EXPIRES=$(($(date +%s) - 3600))
echo "[TEST] Enforcement: Expired Token (410 Gone)..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" "$EDGE_URL/hls/00000000-0000-0000-0000-000000000001/$VIDEO_ID/master.m3u8?token=expired-token&expires=$PAST_EXPIRES")
if [ "$STATUS" == "410" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

# 6. Negative Test: Invalid Token
echo "[TEST] Enforcement: Invalid Token (403 Forbidden)..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" "$EDGE_URL/hls/00000000-0000-0000-0000-000000000001/$VIDEO_ID/master.m3u8?token=badtoken&expires=$EXPIRES")
if [ "$STATUS" == "403" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

# 7. Negative Test: Path Traversal / Direct Access without Token
echo "[TEST] Enforcement: Direct Path Access (401 Unauthorized)..."
STATUS=$(curl -k -s -o /dev/null -w "%{http_code}" "$EDGE_URL/hls/00000000-0000-0000-0000-000000000001/$VIDEO_ID/master.m3u8")
if [ "$STATUS" == "401" ] || [ "$STATUS" == "403" ] || [ "$STATUS" == "404" ]; then echo " PASSED"; else echo " FAILED ($STATUS)"; exit 1; fi

echo ""
echo "✔ FULL-SPECTRUM PERIMETER VERIFIED: Every HLS layer, Auth scope, and HTTPS policy is enforced."
