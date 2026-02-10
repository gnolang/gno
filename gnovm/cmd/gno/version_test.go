package main

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/version"
)

func TestVersionApp(t *testing.T) {
	originalVersion := version.Version

	t.Cleanup(func() {
		version.Version = originalVersion
	})

	versionValues := []string{"chain/test4.2", "develop", "master"}

	testCases := make([]testMainCase, len(versionValues))
	for i, v := range versionValues {
		testCases[i] = testMainCase{
			args:                []string{"version"},
			stdoutShouldContain: "gno version: " + v,
		}
	}

	for i, testCase := range testCases {
		t.Run(versionValues[i], func(t *testing.T) {
			version.Version = versionValues[i]
			testMainCaseRun(t, []testMainCase{testCase})
		})
	}
}
