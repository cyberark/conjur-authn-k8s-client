package logging

import (
	"fmt"
	"log"
	"os"
)

var InfoLogger = log.New(os.Stdout, "INFO:  ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLogger = log.New(os.Stderr, "ERROR: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)

func PrintAndReturnError(errorMessage string, args ...interface{}) error {
	ErrorLogger.Output(2, fmt.Sprintf(errorMessage, args...))
	return fmt.Errorf(fmt.Sprintf(errorMessage, args...))
}

// ERROR MESSAGES
const CAKC001E string = "CAKC001E Error creating the secret handler object"
const CAKC002E string = "CAKC002E Error creating the access token object"
const CAKC003E string = "CAKC003E Error creating the secrets config"
const CAKC004E string = "CAKC004E Error creating access token object"
const CAKC005E string = "CAKC005E Store type %s is invalid"
const CAKC006E string = "CAKC006E Error deleting access token"
const CAKC007E string = "CAKC007E Error writing access token, reason: failed to write file"
const CAKC008E string = "CAKC008E Error writing access token, reason: failed to create directory"
const CAKC009E string = "CAKC009E Error writing access token, reason: data is empty"
const CAKC010E string = "CAKC010E Error reading access token, reason: data is empty"
const CAKC011E string = "CAKC011E At least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided"
const CAKC012E string = "CAKC012E Namespace or podname can't be empty namespace=%v podname=%v"
const CAKC013E string = "CAKC013E Client certificate not found at '%s'"
const CAKC014E string = "CAKC014E Failed to read client certificate file: %s"
const CAKC015E string = "CAKC015E Failed parsing certificate file '%s'. Reason: %s"
const CAKC016E string = "CAKC016E Login failed"
const CAKC017E string = "CAKC017E Environment variable '%s' must be provided"
const CAKC018E string = "CAKC018E Failed to load Conjur config. Reason: %s"
const CAKC019E string = "CAKC019E Failed to create Conjur client from token. Reason: %s"
const CAKC020E string = "CAKC020E Error creating Conjur secrets provider"
const CAKC021E string = "CAKC021E Error retrieving Conjur secrets. Reason: %s"
const CAKC022E string = "CAKC022E Failed to create k8s secrets handler"
const CAKC023E string = "CAKC023E Failure retrieving k8s secretsHandlerK8sUseCase"
const CAKC024E string = "CAKC024E Failed to retrieve access token"
const CAKC025E string = "CAKC025E Error parsing Conjur variable ids"
const CAKC026E string = "CAKC026E Error retrieving Conjur k8sSecretsHandler"
const CAKC027E string = "CAKC027E Failed to update K8s K8sSecretsHandler map"
const CAKC028E string = "CAKC028E Failed to patch K8s K8sSecretsHandler"
const CAKC029E string = "CAKC029E Error map should not be empty"
const CAKC030E string = "CAKC030E k8s secret '%s' has no value defined for the '%s' data entry"
const CAKC031E string = "CAKC031E Failed to parse Conjur variable ID: %s"
const CAKC032E string = "CAKC032E Error reading k8s secrets"
const CAKC033E string = "CAKC033E Failed to patch k8s secret"
const CAKC034E string = "CAKC034E Failed to load in-cluster Kubernetes client config. Reason: %s"
const CAKC035E string = "CAKC035E Failed to create Kubernetes client. Reason: %s"
const CAKC036E string = "CAKC036E Failed to retrieve Kubernetes secret. Reason: %s"
const CAKC037E string = "CAKC037E Failed to parse Kubernetes secret list"
const CAKC038E string = "CAKC038E Failed to patch Kubernetes secret. Reason: %s"
const CAKC039E string = "CAKC039E Data entry map cannot be empty"
const CAKC040E string = "CAKC040E Failed to append Conjur CA cert"
const CAKC041E string = "CAKC041E Failed parse key-pair from pem. Reason: %s"
const CAKC042E string = "CAKC042E Provided incorrect value for environment variable %s"
const CAKC043E string = "CAKC043E k8s secret '%s' has an invalid value for '%s' data entry"
const CAKC044E string = "CAKC044E Failed to parse CONJUR_TOKEN_TIMEOUT. Reason: %s"
const CAKC045E string = "CAKC045E Failed to instantiate authenticator configuration"
const CAKC046E string = "CAKC046E Failed to instantiate storage configuration"
const CAKC047E string = "CAKC047E Setting SECRETS_DESTINATION environment variable to 'k8s_secrets' must run as init container"
const CAKC048E string = "CAKC048E Failed to instantiate storage handler"
const CAKC049E string = "CAKC049E Failed to instantiate authenticator object"
const CAKC050E string = "CAKC050E Failure authenticating"
const CAKC051E string = "CAKC051E Failed to parse authentication response"
const CAKC052E string = "CAKC052E Failed to handle secrets"
const CAKC053E string = "CAKC053E Retransmission backoff exhausted"
const CAKC054E string = "CAKC054E Failed to delete access token"
const CAKC055E string = "CAKC055E Failed to read SSL Certificate. Reason: %s"
const CAKC056E string = "CAKC056E Failed to read body of authenticate HTTP response. Reason: %s"
const CAKC057E string = "CAKC057E Failed to create new authenticate HTTP request. Reason: %s"
const CAKC058E string = "CAKC058E Failed to create new login HTTP request. Reason: %s"
const CAKC059E string = "CAKC059E Failed to decode from PEM. Reason: %s"
const CAKC060E string = "CAKC060E Failed decoding a DER encoded PKCS7 package. Reason: %s"
const CAKC061E string = "CAKC061E Failed to send https authenticate request or receive response. Reason: %s"
const CAKC062E string = "CAKC062E Failed to send https login request or response. Reason: %s"
const CAKC063E string = "CAKC063E Received invalid response to certificate signing request. Reason: %s"
const CAKC064E string = "CAKC064E Failed to generate RSA keypair. Reason: %s"
const CAKC065E string = "CAKC065E AccessTokenHandler failed to delete access token. Reason: %s"
const CAKC066E string = "CAKC066E Failed to find any k8s secrets defined with a '%s’ data entry"

// INFO MESSAGES
const CAKC001I string = "CAKC001I Storage configuration is %s"
const CAKC002I string = "CAKC002I Successfully authenticated"
const CAKC003I string = "CAKC003I Logged in. Continuing authentication"
const CAKC004I string = "CAKC004I Certificate expired. Re-logging in..."
const CAKC005I string = "CAKC005I Trying to login Conjur..."
const CAKC006I string = "CAKC006I Logging in as %s."
const CAKC007I string = "CAKC007I Cert expires: %v"
const CAKC008I string = "CAKC008I Current date: %v"
const CAKC009I string = "CAKC009I Buffer time:  %v"
const CAKC010I string = "CAKC010I Logged in"
const CAKC011I string = "CAKC011I Login request to: %s"
const CAKC012I string = "CAKC012I Authn request to: %s"
const CAKC013I string = "CAKC013I Waiting for %s to re-authenticate and fetch secrets."
const CAKC014I string = "CAKC014I Creating Kubernetes client..."
const CAKC015I string = "CAKC015I Creating Conjur client..."
const CAKC016I string = "CAKC016I Retrieving Kubernetes secret '%s' from namespace '%s'..."
const CAKC017I string = "CAKC017I Patching Kubernetes secret '%s' in namespace '%s'"
const CAKC018I string = "CAKC018I Retrieving following secrets from Conjur: "
const CAKC019I string = "CAKC019I Authenticating as user '%s'"
