package authenticator

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/fullsailor/pkcs7"
)

const (
	nameTypeEmail = 1
	nameTypeDNS   = 2
	nameTypeURI   = 6
	nameTypeIP    = 7
)

func marshalSANs(dnsNames, emailAddresses []string, ipAddresses []net.IP, uris []*url.URL) (derBytes []byte, err error) {
	var rawValues []asn1.RawValue
	for _, name := range dnsNames {
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeDNS, Class: asn1.ClassContextSpecific, Bytes: []byte(name)})
	}
	for _, email := range emailAddresses {
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeEmail, Class: asn1.ClassContextSpecific, Bytes: []byte(email)})
	}
	for _, rawIP := range ipAddresses {
		// If possible, we always want to encode IPv4 addresses in 4 bytes.
		ip := rawIP.To4()
		if ip == nil {
			ip = rawIP
		}
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeIP, Class: asn1.ClassContextSpecific, Bytes: ip})
	}
	for _, uri := range uris {
		rawValues = append(rawValues, asn1.RawValue{Tag: nameTypeURI, Class: asn1.ClassContextSpecific, Bytes: []byte(uri.String())})
	}
	return asn1.Marshal(rawValues)
}

// generateSANURI returns the formatted uri(SPIFFEE format for now) for the certificate.
func generateSANURI(namespace, podname string) (string, error) {
	if namespace == "" || podname == "" {
		return "", fmt.Errorf(
			"namespace or podname can't be empty namespace=%v podname=%v", namespace, podname)
	}
	return fmt.Sprintf("spiffe://cluster.local/namespace/%s/podname/%s", namespace, podname), nil
}

func generateCSR(privateKey *rsa.PrivateKey, id string, sanURI *url.URL) ([]byte, error) {
	commonName := strings.Replace(id, "/", ".", -1)
	subj := pkix.Name{
		CommonName: commonName,
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	subjectAltNamesValue, err := marshalSANs(nil, nil, nil, []*url.URL{
		sanURI,
	})
	if err != nil {
		return nil, err
	}

	extSubjectAltName := pkix.Extension{
		Id:       oidExtensionSubjectAltName,
		Critical: false,
		Value:    subjectAltNamesValue,
	}
	template.ExtraExtensions = []pkix.Extension{extSubjectAltName}

	return x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
}

func DecodeFromPEM(PEMBlock []byte, publicCert *x509.Certificate, privateKey crypto.PrivateKey) ([]byte, error) {
	tokenDerBlock, _ := pem.Decode(PEMBlock)
	p7, err := pkcs7.Parse(tokenDerBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return p7.Decrypt(publicCert, privateKey)
}
