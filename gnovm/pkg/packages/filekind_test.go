package packages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFileKind(t *testing.T) {
	tcs := []struct {
		name          string
		filename      string
		body          string
		fileKind      FileKind
		errorContains string
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
			name:          "err_badpkgclause",
			filename:      "foo_test.gno",
			body:          "pakage foo",
			errorContains: "foo_test.gno:1:1: expected 'package', found pakage",
		},
		{
			name:          "err_notgnofile",
			filename:      "foo.gno.bck",
			errorContains: `foo.gno.bck:1:1: not a gno file`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out, err := GetFileKind(tc.filename, tc.body, nil)
			if len(tc.errorContains) != 0 {
				require.ErrorContains(t, err, tc.errorContains)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.fileKind, out)
		})
	}
}
