package main

import "testing"

func TestBugApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:             []string{"bug", "-h"},
			errShouldContain: "flag: help requested",
		},
		{
			args:        []string{"bug", "unknown"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"bug", "-skip-browser"},
			stdoutShouldContain: "Go version: go1.",
		},
		{
			args:                []string{"bug", "-skip-browser"},
			stdoutShouldContain: "Gno version: develop",
		},
	}
	testMainCaseRun(t, tc)
}
