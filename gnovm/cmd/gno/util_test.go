package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		names    []string
		expected []bool
	}{
		{
			pattern:  "foo",
			names:    []string{"foo", "bar", "baz", "foo/bar"},
			expected: []bool{true, false, false, false},
		},
		{
			pattern:  "foo/...",
			names:    []string{"foo", "foo/bar", "foo/bar/baz", "bar", "baz"},
			expected: []bool{true, true, true, false, false},
		},
		{
			pattern:  "foo/bar/...",
			names:    []string{"foo/bar", "foo/bar/baz", "foo/baz/bar", "foo", "bar"},
			expected: []bool{true, true, false, false, false},
		},
		{
			pattern:  "foo/.../baz",
			names:    []string{"foo/bar", "foo/bar/baz", "foo/baz/bar", "foo", "bar"},
			expected: []bool{false, true, false, false, false},
		},
		{
			pattern:  "foo/.../baz/...",
			names:    []string{"foo/bar/baz", "foo/baz/bar", "foo/bar/baz/qux", "foo/baz/bar/qux"},
			expected: []bool{true, false, true, false},
		},
		{
			pattern:  "...",
			names:    []string{"foo", "bar", "baz", "foo/bar", "foo/bar/baz"},
			expected: []bool{true, true, true, true, true},
		},
		{
			pattern:  ".../bar",
			names:    []string{"foo", "bar", "baz", "foo/bar", "foo/bar/baz"},
			expected: []bool{false, false, false, true, false},
		},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			matchFunc := matchPattern(test.pattern)
			for i, name := range test.names {
				res := matchFunc(name)
				assert.Equal(t, test.expected[i], res, "Expected: %v, Got: %v", test.expected[i], res)
			}
		})
	}
}

func TestTargetsFromPatterns(t *testing.T) {
	tmpDir := t.TempDir()
	createGnoPackages(t, tmpDir)

	for _, tc := range []struct {
		desc               string
		in, expected       []string
		errorShouldContain string
	}{
		{
			desc: "valid1",
			in: []string{
				tmpDir,
			},
			expected: []string{
				tmpDir,
			},
		},
		{
			desc: "valid2",
			in: []string{
				tmpDir + "/foo",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo"),
			},
		},
		{
			desc: "valid_recursive1",
			in: []string{
				tmpDir + "/...",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo"),
				filepath.Join(tmpDir, "bar"),
				filepath.Join(tmpDir, "baz"),
				filepath.Join(tmpDir, "foo", "qux"),
				filepath.Join(tmpDir, "bar", "quux"),
				filepath.Join(tmpDir, "foo", "qux", "corge"),
			},
		},
		{
			desc: "valid_recursive2",
			in: []string{
				tmpDir + "/foo/...",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo"),
				filepath.Join(tmpDir, "foo", "qux"),
				filepath.Join(tmpDir, "foo", "qux", "corge"),
			},
		},
		{
			desc: "valid_recursive2",
			in: []string{
				tmpDir + "/.../qux",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo", "qux"),
			},
		},
		{
			desc: "valid_recursive3",
			in: []string{
				tmpDir + "/.../qux/...",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo", "qux"),
				filepath.Join(tmpDir, "foo", "qux", "corge"),
			},
		},
		{
			desc: "multiple_input",
			in: []string{
				tmpDir + "/foo",
				tmpDir + "/bar",
				tmpDir + "/baz",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo"),
				filepath.Join(tmpDir, "bar"),
				filepath.Join(tmpDir, "baz"),
			},
		},
		{
			desc: "mixed_input1",
			in: []string{
				tmpDir + "/foo",
				tmpDir + "/bar/...",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo"),
				filepath.Join(tmpDir, "bar"),
				filepath.Join(tmpDir, "bar", "quux"),
			},
		},
		{
			desc: "mixed_input2",
			in: []string{
				tmpDir + "/foo",
				tmpDir + "/bar/...",
				tmpDir + "/baz/baz.gno",
			},
			expected: []string{
				filepath.Join(tmpDir, "foo"),
				filepath.Join(tmpDir, "bar"),
				filepath.Join(tmpDir, "bar", "quux"),
				filepath.Join(tmpDir, "baz", "baz.gno"),
			},
		},
		{
			desc: "not_exists1",
			in: []string{
				tmpDir + "/notexists", // dir path
			},
			errorShouldContain: "no such file or directory",
		},
		{
			desc: "not_exists2",
			in: []string{
				tmpDir + "/foo/bar.gno", // file path
			},
			errorShouldContain: "no such file or directory",
		},
		{
			desc: "not_exists3", // mixed
			in: []string{
				tmpDir + "/foo",       // exists
				tmpDir + "/notexists", // not exists
			},
			errorShouldContain: "no such file or directory",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			targets, err := targetsFromPatterns(tc.in)
			if tc.errorShouldContain != "" {
				assert.ErrorContains(t, err, tc.errorShouldContain)
				return
			}
			assert.NoError(t, err)
			require.Equal(t, len(tc.expected), len(targets))
			for _, tr := range targets {
				assert.Contains(t, tc.expected, tr)
			}
		})
	}
}

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

func TestResolvePath(t *testing.T) {
	t.Parallel()

	if os.PathSeparator != '/' {
		t.Skip("ResolvePath test is only written of UNIX-like filesystems")
	}
	wd, err := os.Getwd()
	require.NoError(t, err)
	tt := []struct {
		output  string
		dstPath string
		result  string
	}{
		{
			"transpile-result",
			"./examples/test/test1.gno.gen.go",
			"transpile-result/examples/test/test1.gno.gen.go",
		},
		{
			"/transpile-result",
			"./examples/test/test1.gno.gen.go",
			"/transpile-result/examples/test/test1.gno.gen.go",
		},
		{
			"/transpile-result",
			"/home/gno/examples/test/test1.gno.gen.go",
			"/transpile-result/home/gno/examples/test/test1.gno.gen.go",
		},
		{
			"result",
			"../hello",
			filepath.Join("result", filepath.Join(wd, "../hello")),
		},
	}

	for _, tc := range tt {
		res, err := ResolvePath(tc.output, tc.dstPath)
		// ResolvePath should error only in case we can't get the abs path;
		// so never in normal conditions.
		require.NoError(t, err)
		assert.Equal(t,
			tc.result, res,
			"unexpected result of ResolvePath(%q, %q)", tc.output, tc.dstPath,
		)
	}
}
