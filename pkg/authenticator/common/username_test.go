package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUsername(t *testing.T) {
	t.Run("longer than 4 parts", func(t *testing.T) {
		// SETUP & EXERCISE
		usernameStruct, err := NewUsername(
			"host/path/to/policy/namespace/resource_type/resource_id",
		)
		if !assert.NoError(t, err) {
			return
		}

		// ASSERT
		assert.Equal(t, "host.path.to.policy", usernameStruct.Prefix)
		assert.Equal(t, "namespace.resource_type.resource_id", usernameStruct.Suffix)
	})

	t.Run("shorter than 4 parts", func(t *testing.T) {
		// SETUP & EXERCISE
		usernameStruct, err := NewUsername("host/policy/host_id")
		if !assert.NoError(t, err) {
			return
		}

		// ASSERT
		assert.Equal(t, "host.policy", usernameStruct.Prefix)
		assert.Equal(t, "host_id", usernameStruct.Suffix)
	})

	t.Run("missing host/ prefix", func(t *testing.T) {
		// SETUP & EXERCISE
		_, err := NewUsername("namespace/resource_type/resource_id")
		if !assert.Error(t, err) {
			return
		}

		assert.Contains(t, err.Error(), "CAKC032")
	})

	t.Run("string representation", func(t *testing.T) {
		// SETUP & EXERCISE
		usernameStruct, err := NewUsername(
			"host/path/to/policy/namespace/resource_type/resource_id",
		)
		if !assert.NoError(t, err) {
			return
		}

		// ASSERT
		assert.Equal(
			t,
			"host/path/to/policy/namespace/resource_type/resource_id",
			fmt.Sprintf("%s", usernameStruct),
		)
	})
}
