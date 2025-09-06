package application

import (
	"errors"
	"os"
)

var (
	// ErrEnvVarNotSet is returned when an expected environment variable is not set.
	ErrEnvVarNotSet = errors.New("environment variable not set")
)

// EnvProvider defines an interface for retrieving environment variables.
type EnvProvider interface {
	GetEnv(key string) (string, error)
}

// DefaultEnvProvider is the default implementation using os.Getenv.
type DefaultEnvProvider struct{}

// GetEnv returns the value of an env var or an error if not set.
func (p *DefaultEnvProvider) GetEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", ErrEnvVarNotSet
	}
	return value, nil
}
