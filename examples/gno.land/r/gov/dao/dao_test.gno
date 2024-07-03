package govdao

import (
	"std"
	"testing"
)

func TestPackage(t *testing.T) {
	addr := std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
	members = append(members, addr)

	std.TestSetOrigCaller(addr)

	out := Render("")

	expected := "No proposals found :(."
	if out != expected {
		t.Errorf("expected %s, got %s.", expected, out)
	}

	// TODO: add a new address, from members proposal, after that vote another proposal using two addresses.

}
