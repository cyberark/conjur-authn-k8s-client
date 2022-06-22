module github.com/cyberark/conjur-authn-k8s-client

go 1.16

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa
	github.com/stretchr/testify v1.7.2
	go.opentelemetry.io/otel v1.7.0
)

require (
	// Requires latest version for automated release process
	github.com/cyberark/conjur-opentelemetry-tracer v0.0.1-336
	github.com/davecgh/go-spew v1.1.1 // indirect
)

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1

replace github.com/cyberark/conjur-opentelemetry-tracer => github.com/cyberark/conjur-opentelemetry-tracer latest
