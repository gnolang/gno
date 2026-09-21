package pkgdownload

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryFetcher_Hit(t *testing.T) {
	pkg := &std.MemPackage{
		Name: "foo",
		Path: "gno.land/p/demo/foo",
		Files: []*std.MemFile{
			{Name: "foo.gno", Body: "package foo"},
		},
	}
	f := NewInMemoryFetcher(pkg)

	files, err := f.FetchPackage("gno.land/p/demo/foo")
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "foo.gno", files[0].Name)
	assert.Equal(t, "package foo", files[0].Body)
}

func TestInMemoryFetcher_Miss(t *testing.T) {
	f := NewInMemoryFetcher()
	_, err := f.FetchPackage("gno.land/p/unknown")
	assert.Error(t, err)
}
