package main

import (
	"os"
	"strings"
	"testing"
)

var (
	testBuildVersion string
	testCommitHash   = "e872fa"
)

func init() {
	if isTestTaggedVersion() {
		testBuildVersion = "chain/test4.2"
	} else {
		testBuildVersion = "master.387+e872fa"
	}
	buildVersion = testBuildVersion
	commitHash = testCommitHash
}

func isTestTaggedVersion() bool {
	testDir := os.Getenv("TEST_CASE_DIR")
	return strings.Contains(testDir, "tagged_version")
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
	}

	for _, testCase := range testCases {
		os.Setenv("TEST_CASE_DIR", testCase.testDir)
		if isTestTaggedVersion() {
			buildVersion = "chain/test4.2"
		} else {
			buildVersion = "master.387+e872fa"
		}
		testMainCaseRun(t, []testMainCase{testCase})
	}
}
