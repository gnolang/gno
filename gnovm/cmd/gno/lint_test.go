package main

import "testing"

func TestLintApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"lint"},
			errShouldBe: "flag: help requested",
		}, {
			args:                []string{"lint", "--set_exit_status=0", "../../tests/integ/run-main/"},
			stderrShouldContain: "./../../tests/integ/run-main: missing 'gno.mod' file (code=1).",
		}, {
			args:                []string{"lint", "--set_exit_status=0", "../../tests/integ/undefined-variable-test/undefined_variables_test.gno"},
			stderrShouldContain: "undefined_variables_test.gno:6: name toto not declared (code=2)",
		}, {
			args:                []string{"lint", "--set_exit_status=0", "../../tests/integ/package-not-declared/main.gno"},
			stderrShouldContain: "main.gno:4: name fmt not declared (code=2).",
		}, {
			args: []string{"lint", "--set_exit_status=0", "../../tests/integ/minimalist-gnomod/"},
			// TODO: raise an error because there is a gno.mod, but no .gno files
		}, {
			args: []string{"lint", "--set_exit_status=0", "../../tests/integ/invalid-module-name/"},
			// TODO: raise an error because gno.mod is invalid
		},

		// TODO: 'gno mod' is valid?
		// TODO: is gno source valid?
		// TODO: are dependencies valid?
		// TODO: is gno source using unsafe/discouraged features?
		// TODO: consider making `gno precompile; go lint *gen.go`
		// TODO: check for imports of native libs from non _test.gno files
	}
	testMainCaseRun(t, tc)
}
