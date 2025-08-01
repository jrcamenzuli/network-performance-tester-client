package util

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
)

// CreateHTTPSClient creates an HTTP client that trusts our custom CA
func CreateHTTPSClient() *http.Client {
	// Load CA certificate
	caCert, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		// If CA cert not found, return regular client (for HTTP tests)
		return &http.Client{}
	}

	// Create CA certificate pool
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		// If CA cert parsing fails, return regular client
		return &http.Client{}
	}

	// Create TLS config with our CA
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	// Create HTTP client with custom TLS config
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}
