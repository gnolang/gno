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
	}
	testMainCaseRun(t, tc)
}
