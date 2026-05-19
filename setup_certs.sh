#!/bin/bash
# Selvod TLS Generator

CERT_DIR="./data/certs"
mkdir -p $CERT_DIR

if [ -f "$CERT_DIR/server.crt" ]; then
    echo "Certs already exist. Skipping generation."
    exit 0
fi

echo "Generating self-signed certificates for Selvod..."
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$CERT_DIR/server.key" \
    -out "$CERT_DIR/server.crt" \
    -subj "/C=US/ST=Selvod/L=Dev/O=Selvod/OU=Engineering/CN=localhost"

echo "TLS Setup Complete. Certificates stored in $CERT_DIR"
