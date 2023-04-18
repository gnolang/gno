package main

import "testing"

func TestModApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"mod"},
			errShouldBe: "flag: help requested",
		},

		// test gno.mod
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty-dir",
			simulateExternalRepo: true,
			errShouldBe:          "gno.mod not found",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty-gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "validate: requires module",
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
			errShouldContain:     "fetch: writepackage: querychain",
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
			testDir:              "../../tests/integ/replace-with-dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace-with-module",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace-with-invalid-module",
			simulateExternalRepo: true,
			errShouldContain:     "fetch: writepackage: querychain",
		},
	}
	testMainCaseRun(t, tc)
}
