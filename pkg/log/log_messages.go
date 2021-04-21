package log

/*
	This go file centralizes log messages (in different levels) so we have them all in one place.

	Although having the names of the consts as the error code (i.e CAKC001) and not as a descriptive name (i.e WriteAccessTokenError)
	can reduce readability of the code that raises the error, we decided to do so for the following reasons:
		1.  Improves supportability – when we get this code in the log we can find it directly in the code without going
			through the “log_messages.go” file first
		2. Validates we don’t have error code duplications – If the code is only in the string then 2 errors can have the
			same code (which is something that a developer can easily miss). However, If they are in the object name
			then the compiler will not allow it.
*/

// ERROR MESSAGES
const CAKC001 string = "CAKC001 Error opening the access token file"
const CAKC002 string = "CAKC002 Error deleting access token"
const CAKC003 string = "CAKC003 Error writing access token, reason: failed to write file"
const CAKC004 string = "CAKC004 Error writing access token, reason: failed to create directory"
const CAKC005 string = "CAKC005 Error writing access token, reason: data is empty"
const CAKC006 string = "CAKC006 Error reading access token, reason: data is empty"
const CAKC007 string = "CAKC007 At least one of CONJUR_SSL_CERTIFICATE and CONJUR_CERT_FILE must be provided"
const CAKC008 string = "CAKC008 Namespace or podname can't be empty namespace=%v podname=%v"
const CAKC009 string = "CAKC009 Environment variable '%s' must be provided"
const CAKC010 string = "CAKC010 Failed to parse %s. Reason: %s"
const CAKC011 string = "CAKC011 Client certificate not found at '%s'"
const CAKC012 string = "CAKC012 Failed to read client certificate file: %s"
const CAKC013 string = "CAKC013 Failed parsing certificate file '%s'. Reason: %s"
const CAKC014 string = "CAKC014 Failed to append Conjur CA cert"
const CAKC015 string = "CAKC015 Login failed"
const CAKC016 string = "CAKC016 Failed to authenticate"
const CAKC017 string = "CAKC017 Failed to parse key-pair from pem. Reason: %s"
const CAKC018 string = "CAKC018 Failed to instantiate authenticator configuration"
const CAKC019 string = "CAKC019 Failed to instantiate authenticator object"
const CAKC020 string = "CAKC020 Failed to parse authentication response"
const CAKC021 string = "CAKC021 Failed to read SSL Certificate. Reason: %s"
const CAKC022 string = "CAKC022 Failed to read body of authenticate HTTP response. Reason: %s"
const CAKC023 string = "CAKC023 Failed to create new authenticate HTTP request. Reason: %s"
const CAKC024 string = "CAKC024 Failed to create new login HTTP request. Reason: %s"
const CAKC025 string = "CAKC025 Failed to decode from PEM. Reason: %s"
const CAKC026 string = "CAKC026 Failed to decode a DER encoded PKCS7 package. Reason: %s"
const CAKC027 string = "CAKC027 Failed to send https authenticate request or receive response. Reason: %s"
const CAKC028 string = "CAKC028 Failed to send https login request or response. Reason: %s"
const CAKC029 string = "CAKC029 Received invalid response to certificate signing request. Reason: %s"
const CAKC030 string = "CAKC030 Failed to generate RSA keypair. Reason: %s"
const CAKC031 string = "CAKC031 Retransmission backoff exhausted"
const CAKC032 string = "CAKC032 CONJUR_AUTHN_LOGIN %s must start with 'host/'"
const CAKC033 string = "CAKC033 Timed out after waiting for %d seconds for file to exist: %s"
const CAKC034 string = "CAKC034 Incorrect value '%s' provided for enabling debug mode. Allowed value: '%s'"
const CAKC035 string = "CAKC035 Successfully authenticated"
const CAKC036 string = "CAKC036 Logged in"
const CAKC037 string = "CAKC037 Logged in. Continuing authentication"
const CAKC038 string = "CAKC038 Certificate expired. Re-logging in..."
const CAKC039 string = "CAKC039 Trying to log in to Conjur..."
const CAKC040 string = "CAKC040 Authenticating as user '%s'"
const CAKC041 string = "CAKC041 Logging in as user '%s'"
const CAKC042 string = "CAKC042 Cert expires: %v"
const CAKC043 string = "CAKC043 Current date: %v"
const CAKC044 string = "CAKC044 Buffer time:  %v"
const CAKC045 string = "CAKC045 Login request to: %s"
const CAKC046 string = "CAKC046 Authn request to: %s"
const CAKC047 string = "CAKC047 Waiting for %s to re-authenticate"
const CAKC048 string = "CAKC048 Kubernetes Authenticator Client v%s starting up..."
const CAKC049 string = "CAKC049 Loaded client certificate successfully from %s"
const CAKC050 string = "CAKC050 Deleted client certificate from memory"
const CAKC051 string = "CAKC051 Waiting for file %s to become available..."
const CAKC052 string = "CAKC052 Debug mode is enabled"
const CAKC053 string = "CAKC053 Failed to read file %s"
const CAKC054 string = "CAKC054 Failed to delete file %s"
const CAKC055 string = "CAKC055 Cert placement failed with the following error:\n%s"
const CAKC056 string = "CAKC056 File %s does not exist"
const CAKC057 string = "CAKC057 Removing file %s"
const CAKC058 string = "CAKC058 Permissions error occured when checking if file exists: %s"
const CAKC059 string = "CAKC059 Path exists but does not contain regular file: %s"
