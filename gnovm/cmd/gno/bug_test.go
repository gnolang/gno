package main

import "testing"

func TestBugApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"bug -h"},
			errShouldBe: "flag: help requested",
		},
		{
			args:        []string{"bug unknown"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                []string{"bug", "-skip-browser"},
			stdoutShouldContain: "Go version go1.",
		},
	}
	testMainCaseRun(t, tc)
}
