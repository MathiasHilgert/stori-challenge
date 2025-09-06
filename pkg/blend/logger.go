package blend

import (
	"context"
	"io"
)

// Default returns a default Logger implementation for the given output writer.
//
// It creates and returns an instance of ZerologLogger, which is a structure
// that implements the Logger interface, using the zerolog library as the
// underlying logging implementation.
//
// See github.com/rs/zerolog for more information.
func Default(output io.Writer) (logger Logger, err error) {
	return NewZerologLogger(output)
}

// Logger is an interface that exposes different methods for logging messages
// in a variety of different levels.
type Logger interface {
	// Debug logs a message at the debug level.
	// The debug level is used for debugging purposes, and is usually disabled
	// in production environments. It is used to log messages that are only
	// relevant to developers.
	//
	// This method accepts a message string and a variadic number of arguments,
	// which will be used to format the message string (similar to fmt.Printf verbs).
	// See https://pkg.go.dev/fmt#hdr-Printing for more information.
	Debug(ctx context.Context, message string, args ...any) error

	// Info logs a message at the info level.
	// The info level is used for logging general information about the application,
	// such as startup messages, shutdown messages, etc.
	//
	// This method accepts a message string and a variadic number of arguments,
	// which will be used to format the message string (similar to fmt.Printf verbs).
	// See https://pkg.go.dev/fmt#hdr-Printing for more information.
	Info(ctx context.Context, message string, args ...any) error

	// Warn logs a message at the warn level.
	// The warn level is used for logging messages that are not critical, but
	// might be worth investigating. For example, a warning message might be
	// logged when a request to an external service fails, but the application
	// is able to recover from it.
	//
	// This method accepts a message string and a variadic number of arguments,
	// which will be used to format the message string (similar to fmt.Printf verbs).
	// See https://pkg.go.dev/fmt#hdr-Printing for more information.
	Warn(ctx context.Context, message string, args ...any) error

	// Error logs a message at the error level.
	// The error level is used for logging messages that are critical, and
	// require immediate attention. For example, an error message might be
	// logged when a request to an external service fails, and the application
	// is not able to recover from it.
	//
	// This method accepts a message string and a variadic number of arguments,
	// which will be used to format the message string (similar to fmt.Printf verbs).
	// See https://pkg.go.dev/fmt#hdr-Printing for more information.
	Error(ctx context.Context, message string, args ...any) error

	// Fatal logs a message at the fatal level.
	// The fatal level is used for logging messages that are critical, and
	// require immediate attention. For example, a fatal message might be
	// logged when the application is not able to connect to the database.
	//
	// Be careful when using this method, as it will terminate the application
	// after logging the message. It's recommended only on main startup functions.
	//
	// If you are a newbie, get away from this method as far as possible ヾ(￣▽￣) Bye~Bye~
	//
	// This method accepts a message string and a variadic number of arguments,
	// which will be used to format the message string (similar to fmt.Printf verbs).
	// See https://pkg.go.dev/fmt#hdr-Printing for more information.
	Fatal(ctx context.Context, message string, args ...any) error
}

// DummyLogger is a dummy implementation of the Logger interface.
// It does not log anything, and is used for testing purposes.
//
// It's dummy like me ( ͡❛ ͜ʖ ͡❛)
type DummyLogger struct{}

// NewDummyLogger creates a new DummyLogger instance.
// It does not log anything, and is used for testing purposes.
func NewDummyLogger() *DummyLogger {
	return &DummyLogger{}
}

func (logger *DummyLogger) Debug(ctx context.Context, message string, args ...any) error {
	return nil
}

func (logger *DummyLogger) Info(ctx context.Context, message string, args ...any) error {
	return nil
}

func (logger *DummyLogger) Warn(ctx context.Context, message string, args ...any) error {
	return nil
}

func (logger *DummyLogger) Error(ctx context.Context, message string, args ...any) error {
	return nil
}

func (logger *DummyLogger) Fatal(ctx context.Context, message string, args ...any) error {
	return nil
}
