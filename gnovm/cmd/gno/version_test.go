package main

import (
	"os"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/version"
)

var testVersionString string

func init() {
	if isTestTaggedVersion() {
		testVersionString = "chain/test4.2"
	} else if isTestDevelopVersion() {
		testVersionString = "develop"
	} else {
		testVersionString = "master.387+e872fa"
	}
	version.Version = testVersionString
}

func isTestTaggedVersion() bool {
	testDir := os.Getenv("TEST_CASE_DIR")
	return strings.Contains(testDir, "tagged_version")
}

func isTestDevelopVersion() bool {
	testDir := os.Getenv("TEST_CASE_DIR")
	return strings.Contains(testDir, "develop_version")
}

func TestVersionApp(t *testing.T) {
	testCases := []testMainCase{
		{
			args:                []string{"version"},
			testDir:             "testdata/gno_version/branch_commit_version.txtar",
			stdoutShouldContain: "gno version: master.387+e872fa",
		},
		{
			args:                []string{"version"},
			testDir:             "testdata/gno_version/tagged_version.txtar",
			stdoutShouldContain: "gno version: chain/test4.2",
		},
		{
			args:                []string{"version"},
			testDir:             "testdata/gno_version/develop_version.txtar",
			stdoutShouldContain: "gno version: develop",
		},
	}

	for _, testCase := range testCases {
		os.Setenv("TEST_CASE_DIR", testCase.testDir)
		if isTestTaggedVersion() {
			version.Version = "chain/test4.2"
		} else if isTestDevelopVersion() {
			version.Version = "develop"
		} else {
			version.Version = "master.387+e872fa"
		}
		testMainCaseRun(t, []testMainCase{testCase})
	}
}
