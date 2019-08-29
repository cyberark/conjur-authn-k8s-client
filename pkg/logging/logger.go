package logging

import (
"fmt"
"log"
"os"
)

var InfoLogger = log.New(os.Stdout, "INFO: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLogger = log.New(os.Stderr, "ERROR: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)


func PrintAndReturnError(errorMessage string, err error, chainErrorToMessage bool) error {
	if chainErrorToMessage == true {
		ErrorLogger.Printf("%s, reason: %s", errorMessage, err.Error())
	} else {
		ErrorLogger.Printf(errorMessage)
	}

	if err == nil {
		return fmt.Errorf(errorMessage)
	} else {
		return fmt.Errorf("%s, reason: %s", errorMessage, err.Error())
	}
}

// ERROR MESSAGES
const CAKC001E string = "CAKC001E Error creating secret handler object"
const CAKC002E string = "CAKC002E Error creating access token object"
const CAKC003E string = "CAKC003E Error creating secrets config"
const CAKC004E string = "CAKC004E Error creating access token object"
const CAKC005E string = "CAKC005E Store type %s is invalid"
const CAKC006E string = "CAKC006E Error deleting access token"
const CAKC007E string = "CAKC007E Error writing access token, reason: failed to write file"
const CAKC008E string = "CAKC008E Error writing access token, reason: failed to create directory"
const CAKC009E string = "CAKC009E Error writing access token, reason: data is empty"
const CAKC010E string = "CAKC010E Error reading access token, reason: data is empty"
const CAKC011E string = "CAKC011E At least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided"
const CAKC012E string = "CAKC012E Namespace or podname can't be empty namespace=%v podname=%v"
const CAKC013E string = "CAKC013E Client certificate not found at %s"
const CAKC014E string = "CAKC014E Failed reading client certificate file: %s"
const CAKC015E string = "CAKC015E Failed parsing certificate: %s"
const CAKC016E string = "CAKC016E Login failed"
const CAKC017E string = "CAKC017E Environment variable %s must be provided"
const CAKC018E string = "CAKC018E Failed to load Conjur config"
const CAKC019E string = "CAKC019E Failed to create Conjur client from token"
const CAKC020E string = "CAKC020E Error creating Conjur secrets provider"
const CAKC021E string = "CAKC021E Error retrieving Conjur secrets"
const CAKC022E string = "CAKC022E Failed to create k8s secrets handler"
const CAKC023E string = "CAKC023E Failure retrieving k8s secretsHandlerK8sUseCase"
const CAKC024E string = "CAKC024E Failure retrieving access token"
const CAKC025E string = "CAKC025E Error parsing Conjur variable ids"
const CAKC026E string = "CAKC026E Error retrieving Conjur k8sSecretsHandler"
const CAKC027E string = "CAKC027E Failure updating K8s K8sSecretsHandler map"
const CAKC028E string = "CAKC028E Failure patching K8s K8sSecretsHandler"
const CAKC029E string = "CAKC029E Error map should not be empty"
const CAKC030E string = "CAKC030E Failed to update k8s k8sSecretsHandler map"
const CAKC031E string = "CAKC031E Failed to parse Conjur variable ID: %s"
const CAKC032E string = "CAKC032E Error reading k8s secrets"
const CAKC033E string = "CAKC033E Failed to patch k8s secret"
const CAKC034E string = "CAKC034E Failed to load in-cluster Kubernetes client config"
const CAKC035E string = "CAKC035E Failed to create Kubernetes client"
const CAKC036E string = "CAKC036E Failed to retrieve Kubernetes secret"
const CAKC037E string = "CAKC037E Failed to parse Kubernetes secret list"
const CAKC038E string = "CAKC038E Failed to patch Kubernetes secret"
const CAKC039E string = "CAKC039E Data entry map cannot be empty"
const CAKC040E string = "CAKC040E Failure to append Conjur CA cert"
const CAKC041E string = "CAKC041E Failed parse key-pair from pem"
const CAKC042E string = "CAKC042E Provided incorrect value for environment variable %s"
const CAKC043E string = "CAKC043E Environment variable %s must be provided"
const CAKC044E string = "CAKC044E Failed to parse CONJUR_TOKEN_TIMEOUT"
const CAKC045E string = "CAKC045E Failed to instantiate authenticator configuration"
const CAKC046E string = "CAKC046E Failed to instantiate storage configuration"
const CAKC047E string = "CAKC047E Setting SECRETS_DESTINATION environment variable to 'k8s_secrets' must run as init container"
const CAKC048E string = "CAKC048E Failed to instantiate storage handler"
const CAKC049E string = "CAKC049E Failed to instantiate authenticator object"
const CAKC050E string = "CAKC050E Failure authenticating"
const CAKC051E string = "CAKC051E Failure parsing authentication response"
const CAKC052E string = "CAKC052E Failed to handle secrets"
const CAKC053E string = "CAKC053E Backoff exhausted"
const CAKC054E string = "CAKC054E Failed to delete access token"


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
const CAKC017I string = "CAKC017I Patching Kubernetes secret '%s' in namespace '%s'..."
const CAKC018I string = "CAKC018I Retrieving following secrets from Conjur: "
const CAKC019I string = "CAKC019I Authenticating as user '%s'"


