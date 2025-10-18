package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
}
