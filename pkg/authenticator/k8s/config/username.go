package config

import (
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
)

// Represents the username of the host that is authenticating with Conjur.
// We separate the username into 2 parts:
//   - Suffix: includes the host id
//   - Prefix: includes the policy id (and the "host/" prefix)
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

// String is used to format the username to only the needed user-visible information
// rather than all the fields
func (username Username) String() string {
	return username.FullUsername
}

func NewUsername(username string) (*Username, error) {
	usernameSplit := strings.Split(username, "/")
	// Verify that the host-id starts with 'host/'
	if usernameSplit[0] != "host" {
		return nil, log.RecordedError(log.CAKC032, username)
	}

	separator := getIdSeparator(len(usernameSplit))

	prefix := toRequestFormat(usernameSplit[:len(usernameSplit)-separator])
	suffix := toRequestFormat(usernameSplit[len(usernameSplit)-separator:])

	return &Username{
		FullUsername: username,
		Prefix:       prefix,
		Suffix:       suffix,
	}, nil
}

func toRequestFormat(usernameParts []string) string {
	return strings.Join(usernameParts, ".")
}

// Return an index that will be the separator between the prefix and suffix of
// the username.
// By default, the suffix includes only the last part of the id, which indicates
// the actual host id, while the prefix is the policy id.
//
// To maintain backwards compatibility with an old Conjur server, we need to support
// a suffix that includes the 3-part application identity (and as mentioned
// earlier, the suffix should include the host-id which in this case is the 3-part
// application identity). In such a case, we take the last 3 parts of the host
// id as the suffix so it will have the application identity as the id.
//
// Note: In case the host id's length is higher than 4 but does not have the
// application identity in the id (e.g host/long/policy/name/<host_id>),
// the suffix will have the last 3 parts that may not be the application identity.
// This may look weird (prefix = host/long, suffix = policy/name/<host_id>)
// but although the suffix doesn't include the host id, it will work with a new
// Conjur server as it will concatenate the parts.
// It won't work with an old Conjur server but it shouldn't anyway.
func getIdSeparator(usernameLen int) int {
	separator := 1
	if usernameLen >= 4 {
		separator = 3
	}

	return separator
}
