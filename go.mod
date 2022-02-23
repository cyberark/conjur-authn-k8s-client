module github.com/cyberark/conjur-authn-k8s-client

go 1.17

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cyberark/conjur-opentelemetry-tracer v0.0.0-20220113161145-73452511df0c
	github.com/fullsailor/pkcs7 v0.0.0-20190404230743-d7302db945fa
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.3.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.3.0 // indirect
	go.opentelemetry.io/otel/sdk v1.3.0 // indirect
	go.opentelemetry.io/otel/trace v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/cyberark/conjur-opentelemetry-tracer =>	./modules/conjur-opentelemetry-tracer