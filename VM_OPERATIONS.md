# VM Operations Guide

**VM:** `20.75.77.58`
**SSH Key:** `/Users/raveen/Desktop/choreocertmgt.pem`

---

## SSH into VM

```bash
VM_IP="20.75.77.58"
SSH_KEY="/Users/raveen/Desktop/choreocertmgt.pem"

ssh -i "$SSH_KEY" azureuser@"$VM_IP"
```

---

## One-Way TLS Server (`tls_server_linux_amd64`)

Listens on port **443**. Verifies only the server certificate — no client cert required.

### Start
```bash
cd ~/tls-server
sudo nohup ./tls_server_linux_amd64 > server.log 2>&1 &
```

### Stop
```bash
sudo pkill -f tls_server_linux_amd64 || true
```

### Check if running
```bash
sudo ss -tlnp | grep 443
sudo pgrep -a -f tls_server_linux_amd64
```

### View logs
```bash
# Live
tail -f ~/tls-server/server.log

# Last 50 lines
tail -50 ~/tls-server/server.log
```

### Test from laptop
```bash
# Set these once in your terminal session (from the repo root)
VM_IP="20.75.77.58"
TLS_CERTS="$(pwd)/generated-certs/tls"

# Quick test (skips cert verification)
curl -k https://"$VM_IP"/health

# Proper test (verifies server cert)
curl --cacert "$TLS_CERTS/server.pem" https://"$VM_IP"/health

# Echo endpoint — inspect headers and TLS details
curl -k https://"$VM_IP"/echo

# Root endpoint
curl -k https://"$VM_IP"/
```

---

## mTLS Server (`mtls_server_linux_amd64`)

Listens on port **443** (stop the one-way TLS server first). Both sides verify each other's certificates.

### Start
```bash
cd ~/tls-server

# Stop one-way TLS server first
sudo pkill -f tls_server_linux_amd64 || true

# Start mTLS server
sudo PORT=443 CLIENT_CA_FILE=combined_ca.pem nohup ./mtls_server_linux_amd64 > mtls_server.log 2>&1 &
```

### Stop
```bash
sudo pkill -f mtls_server_linux_amd64 || true
```

### Check if running
```bash
sudo ss -tlnp | grep 443
sudo pgrep -a -f mtls_server_linux_amd64
```

### View logs
```bash
# Live
tail -f ~/tls-server/mtls_server.log

# Last 50 lines
tail -50 ~/tls-server/mtls_server.log
```

### Test from laptop (mTLS — requires client cert)
```bash
# Set these once in your terminal session (from the repo root)
VM_IP="20.75.77.58"
TLS_CERTS="$(pwd)/generated-certs/tls"
MTLS_CERTS="$(pwd)/generated-certs/mtls"

# Health check
curl --cacert "$TLS_CERTS/server.pem" \
     --cert "$MTLS_CERTS/laptop_client.crt" \
     --key "$MTLS_CERTS/laptop_client.key" \
     https://"$VM_IP"/health

# Echo endpoint — inspect headers, TLS details, and client_cert block
curl --cacert "$TLS_CERTS/server.pem" \
     --cert "$MTLS_CERTS/laptop_client.crt" \
     --key "$MTLS_CERTS/laptop_client.key" \
     https://"$VM_IP"/echo

# Root endpoint
curl --cacert "$TLS_CERTS/server.pem" \
     --cert "$MTLS_CERTS/laptop_client.crt" \
     --key "$MTLS_CERTS/laptop_client.key" \
     https://"$VM_IP"/

# Skip cert verification (quick sanity check only)
curl -k --cert "$MTLS_CERTS/laptop_client.crt" \
        --key "$MTLS_CERTS/laptop_client.key" \
        https://"$VM_IP"/health
```

---

## Files on the VM (`~/tls-server/`)

| File | Description |
|---|---|
| `tls_server_linux_amd64` | One-way TLS server binary |
| `mtls_server_linux_amd64` | mTLS server binary |
| `server.crt` | Server public certificate |
| `server.key` | Server private key |
| `choreo_client_ca.pem` | Choreo-generated client certificate (trusted CA for mTLS) |
| `laptop_client.crt` | Laptop test client certificate |
| `combined_ca.pem` | Choreo + laptop certs combined — used as CLIENT_CA_FILE |
| `server.log` | One-way TLS server logs |
| `mtls_server.log` | mTLS server logs |

---

## Certs on Local Machine (`generated-certs/`)

### `generated-certs/tls/` — Server certificates (one-way TLS)

| File | Description |
|---|---|
| `tls/server.crt` | Server public certificate (used by the server at runtime) |
| `tls/server.key` | Server private key |
| `tls/server.pem` | Server public cert in PEM format — upload to Choreo as endpoint certificate |

### `generated-certs/mtls/` — Client certificates (mTLS)

| File | Description |
|---|---|
| `mtls/laptop_client.crt` | Laptop client cert — used with curl for mTLS testing |
| `mtls/laptop_client.key` | Laptop client private key — used with curl for mTLS testing |

### `generated-certs/generate_certs.sh`
Regenerates the server cert. Run from the `generated-certs/` folder:
```bash
cd generated-certs
./generate_certs.sh <VM_IP>
```

---

## Notes

- Only **one server can run on port 443 at a time** — always stop one before starting the other.
- Port **443** is pre-opened in the Azure NSG by default.
- The `/echo` endpoint is the most useful for debugging — it reflects back all headers and TLS handshake details including the client certificate info when mTLS is working.
- When mTLS is working, the `/echo` response will contain a `client_cert` block. If `client_cert` is `null`, the client did not present a certificate.
