package packages

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/require"
)

func TestImports(t *testing.T) {
	workingDir, err := os.Getwd()
	require.NoError(t, err)

	// create external dir
	tmpDir := t.TempDir()

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
			name: "file2_test.gno",
			data: `
			package tmp_test

			import (
				"testing"

				"gno.land/p/demo/testpkg"
				"gno.land/p/demo/xtestdep"
			)
			`,
		},
		{
			name: "z_0_filetest.gno",
			data: `
			package main

			import (
				"gno.land/p/demo/filetestdep"
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

	// Expected lists of imports
	// - ignore subdirs
	// - ignore duplicate
	// - should be sorted
	expected := map[FileKind][]string{
		FileKindPackageSource: {
			"gno.land/p/demo/pkg1",
			"gno.land/p/demo/pkg2",
			"std",
		},
		FileKindTest: {
			"gno.land/p/demo/testpkg",
			"testing",
		},
		FileKindXTest: {
			"gno.land/p/demo/testpkg",
			"gno.land/p/demo/xtestdep",
			"testing",
		},
		FileKindFiletest: {
			"gno.land/p/demo/filetestdep",
		},
	}

	// Create subpkg dir
	err = os.Mkdir("subtmp", 0o700)
	require.NoError(t, err)

	// Create files
	for _, f := range files {
		err = os.WriteFile(f.name, []byte(f.data), 0o644)
		require.NoError(t, err)
	}

	pkg, err := gnolang.ReadMemPackage(tmpDir, "test")
	require.NoError(t, err)

	importsMap, err := Imports(pkg, nil)
	require.NoError(t, err)

	// ignore specs
	got := map[FileKind][]string{}
	for key, vals := range importsMap {
		stringVals := make([]string, len(vals))
		for i, val := range vals {
			stringVals[i] = val.PkgPath
		}
		got[key] = stringVals
	}

	require.Equal(t, expected, got)
}
