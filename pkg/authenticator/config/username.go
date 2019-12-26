package config

import (
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// Represents the username of the host that is authenticating with Conjur.
// We separate the username into 2 parts:
//   - Suffix: includes the machine identity (e.g [namespace]/*/*)
//   - Prefix: Everything that precedes the machine identity (e.g host/path/to/policy)
// The separation above comes to support backwards compatibility of the username
// that is sent to the server. Previously, only hosts under the
// `conjur/authn-k8s/<service-id>/apps` policy branch were able to authenticate
// with Conjur, and for that to work only the suffix was sent in the CSR request.
// To let hosts from all around the policy tree to authenticate we need to send
// the full username, but we can't change the way the suffix was sent without
// breaking backwards compatibility. This is why we separate the username into
// prefix and suffix and send them separately in the CSR request.
type Username struct {
	FullUsername string
	Prefix       string
	Suffix       string
}

func NewUsername(username string) (*Username, error) {
	usernameSplit := strings.Split(username, "/")
	// Verify that the host-id starts with 'host/' and contains a 3-part machine-identity
	if len(usernameSplit) < 4 || usernameSplit[0] != "host" {
		return nil, log.RecordedError(log.CAKC032E, username)
	}

	prefix := toRequestFormat(usernameSplit[:len(usernameSplit)-3])
	suffix := toRequestFormat(usernameSplit[len(usernameSplit)-3:])

	return &Username{
		FullUsername: username,
		Prefix:       prefix,
		Suffix:       suffix,
	}, nil
}

func toRequestFormat(usernameParts []string) string {
	return strings.Join(usernameParts, ".")
}
