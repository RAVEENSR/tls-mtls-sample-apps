package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HEALTH] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ECHO] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	headers := make(map[string][]string)
	for k, v := range r.Header {
		headers[k] = v
	}

	tlsInfo := map[string]any{}
	if r.TLS != nil {
		tlsInfo["version"] = tlsVersionName(r.TLS.Version)
		tlsInfo["cipher_suite"] = tls.CipherSuiteName(r.TLS.CipherSuite)
		tlsInfo["server_name"] = r.TLS.ServerName
		tlsInfo["negotiated_protocol"] = r.TLS.NegotiatedProtocol

		if len(r.TLS.PeerCertificates) > 0 {
			clientCert := r.TLS.PeerCertificates[0]
			ips := []string{}
			for _, ip := range clientCert.IPAddresses {
				ips = append(ips, ip.String())
			}
			tlsInfo["client_cert"] = map[string]any{
				"subject":      clientCert.Subject.String(),
				"issuer":       clientCert.Issuer.String(),
				"not_before":   clientCert.NotBefore.UTC().Format(time.RFC3339),
				"not_after":    clientCert.NotAfter.UTC().Format(time.RFC3339),
				"dns_names":    clientCert.DNSNames,
				"ip_addresses": ips,
			}
		} else {
			tlsInfo["client_cert"] = nil
		}
	}

	response := map[string]any{
		"method":      r.Method,
		"url":         r.URL.String(),
		"remote_addr": r.RemoteAddr,
		"headers":     headers,
		"tls":         tlsInfo,
		"time":        time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(response)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ROOT] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	fmt.Fprintf(w, "mTLS Test Server is running.\nUse /health or /echo for testing.\n")
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown(0x%04x)", v)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "443"
	}

	certFile := os.Getenv("CERT_FILE")
	if certFile == "" {
		certFile = "server.crt"
	}

	keyFile := os.Getenv("KEY_FILE")
	if keyFile == "" {
		keyFile = "server.key"
	}

	// CLIENT_CA_FILE: CA cert used to verify the Choreo proxy's client certificate.
	// If set  → mTLS is enforced (client cert required and verified).
	// If unset → one-way TLS only.
	clientCAFile := os.Getenv("CLIENT_CA_FILE")

	for _, f := range []string{certFile, keyFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			log.Fatalf("File not found: %s", f)
		}
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if clientCAFile != "" {
		caCert, err := os.ReadFile(clientCAFile)
		if err != nil {
			log.Fatalf("Failed to read CLIENT_CA_FILE %s: %v", clientCAFile, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			log.Fatalf("Failed to parse CA cert from %s", clientCAFile)
		}
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		log.Printf("mTLS enabled — client CA: %s", clientCAFile)
	} else {
		cfg.ClientAuth = tls.NoClientCert
		log.Printf("mTLS disabled — set CLIENT_CA_FILE env var to enable")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/", rootHandler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		TLSConfig:    cfg,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Starting mTLS server on port %s", port)
	log.Printf("  Certificate : %s", certFile)
	log.Printf("  Private Key : %s", keyFile)
	log.Printf("  Endpoints   : /health  /echo  /")

	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
