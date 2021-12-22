package common

import (
	"time"
)

type ConfigurationInterface interface {
	LoadConfig(settings map[string]string)
	GetAuthenticationType() string
	GetEnvVariables() []string
	GetRequiredVariables() []string
	GetDefaultValues() map[string]string
	GetContainerMode() string
	GetTokenFilePath() string
	GetTokenTimeout() time.Duration
}
