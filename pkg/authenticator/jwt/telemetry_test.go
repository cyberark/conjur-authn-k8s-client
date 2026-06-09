package jwt

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const conjurTelemetryHeader = "x-cybr-telemetry"

func decodeTelemetry(t *testing.T, encoded string) string {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(encoded)
	require.NoError(t, err, "telemetry header must be valid base64url")
	return string(b)
}

func TestTelemetryData_FieldsAreSet(t *testing.T) {
	assert.Equal(t, "Authn K8s Client", telemetryData.IntegrationName)
	assert.Equal(t, "idira-secretsmanager", telemetryData.IntegrationType)
	assert.Equal(t, "Idira", telemetryData.VendorName)
}

func TestTelemetryData_VersionMatchesPackageVars(t *testing.T) {
	expected := "v" + Version + TagSuffix
	assert.Equal(t, expected, telemetryData.IntegrationVersion,
		"telemetryData version must be composed from Version and TagSuffix ldflag variables")
}

func TestTelemetryData_EncodesExpectedFields(t *testing.T) {
	raw := "in=" + telemetryData.IntegrationName +
		"&it=" + telemetryData.IntegrationType +
		"&vn=" + telemetryData.VendorName
	encoded := base64.RawURLEncoding.EncodeToString([]byte(raw))
	decoded := decodeTelemetry(t, encoded)

	assert.True(t, strings.Contains(decoded, "in=Authn K8s Client"))
	assert.True(t, strings.Contains(decoded, "it=idira-secretsmanager"))
	assert.True(t, strings.Contains(decoded, "vn=Idira"))
}
