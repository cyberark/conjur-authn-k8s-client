module github.com/cyberark/conjur-authn-k8s-client

go 1.24.2

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	// Version number used here is ignored
	github.com/cyberark/conjur-opentelemetry-tracer v1.55.55
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/otel v1.35.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1

replace golang.org/x/sys v0.0.0-20210423185535-09eb48e85fd7 => golang.org/x/sys v0.8.0

// DO NOT EDIT: CHANGES TO THE BELOW LINE WILL BREAK AUTOMATED RELEASES
replace github.com/cyberark/conjur-opentelemetry-tracer => github.com/cyberark/conjur-opentelemetry-tracer latest
