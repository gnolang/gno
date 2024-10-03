package main

import "testing"

func TestLintApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"lint"},
			errShouldBe: "flag: help requested",
		}, {
			args:                []string{"lint", "../../tests/integ/run_main/"},
			stderrShouldContain: "./../../tests/integ/run_main: missing 'gno.mod' file (code=1).",
			errShouldBe:         "exit code: 1",
		}, {
			args:                []string{"lint", "../../tests/integ/undefined_variable_test/undefined_variables_test.gno"},
			stderrShouldContain: "undefined_variables_test.gno:6: name toto not declared (code=2)",
			errShouldBe:         "exit code: 1",
		}, {
			args:                []string{"lint", "../../tests/integ/package_not_declared/main.gno"},
			stderrShouldContain: "main.gno:4: name fmt not declared (code=2).",
			errShouldBe:         "exit code: 1",
		}, {
			args:                []string{"lint", "../../tests/integ/several-lint-errors/main.gno"},
			stderrShouldContain: "../../tests/integ/several-lint-errors/main.gno:5: expected ';', found example (code=2).\n../../tests/integ/several-lint-errors/main.gno:6",
			errShouldBe:         "exit code: 1",
		}, {
			args:                []string{"lint", "../../tests/integ/run_main/"},
			stderrShouldContain: "./../../tests/integ/run_main: missing 'gno.mod' file (code=1).",
			errShouldBe:         "exit code: 1",
		}, {
			args: []string{"lint", "../../tests/integ/minimalist_gnomod/"},
			// TODO: raise an error because there is a gno.mod, but no .gno files
		}, {
			args: []string{"lint", "../../tests/integ/invalid_module_name/"},
			// TODO: raise an error because gno.mod is invalid
		},

		// TODO: 'gno mod' is valid?
		// TODO: is gno source valid?
		// TODO: are dependencies valid?
		// TODO: is gno source using unsafe/discouraged features?
		// TODO: consider making `gno transpile; go lint *gen.go`
		// TODO: check for imports of native libs from non _test.gno files
	}
	testMainCaseRun(t, tc)
}
