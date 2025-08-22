package packages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFileKind(t *testing.T) {
	tcs := []struct {
		name     string
		filename string
		body     string
		fileKind FileKind
	}{
		{
			name:     "compiled",
			filename: "foo.gno",
			fileKind: FileKindPackageSource,
		},
		{
			name:     "test",
			filename: "foo_test.gno",
			body:     "package foo",
			fileKind: FileKindTest,
		},
		{
			name:     "test_badpkgclause",
			filename: "foo_test.gno",
			body:     "pakage foo",
			fileKind: FileKindTest,
		},
		{
			name:     "xtest",
			filename: "foo_test.gno",
			body:     "package foo_test",
			fileKind: FileKindXTest,
		},
		{
			name:     "filetest",
			filename: "foo_filetest.gno",
			fileKind: FileKindFiletest,
		},
		{
			name:     "notgnofile",
			filename: "foo.gno.bck",
			fileKind: FileKindOther,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := GetFileKind(tc.filename, tc.body, nil)
			require.Equal(t, tc.fileKind, out)
		})
	}
}
