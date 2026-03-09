#!/bin/bash
# Generates a self-signed TLS certificate for the VM.
#
# Usage:
#   ./generate_certs.sh                        # localhost only
#   ./generate_certs.sh 20.10.5.123            # include VM public IP
#   ./generate_certs.sh 20.10.5.123 myhost.com # include VM public IP + domain

set -e

VM_IP="${1:-}"
DOMAIN="${2:-}"
DAYS=3650

# Build SAN list
SAN="DNS:localhost,IP:127.0.0.1"
if [ -n "$VM_IP" ]; then
  SAN="${SAN},IP:${VM_IP}"
fi
if [ -n "$DOMAIN" ]; then
  SAN="${SAN},DNS:${DOMAIN}"
fi

CN="${DOMAIN:-${VM_IP:-localhost}}"

echo "Generating self-signed certificate..."
echo "  CN  : ${CN}"
echo "  SAN : ${SAN}"
echo "  Days: ${DAYS}"
echo ""

openssl req -x509 -newkey rsa:4096 -sha256 -days "${DAYS}" \
  -nodes \
  -keyout server.key \
  -out server.crt \
  -subj "/CN=${CN}" \
  -addext "subjectAltName=${SAN}"

echo ""
echo "Done! Files generated:"
echo "  server.crt  <- upload this public cert to WSO2 Choreo as the endpoint certificate"
echo "  server.key  <- keep this on the VM only, never share it"
echo ""
echo "To inspect the cert:"
echo "  openssl x509 -in server.crt -text -noout"
