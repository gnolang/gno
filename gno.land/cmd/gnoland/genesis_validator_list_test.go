package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
)

func TestGenesis_Validator_List(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis", "validator", "list", "--genesis-path", "dummy-path",
		}
		assert.ErrorContains(t,
			cmd.ParseAndRun(context.Background(), args),
			errUnableToLoadGenesis.Error(),
		)
	})

	// There's not much else to test for now
}
