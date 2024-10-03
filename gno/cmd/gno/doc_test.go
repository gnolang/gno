package main

import "testing"

func TestGnoDoc(t *testing.T) {
	tc := []testMainCase{
		{
			args:                []string{"doc", "io.Writer"},
			stdoutShouldContain: "Writer is the interface that wraps",
		},
		{
			args:                []string{"doc", "avl"},
			stdoutShouldContain: "func NewTree",
		},
		{
			args:                []string{"doc", "-u", "avl.Node"},
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
