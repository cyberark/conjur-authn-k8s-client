package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginRequest(t *testing.T) {
	// SETUP
	conjurVersion := "5"
	authnURL := "dummyURL"
	csrBytes := []byte("dummyCSRBytes")

	// EXERCISE
	req, err := LoginRequest(authnURL, conjurVersion, csrBytes, "host.path.to.policy")
	if !assert.NoError(t, err) {
		return
	}

	// ASSERT
	assert.Equal(t, "host.path.to.policy", req.Header.Get("Host-Id-Prefix"))
}
