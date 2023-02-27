package main

import "testing"

func TestModApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:           []string{"mod"},
			errShouldBe:    "invalid command",
			stderrShouldBe: "Usage: mod [flags] <command>\n",
		},
		{
			args:                []string{"mod", "--help"},
			stdoutShouldContain: "# modFlags options\n-",
		},

		// test gno.mod
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty-dir",
			simulateExternalRepo: true,
			errShouldBe:          "mod download: gno.mod not found",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty-gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "mod download: validate: requires module",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/invalid-module-name",
			simulateExternalRepo: true,
			errShouldContain:     "usage: module module/path",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/minimalist-gnomod",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require-remote-module",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require-invalid-module",
			simulateExternalRepo: true,
			errShouldContain:     "mod download: fetch: writepackage: querychain:",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/invalid-module-version1",
			simulateExternalRepo: true,
			errShouldContain:     "usage: require module/path v1.2.3",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/invalid-module-version2",
			simulateExternalRepo: true,
			errShouldContain:     "invalid: must be of the form v1.2.3",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace-module",
			simulateExternalRepo: true,
		},
	}
	testMainCaseRun(t, tc)
}
