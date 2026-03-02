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
			errShouldContain:     "gnowork.toml file not found in current or any parent directory",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_workspace",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno: warning: \"./...\" matched no packages\n",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "1 build error(s)",
			stderrShouldContain:  "invalid gnomod.toml: 'module' is required",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/invalid_module_name",
			simulateExternalRepo: true,
			errShouldBe:          "1 build error(s)",
			stderrShouldContain:  "invalid gnomod.toml: 'module' is required (type: *errors.errorString)",
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
			stderrShouldContain:  "gno: downloading gno.land/p/nt/avl/v0",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_invalid_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "query files list for pkg \"gno.land/p/demo/notexists\": package \"gno.land/p/demo/notexists\" is not available",
			errShouldBe:          "1 build error(s)",
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
			stderrShouldContain:  "gno: downloading gno.land/p/nt/avl/v0",
		},
		// TODO: that functionality is not available on gnomod.toml anymore. should we remove this?
		// {
		// 	args:                 []string{"mod", "download"},
		// 	testDir:              "../../tests/integ/replace_with_invalid_module",
		// 	simulateExternalRepo: true,
		// 	stderrShouldContain:  "gno: downloading gno.land/p/demo/notexists",
		// 	errShouldContain:     "query files list for pkg \"gno.land/p/demo/notexists\": package \"gno.land/p/demo/notexists\" is not available",
		// },

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
			errShouldBe:          "create gnomod.toml: file already exists",
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
			errShouldContain:     "gnomod.toml doesn't exist",
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
			errShouldContain:     "gnomod.toml doesn't exist",
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
(module gno.land/t/minim does not need package std)
`,
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/t/importavl does not need package std)
`,
		},
		{
			args:                 []string{"mod", "why", "std", "gno.land/p/nt/avl/v0"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/p/integ/valid does not need package std)

# gno.land/p/nt/avl/v0
valid.gno
`,
		},

		// test `gno mod graph`
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			stdoutShouldBe:       ``,
		},
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/valid1",
			simulateExternalRepo: true,
			stdoutShouldBe: `gno.vm/r/tests/integ/valid1 testing
`,
		},
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno: downloading gno.land/p/nt/avl/v0\n",
			stdoutShouldBe: `gno.land/p/integ/valid gno.land/p/integ/valid
gno.land/p/integ/valid gno.land/p/nt/avl/v0
gno.land/p/integ/valid testing
gno.land/p/nt/avl/v0 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/ufmt/v0
gno.land/p/nt/avl/v0 sort
gno.land/p/nt/avl/v0 strings
gno.land/p/nt/avl/v0 testing
`,
		},
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno: downloading gno.land/p/nt/avl/v0\n",
			stdoutShouldBe: `gno.land/t/importavl gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/ufmt/v0
gno.land/p/nt/avl/v0 sort
gno.land/p/nt/avl/v0 strings
gno.land/p/nt/avl/v0 testing
`,
		},
		{
			// gno.land/p/nt/avl/v0 is included from the test in the filetests subdir
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/valid3",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/nt/avl/v0\n",
			stdoutShouldBe: `gno.land/p/integ/valid3 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/ufmt/v0
gno.land/p/nt/avl/v0 sort
gno.land/p/nt/avl/v0 strings
gno.land/p/nt/avl/v0 testing
`,
		},
	}

	testMainCaseRun(t, tc)
}
