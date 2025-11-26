package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func LoadSystemTLSConfig(caCertPath string) (*tls.Config, error) {
	certPool := x509.NewCertPool()

	switch caCertPath {
	case "":
		systemCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system certificate pool: %w", err)
		}
		certPool = systemCertPool
	default:
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from %s: %w", caCertPath, err)
		}

		if ok := certPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append CA certificate from %s", caCertPath)
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	return tlsConfig, nil
}
