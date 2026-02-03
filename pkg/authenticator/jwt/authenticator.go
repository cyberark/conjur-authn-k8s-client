package jwt

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"go.opentelemetry.io/otel"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-opentelemetry-tracer/pkg/trace"
)

// Authenticator contains the configuration and client
// for the authentication connection to Conjur
type Authenticator struct {
	accessToken  access_token.AccessToken
	Config       *Config
	conjurClient *conjurapi.Client
}

// NewWithAccessToken creates a new authenticator instance from a given access token
func NewWithAccessToken(config Config, accessToken access_token.AccessToken) (*Authenticator, error) {
	baseURL, serviceID, err := parseAuthnJWTURL(config.Common.URL)
	if err != nil {
		return nil, err
	}

	conjurClient, err := conjurapi.NewClient(conjurapi.Config{
		Account:           config.Common.Account,
		ApplianceURL:      baseURL,
		SSLCert:           string(config.Common.SSLCertificate),
		ServiceID:         serviceID,
		AuthnType:         "jwt",
		JWTFilePath:       config.JWTTokenFilePath,
		CredentialStorage: conjurapi.CredentialStorageNone,
		HTTPTimeout:       int(config.Common.HTTPTimeout.Seconds()),
	})
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		conjurClient: conjurClient,
		accessToken:  accessToken,
		Config:       &config,
	}, nil
}

// parseAuthnJWTURL extracts the base URL and service ID from a JWT authentication URL.
// Expected format: {scheme}://{host}[/api]/authn-jwt/{service_id}
func parseAuthnJWTURL(authnURL string) (string, string, error) {
	parsedURL, err := url.Parse(authnURL)
	if err != nil {
		return "", "", fmt.Errorf(log.CAKC083, err)
	}

	pathParts := strings.Split(parsedURL.Path, "/")
	// Remove empty parts resulting from leading, trailing, or multiple slashes
	pathParts = slices.DeleteFunc(pathParts, func(s string) bool { return s == "" })

	// Validate: must end with /authn-jwt/{service_id}
	if len(pathParts) < 2 || pathParts[len(pathParts)-2] != "authn-jwt" {
		return "", "", fmt.Errorf(log.CAKC083, "expected path to end with /authn-jwt/<service_id>")
	}

	serviceID := pathParts[len(pathParts)-1]
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	// Add back all parts of the path except for /authn-jwt/{service_id}
	// This will include "/api" for Conjur Cloud / Edge deployments, as
	// well as any other custom path segments used with a proxy.
	for _, part := range pathParts[:len(pathParts)-2] {
		baseURL += "/" + part
	}

	return baseURL, serviceID, nil
}

// GetAccessToken is getter for accessToken
func (auth *Authenticator) GetAccessToken() access_token.AccessToken {
	return auth.accessToken
}

// Authenticate sends Conjur an authenticate request and writes the response
// to the token file (after decrypting it if needed). It also manages state of
// certificates.
// @deprecated Use AuthenticateWithContext instead
func (auth *Authenticator) Authenticate() error {
	return auth.AuthenticateWithContext(context.TODO())
}

func (auth *Authenticator) AuthenticateWithContext(ctx context.Context) error {
	log.Info(log.CAKC066)

	tr := trace.NewOtelTracer(otel.Tracer("conjur-authn-k8s-client"))
	spanCtx, span := tr.Start(ctx, "Authenticate")
	defer span.End()

	authenticationResponse, err := auth.sendAuthenticationRequest(spanCtx, tr)
	if err != nil {
		span.RecordErrorAndSetStatus(err)
		span.End()
		return err
	}

	err = auth.accessToken.Write(authenticationResponse)
	if err != nil {
		span.RecordErrorAndSetStatus(err)
		span.End()
		return err
	}

	log.Info(log.CAKC035)
	return nil
}

// sendAuthenticationRequest reads the JWT token from the file system and sends
// an authentication request to the Conjur server. It also validates the response
// code before returning its body
func (auth *Authenticator) sendAuthenticationRequest(ctx context.Context, tracer trace.Tracer) ([]byte, error) {
	var authenticatingIdentity string
	_, span := tracer.Start(ctx, "Send authentication request")
	defer span.End()

	jwtToken, err := loadJWTToken(auth.Config.JWTTokenFilePath)
	if err != nil {
		span.RecordErrorAndSetStatus(err)
		return nil, err
	}

	log.Debug(log.CAKC078)
	if auth.Config.Common.Username != nil {
		authenticatingIdentity = auth.Config.Common.Username.FullUsername
		log.Debug(log.CAKC079, authenticatingIdentity)
	} else {
		log.Debug(log.CAKC080)
		authenticatingIdentity = ""
	}
	log.Debug(log.CAKC069, AuthnType)
	token, err := auth.conjurClient.JWTAuthenticate(jwtToken, authenticatingIdentity)
	if err != nil {
		wrappedErr := log.RecordedError(log.CAKC027, err)
		span.RecordErrorAndSetStatus(wrappedErr)
		return nil, wrappedErr
	}

	return token, nil
}

func loadJWTToken(path string) (string, error) {
	log.Debug(log.CAKC076, path)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(log.CAKC067, path)
	}

	log.Debug(log.CAKC077)

	return string(data), nil
}
