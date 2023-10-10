package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/utils"
	"github.com/cyberark/conjur-opentelemetry-tracer/pkg/trace"
)

// Authenticator contains the configuration and client
// for the authentication connection to Conjur
type Authenticator struct {
	client      *http.Client
	privateKey  *rsa.PrivateKey
	accessToken access_token.AccessToken
	Config      *Config
}

// NewWithAccessToken creates a new authenticator instance from a given access token
func NewWithAccessToken(config Config, accessToken access_token.AccessToken) (*Authenticator, error) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, log.RecordedError(log.CAKC030, err)
	}

	client, err := common.NewHTTPSClient(config.Common.SSLCertificate, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		client:      client,
		privateKey:  signingKey,
		accessToken: accessToken,
		Config:      &config,
	}, nil
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
	ctx, tracer, cleanup, err := trace.Create(
		trace.NoopProviderType,
		trace.TracerProviderConfig{},
	)
	if err != nil {
		return err
	}
	defer cleanup(ctx)

	return auth.AuthenticateWithContext(ctx, tracer)
}

func (auth *Authenticator) AuthenticateWithContext(ctx context.Context, tr trace.Tracer) error {
	log.Info(log.CAKC066)

	ctx, span := tr.Start(ctx, "Authenticate")
	defer span.End()

	authenticationResponse, err := auth.sendAuthenticationRequest(ctx, tr)
	if err != nil {
		span.RecordErrorAndSetStatus(err)
		return err
	}

	err = auth.accessToken.Write(authenticationResponse)
	if err != nil {
		span.RecordErrorAndSetStatus(err)
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

	req, err := AuthenticateRequest(
		auth.Config.Common.URL,
		auth.Config.Common.Account,
		authenticatingIdentity,
		jwtToken,
	)

	if err != nil {
		span.RecordErrorAndSetStatus(err)
		return nil, err
	}

	log.Debug(log.CAKC069, AuthnType)
	resp, err := auth.client.Do(req)

	if err != nil {
		span.RecordErrorAndSetStatus(err)
		return nil, log.RecordedError(log.CAKC027, err)
	}

	err = utils.ValidateResponse(resp)
	if err != nil {
		span.RecordErrorAndSetStatus(err)
		return nil, err
	}

	return utils.ReadResponseBody(resp)
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
