package authenticator

import (
	"crypto/fips140"
	_ "errors"
	"testing"
)

func TestFIPS140Active(t *testing.T) {
	if !fips140.Enabled() {
		t.Fatal("The environment is not running in FIPS140-3 mode")
	}
}
