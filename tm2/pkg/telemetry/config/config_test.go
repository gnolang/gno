package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("exporter endpoint not set", func(t *testing.T) {
		t.Parallel()

		c := DefaultTelemetryConfig()
		c.ExporterEndpoint = "" // empty

		assert.ErrorIs(t, c.ValidateBasic(), errEndpointNotSet)
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		c := DefaultTelemetryConfig()
		c.ExporterEndpoint = "0.0.0.0:8080"

		assert.NoError(t, c.ValidateBasic())
	})
}
