package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Run("should create a new Log instance with default configuration", func(t *testing.T) {
		serviceName := "test-service"
		logInstance := New(serviceName)

		assert.NotNil(t, logInstance)
		assert.Equal(t, InfoLevel, logInstance.level)
	})

}
