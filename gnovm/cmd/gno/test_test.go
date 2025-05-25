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
			name:             "file in wrong language",
			args:             []string{"test", "../../tests/integ/test/wrong_lang/file.go"},
			errShouldContain: "list sub packages: files must be .gno files",
		},
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
		// {`
		// 	name:                "lint failure",
		// 	args:                []string{"test", "-v", "../../tests/integ/test/type_error"},
		// 	errShouldContain:    "FAIL: 0 build errors, 1 test errors", // or whatever type/lint error message you induce
		// },`
		// {
		// 	name:                "foundErr true (but no fatal lint error)",
		// 	args:                []string{"test", "-v", "../../tests/integ/test/found_err"},
		// 	stderrShouldContain: "FAIL",
		// },
	}
	testMainCaseRun(t, tc)
}
