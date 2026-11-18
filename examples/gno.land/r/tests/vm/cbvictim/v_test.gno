package cbvictim

import "testing"

// The runner calls each test through a value it made: the test still writes.
func TestRunnerWrites(t *testing.T) {
	SetOwner("tested")
	if Owner != "tested" {
		t.Fatalf("owner: %s", Owner)
	}
	LibGateless()
	if Lib.Field != "pwnd-same-pkg" {
		t.Fatalf("lib: %s", Lib.Field)
	}
}

func ExampleSetOwner() {
	SetOwner("example")
	println(Owner)
	// Output: example
}
