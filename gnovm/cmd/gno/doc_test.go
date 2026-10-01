package main

import "testing"

func TestGnoDoc(t *testing.T) {
	tc := []testMainCase{
		{
			args:                []string{"doc", "io.Writer"},
			stdoutShouldContain: "Writer is the interface that wraps",
		},
		{
			args:                []string{"doc", "gno.land/p/nt/avl/v0"},
			stdoutShouldContain: "func NewTree",
		},
		{
			args:                []string{"doc", "-u", "gno.land/p/nt/avl/v0.Node"},
			stdoutShouldContain: "node *Node",
		},
		{
			args:             []string{"doc", "dkfdkfkdfjkdfj"},
			errShouldContain: "package not found",
		},
		{
			args:             []string{"doc", "There.Are.Too.Many.Dots"},
			errShouldContain: "invalid arguments",
		},
	}
	testMainCaseRun(t, tc)
}
