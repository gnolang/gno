package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sortPackages(t *testing.T) {
	t.Parallel()

	imports := func(imps ...string) map[string]struct{} {
		m := make(map[string]struct{}, len(imps))
		for _, imp := range imps {
			m[imp] = struct{}{}
		}
		return m
	}

	tt := []struct {
		name   string
		pkgs   []*pkgData
		output []string
		panic  string
	}{
		{
			name: "independent",
			pkgs: []*pkgData{
				{importPath: "a"},
				{importPath: "b"},
			},
			output: []string{"a", "b"},
		},
		{
			name: "importExists",
			pkgs: []*pkgData{
				{importPath: "a"},
				{importPath: "b", imports: imports("a")},
			},
			output: []string{"a", "b"},
		},
		{
			name: "reversed",
			pkgs: []*pkgData{
				{importPath: "a", imports: imports("b")},
				{importPath: "b"},
			},
			output: []string{"b", "a"},
		},

		{
			name: "cyclical0",
			pkgs: []*pkgData{
				{importPath: "a", imports: imports("a")},
			},
			panic: `cyclical package initialization on "a" (a -> a)`,
		},
		{
			name: "cyclical1",
			pkgs: []*pkgData{
				{importPath: "a", imports: imports("b")},
				{importPath: "b", imports: imports("a")},
			},
			panic: `cyclical package initialization on "a" (a -> b -> a)`,
		},
		{
			name: "cyclical2",
			pkgs: []*pkgData{
				{importPath: "a", imports: imports("b")},
				{importPath: "b", imports: imports("c")},
				{importPath: "c", imports: imports("a")},
			},
			panic: `cyclical package initialization on "a" (a -> b -> c -> a)`,
		},
		{
			name: "cyclical1_indirect",
			pkgs: []*pkgData{
				{importPath: "a", imports: imports("b")},
				{importPath: "b", imports: imports("c")},
				{importPath: "c", imports: imports("b")},
			},
			panic: `cyclical package initialization on "b" (b -> c -> b)`,
		},

		{
			name: "notFound",
			pkgs: []*pkgData{
				{importPath: "a", imports: imports("b")},
			},
			panic: `package does not exist: "b" (while processing imports from "a")`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.panic != "" {
				assert.PanicsWithError(t, tc.panic, func() {
					sortPackages(tc.pkgs)
				})
				return
			}
			result := sortPackages(tc.pkgs)
			assert.Equal(t, tc.output, result)
		})
	}
}

func Test_sortPackages_integration(t *testing.T) {
	chdir(t, "testdata/sortPackages")

	pkgs, err := walkStdlibs(".")
	require.NoError(t, err)

	order := sortPackages(pkgs)
	assert.Equal(t, []string{"b", "a"}, order)
}
