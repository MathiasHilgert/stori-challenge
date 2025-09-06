package application

import (
	"context"
	"errors"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var (
	// ErrSecretNotFound is returned when a secret cannot be found.
	ErrSecretNotFound = errors.New("secret not found")

	// ErrSecretTypeMismatch is returned when a secret cannot be converted to the requested type.
	ErrSecretTypeMismatch = errors.New("secret type mismatch")
)

// SecretsProvider defines an interface for retrieving secrets.
type SecretsProvider interface {
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
}

// AWSSecretsProvider uses AWS Secrets Manager.
type AWSSecretsProvider struct {
	client *secretsmanager.Client
}

// NewAWSSecretsProvider constructs an AWSSecretsProvider.
func NewAWSSecretsProvider(client *secretsmanager.Client) *AWSSecretsProvider {
	return &AWSSecretsProvider{client: client}
}

// GetString retrieves a secret as string.
func (p *AWSSecretsProvider) GetString(ctx context.Context, key string) (string, error) {
	secret, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &key,
	})
	if err != nil || secret.SecretString == nil {
		return "", ErrSecretNotFound
	}
	return *secret.SecretString, nil
}

// GetInt retrieves a secret as int.
func (p *AWSSecretsProvider) GetInt(ctx context.Context, key string) (int, error) {
	val, err := p.GetString(ctx, key)
	if err != nil {
		return 0, err
	}
	intVal, convErr := strconv.Atoi(val)
	if convErr != nil {
		return 0, ErrSecretTypeMismatch
	}
	return intVal, nil
}
