package main

import "testing"

func TestTestApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:                []string{"test", "-v", "../../tests/integ/test/basic"},
			stderrShouldContain: "PASS: TestBasic/greater_than_one",
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
		},
		{
			args:                   []string{"test", "-v", "-run", "TestBasic/greater.*", "../../tests/integ/test/basic"},
			stderrShouldContain:    "PASS: TestBasic/greater_than_one",
		},
		{
			args:                []string{"test", "-v", "../../tests/integ/test/basic_ft"},
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
			stdoutShouldContain: "Goodbye from filetest",
			stderrShouldContain: "PASS: ../../tests/integ/test/basic_ft/pass_filetest.gno",
		},
		{
			args:                []string{"test", "../../tests/integ/test/basic_ft"},
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
			stderrShouldContain: "FAIL: ../../tests/integ/test/basic_ft/fail_filetest.gno",
		},
	}
	testMainCaseRun(t, tc)
}
