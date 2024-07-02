package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("exporter endpoint not set", func(t *testing.T) {
		t.Parallel()

		c := DefaultTelemetryConfig()
		c.ExporterEndpoint = "" // empty
		c.PrometheusAddr = ""

		assert.ErrorIs(t, c.ValidateBasic(), errEndpointNotSet)
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		c := DefaultTelemetryConfig()
		c.ExporterEndpoint = "0.0.0.0:8080"

		assert.NoError(t, c.ValidateBasic())
	})

	t.Run("valid hostname", func(t *testing.T) {
		t.Parallel()

		c := DefaultTelemetryConfig()

		hostname, err := os.Hostname()
		if err != nil {
			assert.Equal(t, "gno-node", c.ServiceInstanceID)
		} else {
			assert.Equal(t, hostname, c.ServiceInstanceID)
		}
	})
}
