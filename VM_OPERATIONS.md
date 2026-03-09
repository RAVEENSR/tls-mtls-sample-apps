# VM Operations Guide

**VM:** `20.75.77.58`
**SSH Key:** `/Users/raveen/Desktop/choreocertmgt.pem`

---

## SSH into VM

```bash
ssh -i /Users/raveen/Desktop/choreocertmgt.pem azureuser@20.75.77.58
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
# Quick test (skips cert verification)
curl -k https://20.75.77.58/health

# Proper test (verifies server cert)
curl --cacert /Users/raveen/Desktop/sample-go-app/server.pem https://20.75.77.58/health

# Echo endpoint — inspect headers and TLS details
curl -k https://20.75.77.58/echo

# Root endpoint
curl -k https://20.75.77.58/
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
curl --cacert /Users/raveen/Desktop/sample-go-app/server.pem \
     --cert /Users/raveen/Desktop/sample-go-app/laptop_client.crt \
     --key /Users/raveen/Desktop/sample-go-app/laptop_client.key \
     https://20.75.77.58/echo
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

## Certs on Local Machine

| File | Description |
|---|---|
| `server.pem` | Server public cert — upload to Choreo as endpoint certificate |
| `laptop_client.crt` | Laptop client cert — used with curl for mTLS testing |
| `laptop_client.key` | Laptop client private key — used with curl for mTLS testing |

---

## Notes

- Only **one server can run on port 443 at a time** — always stop one before starting the other.
- Port **443** is pre-opened in the Azure NSG by default.
- The `/echo` endpoint is the most useful for debugging — it reflects back all headers and TLS handshake details including the client certificate info when mTLS is working.
- When mTLS is working, the `/echo` response will contain a `client_cert` block. If `client_cert` is `null`, the client did not present a certificate.
