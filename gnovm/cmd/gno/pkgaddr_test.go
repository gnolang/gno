package main

import (
	"testing"
)

func TestPkgAddrApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"pkgaddr"},
			errShouldBe: "expected 1 arg, got 0",
		},

		{
			args:           []string{"pkgaddr", "gno.land/r/demo/users"},
			stdoutShouldBe: "g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s\n",
		},
	}

	testMainCaseRun(t, tc)
}
