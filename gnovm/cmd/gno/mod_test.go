package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"mod"},
			errShouldBe: "flag: help requested",
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
			testDir:              "../../tests/integ/invalid_module_version1",
			simulateExternalRepo: true,
			errShouldContain:     "error parsing gno.mod file at",
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
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/invalid_gno_file",
			simulateExternalRepo: true,
			errShouldContain:     "expected 'package', found packag",
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
			testDir:              "../../tests/integ/invalid_module_version1",
			simulateExternalRepo: true,
			errShouldContain:     "error parsing gno.mod file at",
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

func TestGetGnoImports(t *testing.T) {
	workingDir, err := os.Getwd()
	require.NoError(t, err)

	// create external dir
	tmpDir, cleanUpFn := createTmpDir(t)
	defer cleanUpFn()

	// cd to tmp directory
	os.Chdir(tmpDir)
	defer os.Chdir(workingDir)

	files := []struct {
		name, data string
	}{
		{
			name: "file1.gno",
			data: `
			package tmp

			import (
				"std"

				"gno.land/p/demo/pkg1"
			)
			`,
		},
		{
			name: "file2.gno",
			data: `
			package tmp

			import (
				"gno.land/p/demo/pkg1"
				"gno.land/p/demo/pkg2"
			)
			`,
		},
		{
			name: "file1_test.gno",
			data: `
			package tmp

			import (
				"testing"

				"gno.land/p/demo/testpkg"
			)
			`,
		},
		{
			name: "z_0_filetest.gno",
			data: `
			package main

			import (
				"gno.land/p/demo/filetestpkg"
			)
			`,
		},

		// subpkg files
		{
			name: filepath.Join("subtmp", "file1.gno"),
			data: `
			package subtmp

			import (
				"std"

				"gno.land/p/demo/subpkg1"
			)
			`,
		},
		{
			name: filepath.Join("subtmp", "file2.gno"),
			data: `
			package subtmp

			import (
				"gno.land/p/demo/subpkg1"
				"gno.land/p/demo/subpkg2"
			)
			`,
		},
	}

	// Expected list of imports
	// - ignore subdirs
	// - ignore duplicate
	// - ignore *_filetest.gno
	// - should be sorted
	expected := []string{
		"gno.land/p/demo/pkg1",
		"gno.land/p/demo/pkg2",
		"gno.land/p/demo/testpkg",
	}

	// Create subpkg dir
	err = os.Mkdir("subtmp", 0o700)
	require.NoError(t, err)

	// Create files
	for _, f := range files {
		err = os.WriteFile(f.name, []byte(f.data), 0o644)
		require.NoError(t, err)
	}

	imports, err := getGnoPackageImports(tmpDir)
	require.NoError(t, err)

	require.Equal(t, len(expected), len(imports))
	for i := range imports {
		assert.Equal(t, expected[i], imports[i])
	}
}
