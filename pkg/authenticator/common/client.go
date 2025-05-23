package common

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// NewHTTPSClient Returns https client to communicate with Conjur
func NewHTTPSClient(CACert []byte, certPEMBlock, keyPEMBlock []byte) (*http.Client, error) {
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(CACert)
	if !ok {
		return nil, log.RecordedError(log.CAKC014)
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	if certPEMBlock != nil && keyPEMBlock != nil {
		cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if err != nil {
			return nil, log.RecordedError(log.CAKC017, err)
		}

		tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &cert, nil
		}
	}
	// Doubt this is necessary because there's only one
	//tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig, Proxy: http.ProxyFromEnvironment}

	return &http.Client{Transport: transport, Timeout: time.Second * 10}, nil
}
