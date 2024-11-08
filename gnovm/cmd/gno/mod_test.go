package main

import (
	"testing"
)

func TestModApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"mod"},
			errShouldBe: "flag: help requested",
		},

		// test `gno mod download`
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldBe:          "gno.mod not found",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "validate: requires module",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/invalid_module_name",
			simulateExternalRepo: true,
			errShouldContain:     "usage: module module/path",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/demo/avl",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_invalid_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/demo/notexists",
			errShouldContain:     "fetch: failed to query files list for pkg \"gno.land/p/demo/notexists\": package \"gno.land/p/demo/notexists\" is not available",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_std_lib",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace_with_dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace_with_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/demo/users",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace_with_invalid_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/demo/notexists",
			errShouldContain:     "fetch: failed to query files list for pkg \"gno.land/p/demo/notexists\": package \"gno.land/p/demo/notexists\" is not available",
		},

		// test `gno mod init` with no module name
		{
			args:                 []string{"mod", "init"},
			testDir:              "../../tests/integ/valid1",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "init"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldBe:          "create gno.mod file: cannot determine package name",
		},
		{
			args:                 []string{"mod", "init"},
			testDir:              "../../tests/integ/empty_gno1",
			simulateExternalRepo: true,
			recoverShouldContain: "expected 'package', found 'EOF'",
		},
		{
			args:                 []string{"mod", "init"},
			testDir:              "../../tests/integ/empty_gno2",
			simulateExternalRepo: true,
			recoverShouldContain: "expected 'package', found 'EOF'",
		},
		{
			args:                 []string{"mod", "init"},
			testDir:              "../../tests/integ/empty_gno3",
			simulateExternalRepo: true,
			recoverShouldContain: "expected 'package', found 'EOF'",
		},
		{
			args:                 []string{"mod", "init"},
			testDir:              "../../tests/integ/empty_gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "create gno.mod file: gno.mod file already exists",
		},

		// test `gno mod init` with module name
		{
			args:                 []string{"mod", "init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gno1",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gno2",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gno3",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "create gno.mod file: gno.mod file already exists",
		},

		// test `gno mod tidy`
		{
			args:                 []string{"mod", "tidy", "arg1"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			errShouldContain:     "flag: help requested",
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldContain:     "could not read gno.mod file",
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
		},

		// test `gno mod why`
		{
			args:                 []string{"mod", "why"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			errShouldContain:     "flag: help requested",
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldContain:     "could not read gno.mod file",
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/invalid_gno_file",
			simulateExternalRepo: true,
			errShouldContain:     "expected 'package', found packag",
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module minim does not need package std)
`,
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/tests/importavl does not need package std)
`,
		},
		{
			args:                 []string{"mod", "why", "std", "gno.land/p/demo/avl"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/p/integ/valid does not need package std)

# gno.land/p/demo/avl
valid.gno
`,
		},
	}

	testMainCaseRun(t, tc)
}
