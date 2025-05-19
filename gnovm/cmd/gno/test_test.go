package main

import "testing"

func TestTestApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:                []string{"test","-v","../../tests/integ/test/basic"},
			stderrShouldContain: "PASS: TestBasic/greater_than_one",
			errShouldBe:         "FAIL: 0 build errors, 1 test errors",
		},
	}
	testMainCaseRun(t, tc)
}