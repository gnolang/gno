package main

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultAccount_Seed = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"

func TestServe(t *testing.T) {
	t.Run("Serve without subcommand", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "help requested")
	})

	t.Run("Serve captcha without captcha-secret", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"captcha",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, ErrCaptchaMissing)
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

	t.Run("Serve github without clientID", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"github",
			"--chain-id",
			"dev",
			"--mnemonic",
			defaultAccount_Seed,
		}
		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errGithubClientIDMissing)
	})

	t.Run("Serve github without client secret", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"github",
			"--chain-id",
			"dev",
			"--mnemonic",
			defaultAccount_Seed,
			"--github-client-id",
			"mock",
		}
		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errGithubClientSecretMissing)
	})

	t.Run("Serve github cannot connect redis", func(t *testing.T) {
		cmd := newServeCmd()
		args := []string{
			"github",
			"--chain-id",
			"dev",
			"--mnemonic",
			defaultAccount_Seed,
			"--github-client-id",
			"mock",
		}
		t.Setenv(envGithubClientSecret, "mock")
		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to connect to redis")
	})

	t.Run("Serve github OK", func(t *testing.T) {
		redisServer := miniredis.RunT(t)
		cmd := newServeCmd()
		args := []string{
			"github",
			"--chain-id",
			"dev",
			"--mnemonic",
			defaultAccount_Seed,
			"--github-client-id",
			"mock",
		}
		t.Setenv(envGithubClientSecret, "mock")
		t.Setenv(envRedisAddr, redisServer.Addr())
		// Run the command
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Millisecond * 300)
			cancel()
		}()
		// Run the command
		cmdErr := cmd.ParseAndRun(ctx, args)
		require.NoError(t, cmdErr)
	})
}
