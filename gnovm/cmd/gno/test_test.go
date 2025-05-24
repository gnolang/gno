package main

import "testing"

func TestTestApp(t *testing.T) {
	tc := []testMainCase{
		{
			name:                "basic test with pass and fail",
			args:                []string{"test", "-v", "../../tests/integ/test/basic"},
			stderrShouldContain: "PASS: TestBasic/greater_than_one",
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
		},
		{
			name:                "basic test with pass only",
			args:                []string{"test", "-v", "-run", "TestBasic/greater.*", "../../tests/integ/test/basic"},
			stderrShouldContain: "PASS: TestBasic/greater_than_one",
		},
		{
			name:                "basic file test with pass only",
			args:                []string{"test", "-v", "../../tests/integ/test/basic_ft"},
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
			stdoutShouldContain: "Goodbye from filetest",
			stderrShouldContain: "PASS: ../../tests/integ/test/basic_ft/pass_filetest.gno",
		},
		{
			name:                "basic file test with pass and fail",
			args:                []string{"test", "../../tests/integ/test/basic_ft"},
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
			stderrShouldContain: "FAIL: ../../tests/integ/test/basic_ft/fail_filetest.gno",
		},
		{
			name:                "no package to test",
			args:                []string{"test"},
			stderrShouldContain: "?       . \t[no test files]",
		},
		{
			name:             "invalid pattern triggers targetsFromPatterns error",
			args:             []string{"test", "**[badpattern"},
			errShouldContain: "list targets from patterns",
		},
		{
			name:                "nonexistent path results in no packages",
			args:                []string{"test", "../../tests/integ/test/empty/..."},
			stderrShouldContain: "no packages to test",
		},
		// {
		// 	name:                "test times out",
		// 	args:                []string{"test", "-v", "-timeout=1ms", "../../tests/integ/test/infinite_loop"},
		// 	stderrShouldContain: "test timed out after 1ms",
		// },
		{
			name:                "sub packages error with malformed modfile",
			args:                []string{"test", "../../tests/integ/test/broken_mod"},
			stderrShouldContain: "--- WARNING: unable to read package path",
		},
		{
			name:                "directory with no test files",
			args:                []string{"test", "../../tests/integ/test/empty"},
			stderrShouldContain: "../../tests/integ/test/empty \t[no test files]",
		},
		{
			name:                "directory with test files containing no tests",
			args:                []string{"test", "../../tests/integ/test/no_tests"},
			stderrShouldContain: "testing: warning: no tests to run",
		},
		// {
		// 	name:                "directory with no gno.mod and nonstandard path",
		// 	args:                []string{"test", "../../tests/integ/test/unknown_pkg"},
		// 	stderrShouldContain: "--- WARNING: unable to read package path",
		// },
		// {
		// 	name:                "lint failure",
		// 	args:                []string{"test", "-v", "../../tests/integ/test/lint_fail"},
		// 	stderrShouldContain: "type mismatch", // or whatever type/lint error message you induce
		// },
		// {
		// 	name:                "foundErr true (but no fatal lint error)",
		// 	args:                []string{"test", "-v", "../../tests/integ/test/found_err"},
		// 	stderrShouldContain: "FAIL",
		// },
	}
	testMainCaseRun(t, tc)
}
