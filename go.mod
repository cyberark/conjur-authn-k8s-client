module github.com/cyberark/conjur-authn-k8s-client

go 1.17

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cyberark/secrets-provider-for-k8s v1.3.0
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
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace go.opentelemetry.io/otel v1.3.0 => ./third-party/go.opentelemetry.io/otel
replace github.com/cyberark/secrets-provider-for-k8s v1.3.0 => ./third-party/secrets-provider-for-k8s
