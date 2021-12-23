package config

import (
	"time"
)

// Configuration defines interface for Configuration of an authentication flow
type Configuration interface {
	LoadConfig(settings map[string]string)
	GetAuthenticationType() string
	GetEnvVariables() []string
	GetRequiredVariables() []string
	GetDefaultValues() map[string]string
	GetContainerMode() string
	GetTokenFilePath() string
	GetTokenTimeout() time.Duration
}
