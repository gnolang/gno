package main

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeGithub(t *testing.T) {
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
		}
		t.Setenv(envGithubClientID, "mock")

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
		}
		t.Setenv(envGithubClientID, "mock")
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
		}
		t.Setenv(envGithubClientSecret, "mock")
		t.Setenv(envGithubClientID, "mock")
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
