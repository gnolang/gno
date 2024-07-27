package main

import "testing"

func TestGnoDepGraph(t *testing.T) {
	tc := []testMainCase{
		{
			// lacking input
			args:        []string{"depgraph"},
			errShouldBe: "flag: help requested",
		},
		{
			// input is not a package
			args:             []string{"depgraph", "supercalifragilisticexpialidocious"},
			errShouldContain: "error in parsing gno.mod",
		},
		{
			// given input where some requires are missing

			args:             []string{"depgraph", "../../../examples/gno.land/p/"},
			errShouldContain: "error in building graph",
		},
	}
	testMainCaseRun(t, tc)
}
