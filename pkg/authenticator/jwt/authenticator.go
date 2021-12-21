package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/access_token/file"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/authenticator/common"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-authn-k8s-client/pkg/utils"
	"net/http"
	"os"
)

// Authenticator contains the configuration and client
// for the authentication connection to Conjur
type Authenticator struct {
	client      *http.Client
	privateKey  *rsa.PrivateKey
	AccessToken access_token.AccessToken
	Config      *Config
	PublicCert  *x509.Certificate
}

// Init the authenticator struct
func (auth *Authenticator) Init(config *common.ConfigurationInterface) (common.AuthenticatorInterface, error) {
	var cfg = (*config).(*Config)
	authn, err := New(*cfg)
	if err != nil {
		return nil, fmt.Errorf(log.CAKC019)
	}

	return authn, nil
}

func (auth *Authenticator) InitWithAccessToken(config *common.ConfigurationInterface, token access_token.AccessToken) (common.AuthenticatorInterface, error) {
	log.Debug(log.CAKC058)
	var cfg = (*config).(*Config)

	return NewWithAccessToken(*cfg, token)
}

// New creates a new authenticator instance from a token file
func New(config Config) (*Authenticator, error) {
	accessToken, err := file.NewAccessToken(config.GetTokenFilePath())
	if err != nil {
		return nil, log.RecordedError(log.CAKC001)
	}

	return NewWithAccessToken(config, accessToken)
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
		AccessToken: accessToken,
		Config:      &config,
	}, nil
}

// Authenticate sends Conjur an authenticate request and writes the response
// to the token file (after decrypting it if needed). It also manages state of
// certificates.
func (auth *Authenticator) Authenticate() error {
	log.Info(log.CAKC066)

	jwtToken, err := loadJWTToken(auth.Config.JWTTokenFilePath)

	if err != nil {
		return err
	}

	authenticationResponse, err := auth.sendAuthenticationRequest(jwtToken)
	if err != nil {
		return err
	}

	err = auth.AccessToken.Write(authenticationResponse)
	if err != nil {
		return err
	}

	log.Info(log.CAKC035)
	return nil
}

func loadJWTToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(log.CAKC067, path)
	}

	return string(data), nil
}

func (authn *Authenticator) GetAccessToken() access_token.AccessToken {
	return authn.AccessToken
}

// sendAuthenticationRequest reads the cert from memory and uses it to send
// an authentication request to the Conjur server. It also validates the response
// code before returning its body
func (auth *Authenticator) sendAuthenticationRequest(jwtToken string) ([]byte, error) {
	var authenticatingIdentity string
	if auth.Config.Common.Username != nil {
		authenticatingIdentity = auth.Config.Common.Username.FullUsername
	} else {
		authenticatingIdentity = ""
	}
	req, err := AuthenticateRequest(
		auth.Config.Common.URL,
		auth.Config.Common.Account,
		authenticatingIdentity,
		jwtToken,
	)
	if err != nil {
		return nil, err
	}

	resp, err := auth.client.Do(req)
	if err != nil {
		return nil, log.RecordedError(log.CAKC027, err)
	}

	err = utils.ValidateResponse(resp)
	if err != nil {
		return nil, err
	}

	return utils.ReadResponseBody(resp)
}
