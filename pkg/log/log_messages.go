package log

/*
	This go file centralizes log messages (in different levels) so we have them all in one place.

	Although having the names of the consts as the error code (i.e CAKC001E) and not as a descriptive name (i.e WriteAccessTokenError)
	can reduce readability of the code that raises the error, we decided to do so for the following reasons:
		1.  Improves supportability – when we get this code in the log we can find it directly in the code without going
			through the “log_messages.go” file first
		2. Validates we don’t have error code duplications – If the code is only in the string then 2 errors can have the
			same code (which is something that a developer can easily miss). However, If they are in the object name
			then the compiler will not allow it.
*/

// ERROR MESSAGES
const CAKC001E string = "CAKC001E Error creating the access token object"
const CAKC002E string = "CAKC002E Error deleting access token"
const CAKC003E string = "CAKC003E Error writing access token, reason: failed to write file"
const CAKC004E string = "CAKC004E Error writing access token, reason: failed to create directory"
const CAKC005E string = "CAKC005E Error writing access token, reason: data is empty"
const CAKC006E string = "CAKC006E Error reading access token, reason: data is empty"
const CAKC007E string = "CAKC007E At least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided"
const CAKC008E string = "CAKC008E Namespace or podname can't be empty namespace=%v podname=%v"
const CAKC009E string = "CAKC009E Environment variable '%s' must be provided"
const CAKC010E string = "CAKC010E Failed to parse %s. Reason: %s"
const CAKC011E string = "CAKC011E Client certificate not found at '%s'"
const CAKC012E string = "CAKC012E Failed to read client certificate file: %s"
const CAKC013E string = "CAKC013E Failed parsing certificate file '%s'. Reason: %s"
const CAKC014E string = "CAKC014E Failed to to append Conjur CA cert"
const CAKC015E string = "CAKC015E Login failed"
const CAKC016E string = "CAKC016E Failed to authenticate"
const CAKC017E string = "CAKC017E Failed to parse key-pair from pem. Reason: %s"
const CAKC018E string = "CAKC018E Failed to instantiate authenticator configuration"
const CAKC019E string = "CAKC019E Failed to instantiate authenticator object"
const CAKC020E string = "CAKC020E Failed to parse authentication response"
const CAKC021E string = "CAKC021E Failed to read SSL Certificate. Reason: %s"
const CAKC022E string = "CAKC022E Failed to read body of authenticate HTTP response. Reason: %s"
const CAKC023E string = "CAKC023E Failed to create new authenticate HTTP request. Reason: %s"
const CAKC024E string = "CAKC024E Failed to create new login HTTP request. Reason: %s"
const CAKC025E string = "CAKC025E Failed to decode from PEM. Reason: %s"
const CAKC026E string = "CAKC026E Failed to decode a DER encoded PKCS7 package. Reason: %s"
const CAKC027E string = "CAKC027E Failed to send https authenticate request or receive response. Reason: %s"
const CAKC028E string = "CAKC028E Failed to send https login request or response. Reason: %s"
const CAKC029E string = "CAKC029E Received invalid response to certificate signing request. Reason: %s"
const CAKC030E string = "CAKC030E Failed to generate RSA keypair. Reason: %s"
const CAKC031E string = "CAKC031E Retransmission backoff exhausted"
const CAKC032E string = "CAKC032E Username %s is invalid"
const CAKC033E string = "CAKC033E Timed out after waiting for %d seconds for file to exist: %s"

// INFO MESSAGES
const CAKC001I string = "CAKC001I Successfully authenticated"
const CAKC002I string = "CAKC002I Logged in"
const CAKC003I string = "CAKC003I Logged in. Continuing authentication"
const CAKC004I string = "CAKC004I Certificate expired. Re-logging in..."
const CAKC005I string = "CAKC005I Trying to login Conjur..."
const CAKC006I string = "CAKC006I Authenticating as user '%s'"
const CAKC007I string = "CAKC007I Logging in as user '%s'"
const CAKC008I string = "CAKC008I Cert expires: %v"
const CAKC009I string = "CAKC009I Current date: %v"
const CAKC010I string = "CAKC010I Buffer time:  %v"
const CAKC011I string = "CAKC011I Login request to: %s"
const CAKC012I string = "CAKC012I Authn request to: %s"
const CAKC013I string = "CAKC013I Waiting for %s to re-authenticate"
const CAKC014I string = "CAKC014I Kubernetes Authenticator Client v%s starting up..."
const CAKC015I string = "CAKC015I Loaded client certificate successfully from %s"
const CAKC016I string = "CAKC016I Deleted client certificate from memory"
const CAKC017I string = "CAKC017I Waiting for file %s to become available..."
