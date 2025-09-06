package application

import (
	"context"
)

// TransactionsDynamoDBConfig holds the configuration details for connecting to DynamoDB.
type TransactionsDynamoDBConfig struct {
	// TableName is the name of the DynamoDB table.
	TableName string
}

// SMTPConfig holds the configuration details for connecting to an SMTP server.
type SMTPConfig struct {
	// Host is the SMTP server host.
	Host string

	// Port is the SMTP server port.
	Port int

	// Username is the SMTP server username.
	Username string

	// Password is the SMTP server password.
	Password string

	// From is the default "from" email address.
	From string
}

// ApplicationConfig holds the configuration for the application.
type ApplicationConfig struct {
	// TransactionsDynamoDB holds the configuration for the transactions DynamoDB.
	TransactionsDynamoDB TransactionsDynamoDBConfig

	// EmailSMTP holds the configuration for the SMTP server used for sending emails.
	EmailSMTP SMTPConfig
}

// Load loads application configuration from providers.
// Returns an error if any critical value is missing.
func (config *ApplicationConfig) Load(ctx context.Context, env EnvProvider, secrets SecretsProvider) error {
	// DynamoDB table name (must be present)
	table, err := env.GetEnv("DYNAMODB_TABLE_NAME")
	if err != nil {
		return err
	}

	// SMTP configuration (all must be present)
	host, err := secrets.GetString(ctx, "SMTP_HOST")
	if err != nil {
		return err
	}
	port, err := secrets.GetInt(ctx, "SMTP_PORT")
	if err != nil {
		return err
	}
	user, err := secrets.GetString(ctx, "SMTP_USERNAME")
	if err != nil {
		return err
	}
	pass, err := secrets.GetString(ctx, "SMTP_PASSWORD")
	if err != nil {
		return err
	}
	from, err := secrets.GetString(ctx, "SMTP_FROM")
	if err != nil {
		return err
	}

	// Assign to config
	config.TransactionsDynamoDB = TransactionsDynamoDBConfig{TableName: table}
	config.EmailSMTP = SMTPConfig{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
	}
	return nil
}
