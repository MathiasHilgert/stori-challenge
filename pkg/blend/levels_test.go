package blend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevel_String(t *testing.T) {
	// Arrange.
	var (
		usedLevel      Level  = "debug"
		expectedString string = "debug"
	)

	// Act.
	actualString := usedLevel.String()

	// Assert.
	assert.Equal(t, expectedString, actualString)
}

func TestLevel_Equal(t *testing.T) {
	t.Run("should return true if the log level is equal to the given log level", func(t *testing.T) {
		// Arrange.
		var (
			usedLevel  Level = "debug"
			otherLevel Level = "debug"
		)

		// Act.
		actualResult := usedLevel.Equal(otherLevel)

		// Assert.
		assert.True(t, actualResult)
	})

	t.Run("should return false if the log level is not equal to the given log level", func(t *testing.T) {
		// Arrange.
		var (
			usedLevel  Level = "debug"
			otherLevel Level = "info"
		)

		// Act.
		actualResult := usedLevel.Equal(otherLevel)

		// Assert.
		assert.False(t, actualResult)
	})
}
