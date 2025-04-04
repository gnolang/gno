package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeCapcha(t *testing.T) {
	t.Run("Serve captcha without captcha-secret", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"captcha",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errCaptchaMissing)
	})

	t.Run("Serve captcha without chain-id", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"captcha",
			"--captcha-secret",
			"dummy-secret",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "invalid chain ID")
	})

	t.Run("Serve captcha with invalid mnemonic", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"captcha",
			"--captcha-secret",
			"dummy-secret",
			"--chain-id",
			"dev",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "invalid mnemonic")
	})

	t.Run("Serve captcha OK", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"captcha",
			"--captcha-secret",
			"dummy-secret",
			"--chain-id",
			"dev",
			"--mnemonic",
			defaultAccount_Seed,
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Millisecond * 100)
			cancel()
		}()
		// Run the command
		cmdErr := cmd.ParseAndRun(ctx, args)
		require.NoError(t, cmdErr)
	})
}
