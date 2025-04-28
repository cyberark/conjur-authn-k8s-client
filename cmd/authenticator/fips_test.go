package main

import (
	"crypto/fips140"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFIPS140Enabled(t *testing.T) {
	assert.True(t, fips140.Enabled(), "FIPS 140-3 mode is not enabled")
}
