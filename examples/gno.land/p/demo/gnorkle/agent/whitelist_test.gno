package agent_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/agent"
	"gno.land/p/nt/uassert"
)

func TestWhitelist(t *testing.T) {
	var whitelist agent.Whitelist

	uassert.False(t, whitelist.HasDefinition(), "whitelist should not be defined initially")

	whitelist.AddAddresses([]string{"a", "b"})
	uassert.True(t, whitelist.HasAddress("a"), `whitelist should have address "a"`)
	uassert.True(t, whitelist.HasAddress("b"), `whitelist should have address "b"`)
	uassert.True(t, whitelist.HasDefinition(), "whitelist should be defined after adding addresses")

	whitelist.RemoveAddress("a")
	uassert.False(t, whitelist.HasAddress("a"), `whitelist should not have address "a"`)
	uassert.True(t, whitelist.HasAddress("b"), `whitelist should still have address "b"`)

	whitelist.ClearAddresses()
	uassert.False(t, whitelist.HasAddress("a"), `whitelist cleared; should not have address "a"`)
	uassert.False(t, whitelist.HasAddress("b"), `whitelist cleared; should still have address "b"`)
	uassert.False(t, whitelist.HasDefinition(), "whitelist cleared; should not be defined")
}
