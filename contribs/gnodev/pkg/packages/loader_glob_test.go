package packages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Glob(t *testing.T) {
	const root = "./testdata"
	cases := []struct {
		GlobPath   string
		PkgResults []string
	}{
		{"abc.xy/pkg/*", []string{TestdataPkgA, TestdataPkgB, TestdataPkgC}},
		{"abc.xy/nested/*", []string{TestdataNestedA}},
		{"abc.xy/**/c", []string{TestdataNestedC, TestdataPkgA, TestdataPkgB, TestdataPkgC}},
		{"abc.xy/*/a", []string{TestdataNestedA, TestdataPkgA}},
	}

	fsresolver := NewFSResolver("./testdata")
	globloader := NewGlobLoader("./testdata", fsresolver)

	for _, tc := range cases {
		t.Run(tc.GlobPath, func(t *testing.T) {
			pkgs, err := globloader.Load(tc.GlobPath)
			require.NoError(t, err)
			require.Len(t, pkgs, len(tc.PkgResults))
			for i, expected := range tc.PkgResults {
				assert.Equal(t, expected, pkgs[i].Path)
			}
		})
	}
}
