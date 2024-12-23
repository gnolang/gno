package testcycles_a

import (
	"testing"

	"gno.land/p/demo/tests/testcycles/testcycles_b"
)

func TestFoo(t *testing.T) {
	testcycles_b.TCB()
}
