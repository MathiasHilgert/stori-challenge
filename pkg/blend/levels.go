package blend

// Level represents a log level which can be used to filter log messages.
type Level string

// String returns the string representation of the log level.
func (level Level) String() string {
	return string(level)
}

// Equal returns true if the log level is equal to the given log level.
func (level Level) Equal(other Level) bool {
	return string(level) == string(other)
}

const (
	// Debug is a log level that is used for debugging purposes, and is usually
	// disabled in production environments. It is used to log messages that are
	// only relevant to developers.
	Debug Level = "debug"

	// Info is a log level that is used for logging general information about
	// the application, such as startup messages, shutdown messages, etc.
	Info Level = "info"

	// Warn is a log level that is used for logging messages that are not
	// critical, but might be worth investigating. For example, a warning
	// message might be logged when a request to an external service fails,
	// but the application is able to recover from it.
	Warn Level = "warn"

	// Error is a log level that is used for logging messages that are critical,
	// and require immediate attention. For example, an error message might be
	// logged when a request to an external service fails, and the application
	// is not able to recover from it.
	Error Level = "error"

	// Fatal is a log level that is used for logging messages that are critical,
	// and require immediate attention. For example, a fatal message might be
	// logged when the application is not able to connect to the database.
	//
	// Be careful when using this level, as it will terminate the application
	// after logging the message. It's recommended only on main startup functions.
	//
	// If you are a newbie, get away from this level as far as possible ヾ(￣▽￣) Bye~Bye~
	Fatal Level = "fatal"
)
