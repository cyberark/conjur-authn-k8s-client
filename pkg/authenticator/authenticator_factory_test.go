package authenticator

import (
	"reflect"
	"testing"
	"time"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/memory"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/config"
	jwtAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/jwt"
	k8sAuthenticator "github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/k8s"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Valid PEM certificate for testing
var validCert = []byte(`-----BEGIN CERTIFICATE-----
MIIDhDCCAmygAwIBAgIRAOvfs9f9fsNFNaZ/2+J6UrAwDQYJKoZIhvcNAQELBQAw
GDEWMBQGA1UEAxMNY29uanVyLW9zcy1jYTAeFw0yMTAzMjAxNTQ5MTNaFw0yMjAz
MjAxNTQ5MTNaMBsxGTAXBgNVBAMTEGNvbmp1ci5teW9yZy5jb20wggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDKxOHxRYU+5v7VuMle2cKd964JRr2uG8hy
5s93RMa7ldwOCemL3KPflyRAeJf/Xwf21jfVJGPZLlMFNlYTDn2d25nyk+micqYh
8hAQXQoR3adLxcgsGLXYfhaa+J9JsRsE5KvrBkmFIDtj3HENZp9s4793sm9zeEEe
nVUPPvLjB3t4xHGeRqiqH3TLzqIzn2JO8TdsFj4Z/VP+soKt7Mgoo7DX2DrDhllt
seRjsa7rdZBdvrE5lwIE8Np/OIjv/6fLyFxRZmon50y3mexmyp/SoLUkAdrYR9qu
ih1xXt5PG8sCqkcyzVHP1xmXmHX/tm6kdLlmEKu/CHURMar29ruvAgMBAAGjgcUw
gcIwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcD
AjAMBgNVHRMBAf8EAjAAMIGCBgNVHREEezB5ghBjb25qdXIubXlvcmcuY29tggpj
b25qdXItb3NzghVjb25qdXItb3NzLmNvbmp1ci1vc3OCGWNvbmp1ci1vc3MuY29u
anVyLW9zcy5zdmOCJ2Nvbmp1ci1vc3MuY29uanVyLW9zcy5zdmMuY2x1c3Rlci5s
b2NhbDANBgkqhkiG9w0BAQsFAAOCAQEAQy6w5w+hJikQ3m/nVexOzW2tfxD8fPha
qiNVNvVoK2pt693WEE5wSdG+h4U+0yRRr//12PkkZoZ6YGh+JlS1+lrp94jJzYcd
joEpt0I6saNMGz8Aj4SeA5no9w708kS7UTcbHcYsdjH3mwgL6/NSsQmRNRObSax+
75oNu8qKW2JHntXSLWLIvlYJlPkybBPdZlTl0DeGUQVTCnMbjodGt9Xrf43WSHBU
hvGOLg3fDwp3II8a2p+s+3QUJ7pl1nfuehnusp0zHv6dfQs67y/NL95xZCoYXPZl
LSv3Pj2ixKsaKCYly9Cy8Qgdf25ON8xcrzwc4F8goozEhINBY5hf1A==
-----END CERTIFICATE-----`)

// Create test cases that we can use for both `NewAuthenticator` and `NewAuthenticatorWithAccessToken`
type testCase struct {
	name              string
	config            config.Configuration
	expectedError     string
	expectedAuthnType reflect.Type
}

func buildTestCases() []testCase {
	validK8sConfig := &k8sAuthenticator.Config{
		Common: common.Config{
			Account:        "test-account",
			URL:            "https://conjur.example.com/authn-k8s/test-service",
			SSLCertificate: validCert,
		},
		PodName:           "test-pod",
		PodNamespace:      "test-namespace",
		InjectCertLogPath: "/tmp/test.log",
	}

	validJWTConfig := &jwtAuthenticator.Config{
		Common: common.Config{
			Account:        "test-account",
			URL:            "https://conjur.example.com/authn-jwt/test-service",
			SSLCertificate: validCert,
		},
		JWTTokenFilePath: "/var/run/secrets/tokens/jwt",
	}

	return []testCase{
		{
			name:              "Creates K8s authenticator successfully",
			config:            validK8sConfig,
			expectedError:     "",
			expectedAuthnType: reflect.TypeOf(&k8sAuthenticator.Authenticator{}),
		},
		{
			name:              "Creates JWT authenticator successfully",
			config:            validJWTConfig,
			expectedError:     "",
			expectedAuthnType: reflect.TypeOf(&jwtAuthenticator.Authenticator{}),
		},
		{
			name:              "Returns error for unsupported config type",
			config:            &unsupportedConfig{},
			expectedError:     log.CAKC064,
			expectedAuthnType: nil,
		},
		{
			name: "Returns error for K8s config with invalid SSL certificate",
			config: &k8sAuthenticator.Config{
				Common: common.Config{
					Account:        "test-account",
					URL:            "https://conjur.example.com/authn-k8s/test-service",
					SSLCertificate: []byte("invalid-cert"),
				},
				PodName:      "test-pod",
				PodNamespace: "test-namespace",
			},
			expectedError:     "CAKC014 Failed to append Conjur CA cert",
			expectedAuthnType: nil,
		},
		{
			name: "Returns error for JWT config with invalid SSL certificate",
			config: &jwtAuthenticator.Config{
				Common: common.Config{
					Account:        "test-account",
					URL:            "https://conjur.example.com/authn-jwt/test-service",
					SSLCertificate: []byte("invalid-cert"),
				},
				JWTTokenFilePath: "/var/run/secrets/tokens/jwt",
			},
			expectedError:     "Can't append Secrets Manager SSL cert",
			expectedAuthnType: nil,
		},
		{
			name: "Returns error for JWT config with malformed URL",
			config: &jwtAuthenticator.Config{
				Common: common.Config{
					Account:        "test-account",
					URL:            "://invalid-url",
					SSLCertificate: validCert,
				},
				JWTTokenFilePath: "/var/run/secrets/tokens/jwt",
			},
			expectedError:     "CAKC083 Invalid URL format",
			expectedAuthnType: nil,
		},
	}
}

// runTests executes the shared test cases against the provided factory function.
func runTests(t *testing.T, factory func(config.Configuration) (Authenticator, error)) {
	for _, tc := range buildTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			authn, err := factory(tc.config)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, authn)
			} else {
				require.NoError(t, err)
				require.NotNil(t, authn)
				assert.Equal(t, tc.expectedAuthnType, reflect.TypeOf(authn))
			}
		})
	}
}

func TestNewAuthenticatorWithAccessToken(t *testing.T) {
	runTests(t, func(conf config.Configuration) (Authenticator, error) {
		token, _ := memory.NewAccessToken()
		return NewAuthenticatorWithAccessToken(conf, token)
	})
}

func TestNewAuthenticator(t *testing.T) {
	runTests(t, func(conf config.Configuration) (Authenticator, error) {
		return NewAuthenticator(conf)
	})
}

// unsupportedConfig is a test type that doesn't match any known authenticator config
type unsupportedConfig struct{}

func (u *unsupportedConfig) LoadConfig(settings map[string]string) {}

func (u *unsupportedConfig) GetEnvVariables() []string {
	return []string{}
}

func (u *unsupportedConfig) GetRequiredVariables() []string {
	return []string{}
}

func (u *unsupportedConfig) GetDefaultValues() map[string]string {
	return map[string]string{}
}

func (u *unsupportedConfig) GetContainerMode() string {
	return "test"
}

func (u *unsupportedConfig) GetTokenFilePath() string {
	return "/tmp/token"
}

func (u *unsupportedConfig) GetTokenTimeout() time.Duration {
	return time.Minute
}
