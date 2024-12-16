package main

import (
	"fmt"
	"testing"
)

func TestEnvApp(t *testing.T) {
	const (
		testGnoRootEnv = "/faster/better/stronger"
		testGnoHomeEnv = "/around/the/world"
	)

	t.Setenv("GNOROOT", testGnoRootEnv)
	t.Setenv("GNOHOME", testGnoHomeEnv)
	tc := []testMainCase{
		// shell
		{args: []string{"env", "foo"}, stdoutShouldBe: "\n"},
		{args: []string{"env", "foo", "bar"}, stdoutShouldBe: "\n\n"},
		{args: []string{"env", "GNOROOT"}, stdoutShouldBe: testGnoRootEnv + "\n"},
		{args: []string{"env", "GNOHOME", "storm"}, stdoutShouldBe: testGnoHomeEnv + "\n\n", noTmpGnohome: true},
		{args: []string{"env"}, stdoutShouldContain: fmt.Sprintf("GNOROOT=%q", testGnoRootEnv)},
		{args: []string{"env"}, stdoutShouldContain: fmt.Sprintf("GNOHOME=%q", testGnoHomeEnv), noTmpGnohome: true},

		// json
		{args: []string{"env", "-json"}, stdoutShouldContain: fmt.Sprintf("\"GNOROOT\": %q", testGnoRootEnv)},
		{args: []string{"env", "-json"}, stdoutShouldContain: fmt.Sprintf("\"GNOHOME\": %q", testGnoHomeEnv), noTmpGnohome: true},
		{
			args:           []string{"env", "-json", "GNOROOT"},
			stdoutShouldBe: fmt.Sprintf("{\n\t\"GNOROOT\": %q\n}\n", testGnoRootEnv),
		},
		{
			args:           []string{"env", "-json", "GNOROOT", "storm"},
			stdoutShouldBe: fmt.Sprintf("{\n\t\"GNOROOT\": %q,\n\t\"storm\": \"\"\n}\n", testGnoRootEnv),
		},
		{
			args:           []string{"env", "-json", "storm"},
			stdoutShouldBe: "{\n\t\"storm\": \"\"\n}\n",
		},
	}

	testMainCaseRun(t, tc)
}
