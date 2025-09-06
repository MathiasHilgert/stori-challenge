package blend

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ZerologLogger is a structure that implements the Logger interface, using
// the zerolog library as the underlying logging implementation.
//
// See github.com/rs/zerolog for more information.
type ZerologLogger struct {
	// engine is the underlying logging implementation, which is an instance
	// of zerolog.Logger.
	engine *zerolog.Logger

	// output is the output writer that will be used by the underlying
	// logging implementation.
	output io.Writer

	// now is a function that returns the current time.
	// This is useful for testing purposes, as it allows us to mock the
	// current time.
	now func() time.Time
}

// NewZerologLogger returns a new instance of ZerologLogger.
func NewZerologLogger(output io.Writer) (logger *ZerologLogger, err error) {
	// Check if the output writer is nil.
	if output == nil {
		err = io.ErrClosedPipe
		return
	}

	// Create a new instance of zerolog.Logger.
	engine := zerolog.New(output)

	// Create a new instance of ZerologLogger.
	logger = &ZerologLogger{
		engine: &engine,
		output: output,
		now:    time.Now,
	}
	return

}

// log is an internal method that logs a message at the specified level.
func (logger *ZerologLogger) log(ctx context.Context, level Level, message string, args ...any) (err error) {
	// If the logger engine or the output writer are not set, return an error.
	if logger.engine == nil || logger.output == nil {
		err = io.ErrClosedPipe
		return
	}

	// Nice, we can log the message.
	leveledLogger := logger.engine.WithLevel(zerolog.NoLevel)
	leveledLogger.
		Time("time", logger.now()).
		Str("level", level.String()).
		Msgf(message, args...)
	return
}

// Debug logs a message at the debug level.
// The debug level is used for debugging purposes, and is usually disabled
// in production environments. It is used to log messages that are only
// relevant to developers.
//
// This method accepts a message string and a variadic number of arguments,
// which will be used to format the message string (similar to fmt.Printf verbs).
// See https://pkg.go.dev/fmt#hdr-Printing for more information.
func (logger *ZerologLogger) Debug(ctx context.Context, message string, args ...any) (err error) {
	return logger.log(ctx, Debug, message, args...)
}

// Info logs a message at the info level.
// The info level is used for logging general information about the application,
// such as startup messages, shutdown messages, etc.
//
// This method accepts a message string and a variadic number of arguments,
// which will be used to format the message string (similar to fmt.Printf verbs).
// See https://pkg.go.dev/fmt#hdr-Printing for more information.
func (logger *ZerologLogger) Info(ctx context.Context, message string, args ...any) (err error) {
	return logger.log(ctx, Info, message, args...)
}

// Warn logs a message at the warn level.
// The warn level is used for logging messages that are not critical, but
// might be worth investigating. For example, a warning message might be
// logged when a request to an external service fails, but the application
// is able to recover from it.
//
// This method accepts a message string and a variadic number of arguments,
// which will be used to format the message string (similar to fmt.Printf verbs).
// See https://pkg.go.dev/fmt#hdr-Printing for more information.
func (logger *ZerologLogger) Warn(ctx context.Context, message string, args ...any) (err error) {
	return logger.log(ctx, Warn, message, args...)
}

// Error logs a message at the error level.
// The error level is used for logging messages that are critical, and
// require immediate attention. For example, an error message might be
// logged when a request to an external service fails, and the application
// is not able to recover from it.

// This method accepts a message string and a variadic number of arguments,
// which will be used to format the message string (similar to fmt.Printf verbs).
// See https://pkg.go.dev/fmt#hdr-Printing for more information.
func (logger *ZerologLogger) Error(ctx context.Context, message string, args ...any) (err error) {
	return logger.log(ctx, Error, message, args...)
}

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
func (logger *ZerologLogger) Fatal(ctx context.Context, message string, args ...any) (err error) {
	err = logger.log(ctx, Fatal, message, args...)
	os.Exit(1)
	return
}

// Engine returns the underlying logging implementation.
func (logger *ZerologLogger) Engine() *zerolog.Logger {
	return logger.engine
}
