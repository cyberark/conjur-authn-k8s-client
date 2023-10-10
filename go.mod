module github.com/cyberark/conjur-authn-k8s-client

go 1.19

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.7.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.7.0 // indirect
	go.opentelemetry.io/otel/sdk v1.7.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	// Version number used here is ignored
	github.com/cyberark/conjur-opentelemetry-tracer v0.0.1-1321.0.20231010135527-11285e1be165
	github.com/davecgh/go-spew v1.1.1 // indirect
)

replace gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c => gopkg.in/yaml.v3 v3.0.1

replace golang.org/x/sys v0.0.0-20210423185535-09eb48e85fd7 => golang.org/x/sys v0.8.0

// DO NOT EDIT: CHANGES TO THE BELOW LINE WILL BREAK AUTOMATED RELEASES
replace github.com/cyberark/conjur-opentelemetry-tracer => github.com/cyberark/conjur-opentelemetry-tracer v0.0.1-1321.0.20231010135527-11285e1be165
