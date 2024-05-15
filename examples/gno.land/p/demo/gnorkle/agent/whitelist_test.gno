package agent_test

import (
	"testing"

	"gno.land/p/demo/gnorkle/agent"
)

func TestWhitelist(t *testing.T) {
	var whitelist agent.Whitelist

	if whitelist.HasDefinition() {
		t.Error("whitelist should not be defined initially")
	}

	whitelist.AddAddresses([]string{"a", "b"})
	if !whitelist.HasAddress("a") {
		t.Error(`whitelist should have address "a"`)
	}
	if !whitelist.HasAddress("b") {
		t.Error(`whitelist should have address "b"`)
	}

	if !whitelist.HasDefinition() {
		t.Error("whitelist should be defined after adding addresses")
	}

	whitelist.RemoveAddress("a")
	if whitelist.HasAddress("a") {
		t.Error(`whitelist should not have address "a"`)
	}
	if !whitelist.HasAddress("b") {
		t.Error(`whitelist should still have address "b"`)
	}

	whitelist.ClearAddresses()
	if whitelist.HasAddress("a") {
		t.Error(`whitelist cleared; should not have address "a"`)
	}
	if whitelist.HasAddress("b") {
		t.Error(`whitelist cleared; should still have address "b"`)
	}

	if whitelist.HasDefinition() {
		t.Error("whitelist cleared; should not be defined")
	}
}
