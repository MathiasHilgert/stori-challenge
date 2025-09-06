package blend

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestZerologLogger_log(t *testing.T) {
	t.Run("it should return an error if the logger engine is nil", func(t *testing.T) {
		// Arrange.
		var (
			ctx        context.Context = context.Background()
			usedOutput io.Writer       = new(bytes.Buffer)
			usedLogger *ZerologLogger  = &ZerologLogger{
				engine: nil,
				output: usedOutput,
			}

			expectedOutput string = ""
			expectedErr           = io.ErrClosedPipe
		)

		// Act.
		actualErr := usedLogger.log(ctx, Debug, "Hello, world!")

		// Assert.
		assert.Equal(t, expectedOutput, usedOutput.(*bytes.Buffer).String())
		assert.EqualError(t, actualErr, expectedErr.Error())
	})

	t.Run("it should return an error if the output writer is nil", func(t *testing.T) {
		// Arrange.
		var (
			ctx        context.Context = context.Background()
			usedOutput io.Writer       = nil
			usedLogger *ZerologLogger  = &ZerologLogger{
				engine: new(zerolog.Logger),
				output: usedOutput,
			}

			expectedErr = io.ErrClosedPipe
		)

		// Act.
		actualErr := usedLogger.log(ctx, Debug, "Hello, world!")

		// Assert.
		assert.EqualError(t, actualErr, expectedErr.Error())
	})

	t.Run("it should log a message at the debug level", func(t *testing.T) {
		// Arrange.
		var (
			ctx context.Context = context.Background()

			usedOutput    io.Writer = bytes.NewBuffer([]byte{})
			usedLogger, _           = NewZerologLogger(usedOutput)

			expectedOutput string = "{\"level\":\"debug\",\"time\":\"2003-05-01T00:00:00Z\",\"message\":\"Hello, world! :D\"}"
		)
		usedLogger.now = func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2003-05-01T00:00:00Z")
			return t
		}

		// Act.
		actualErr := usedLogger.log(ctx, Debug, "Hello, world! %s", ":D")

		// Assert.
		assert.JSONEq(t, expectedOutput, usedLogger.output.(*bytes.Buffer).String())
		assert.NoError(t, actualErr)
	})

	t.Run("it should log a message at the info level", func(t *testing.T) {
		// Arrange.
		var (
			ctx context.Context = context.Background()

			usedOutput    io.Writer = bytes.NewBuffer([]byte{})
			usedLogger, _           = NewZerologLogger(usedOutput)

			expectedOutput string = "{\"level\":\"info\",\"time\":\"2003-05-01T00:00:00Z\",\"message\":\"Hello, world! :D\"}"
		)
		usedLogger.now = func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2003-05-01T00:00:00Z")
			return t
		}

		// Act.
		actualErr := usedLogger.log(ctx, Info, "Hello, world! %s", ":D")

		// Assert.
		assert.JSONEq(t, expectedOutput, usedLogger.output.(*bytes.Buffer).String())
		assert.NoError(t, actualErr)
	})

	t.Run("it should log a message at the warn level", func(t *testing.T) {
		// Arrange.
		var (
			ctx context.Context = context.Background()

			usedOutput    io.Writer = bytes.NewBuffer([]byte{})
			usedLogger, _           = NewZerologLogger(usedOutput)

			expectedOutput string = "{\"level\":\"warn\",\"time\":\"2003-05-01T00:00:00Z\",\"message\":\"Hello, world! :D\"}"
		)
		usedLogger.now = func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2003-05-01T00:00:00Z")
			return t
		}

		// Act.
		actualErr := usedLogger.log(ctx, Warn, "Hello, world! %s", ":D")

		// Assert.
		assert.JSONEq(t, expectedOutput, usedLogger.output.(*bytes.Buffer).String())
		assert.NoError(t, actualErr)
	})

	t.Run("it should log a message at the error level", func(t *testing.T) {
		// Arrange.
		var (
			ctx context.Context = context.Background()

			usedOutput    io.Writer = bytes.NewBuffer([]byte{})
			usedLogger, _           = NewZerologLogger(usedOutput)

			expectedOutput string = "{\"level\":\"error\",\"time\":\"2003-05-01T00:00:00Z\",\"message\":\"Hello, world! :D\"}"
		)
		usedLogger.now = func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2003-05-01T00:00:00Z")
			return t
		}

		// Act.
		actualErr := usedLogger.log(ctx, Error, "Hello, world! %s", ":D")

		// Assert.
		assert.JSONEq(t, expectedOutput, usedLogger.output.(*bytes.Buffer).String())
		assert.NoError(t, actualErr)
	})
	// We are not going to test Fatal, as it terminates the application (I don't know how to test that :P).
}
