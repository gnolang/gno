package cbvictim

import (
	"testing"

	"gno.land/p/demo/tests/pfuncbind"
)

// The runner calls each test through a value it made: the test still writes.
func TestRunnerWrites(t *testing.T) {
	SetOwner("tested")
	if Owner != "tested" {
		t.Fatalf("owner: %s", Owner)
	}
	// A library writer called through a value runs view-only, even here;
	// called by name it writes.
	func() {
		defer func() {
			if recover() == nil {
				t.Fatalf("LibGateless wrote through a value")
			}
		}()
		LibGateless()
	}()
	pfuncbind.Clobber(Lib)
	if Lib.Field != "pwnd-same-pkg" {
		t.Fatalf("lib: %s", Lib.Field)
	}
}

func ExampleSetOwner() {
	SetOwner("example")
	println(Owner)
	// Output: example
}
