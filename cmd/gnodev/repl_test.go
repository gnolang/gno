package main

import "testing"

func TestReplApp(t *testing.T) {
	tc := []testMainCase{
		{args: []string{"repl", "invalid-arg"}, errShouldBe: "flag: help requested"},

		// args
		// {args: []string{"repl", "..."}, stdoutShouldContain: "..."},
	}
	testMainCaseRun(t, tc)
}
