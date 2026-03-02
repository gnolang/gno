package urequire

import "testing"

func TestPackage(t *testing.T) {
	Equal(t, 42, 42)

	// XXX: find a way to unit test this package thoroughly,
	// especially the t.FailNow() behavior on assertion failure.
}
