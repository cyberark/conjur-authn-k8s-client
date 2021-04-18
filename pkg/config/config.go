package config

import "time"

// Config global configs used by the authenticator
type Config struct {
	TokenRefreshTimeout time.Duration
	ContainerMode       string
}
