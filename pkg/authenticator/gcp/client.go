package gcp

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

func newHTTPSClient(CACert []byte) (*http.Client, error) {
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(CACert)
	if !ok {
		return nil, log.RecordedError(log.CAKC014)
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	// Doubt this is necessary because there's only one
	//tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{Transport: transport, Timeout: time.Second * 10}, nil
}
