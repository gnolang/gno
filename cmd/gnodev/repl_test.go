package main

import "testing"

func TestReplApp(t *testing.T) {
	tc := []testMainCase{
		// {args: []string{"repl"}, errShouldBe: "invalid args", stderrShouldBe: "Usage: repl [precompile flags] [packages]\n"},
		{args: []string{"repl", "--help"}, stdoutShouldContain: "# replOptions options\n-"},

		// args
		// {args: []string{"repl", "..."}, stdoutShouldContain: "..."},
	}
	testMainCaseRun(t, tc)
}
