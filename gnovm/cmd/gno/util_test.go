package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func createGnoPackages(t *testing.T, tmpDir string) {
	t.Helper()

	type file struct {
		name, data string
	}
	// Gno pkgs to create
	pkgs := []struct {
		dir   string
		files []file
	}{
		// pkg 'foo', 'bar' and 'baz'
		{
			dir: filepath.Join(tmpDir, "foo"),
			files: []file{
				{
					name: "foo.gno",
					data: `package foo`,
				},
			},
		},
		{
			dir: filepath.Join(tmpDir, "bar"),
			files: []file{
				{
					name: "bar.gno",
					data: `package bar`,
				},
			},
		},
		{
			dir: filepath.Join(tmpDir, "baz"),
			files: []file{
				{
					name: "baz.gno",
					data: `package baz`,
				},
			},
		},

		// pkg inside 'foo' pkg
		{
			dir: filepath.Join(tmpDir, "foo", "qux"),
			files: []file{
				{
					name: "qux.gno",
					data: `package qux`,
				},
			},
		},

		// pkg inside 'bar' pkg
		{
			dir: filepath.Join(tmpDir, "bar", "quux"),
			files: []file{
				{
					name: "quux.gno",
					data: `package quux`,
				},
			},
		},

		// pkg inside 'foo/qux' pkg
		{
			dir: filepath.Join(tmpDir, "foo", "qux", "corge"),
			files: []file{
				{
					name: "corge.gno",
					data: `package corge`,
				},
			},
		},
	}

	// Create pkgs
	for _, p := range pkgs {
		err := os.MkdirAll(p.dir, 0o700)
		require.NoError(t, err)
		for _, f := range p.files {
			err = os.WriteFile(filepath.Join(p.dir, f.name), []byte(f.data), 0o644)
			require.NoError(t, err)
		}
	}
}
