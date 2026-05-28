package packages

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
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

// TestGetMemFileKind asserts GetMemFileKind prefers the explicit Kind field
// (which carries new-style filetests with bare basenames) and falls back to
// GetFileKind for legacy MemFiles with Kind unset.
func TestGetMemFileKind(t *testing.T) {
	tcs := []struct {
		name string
		mf   std.MemFile
		want FileKind
	}{
		{"explicit_filetest_no_suffix", std.MemFile{Name: "foo.gno", Kind: std.KindFiletest}, FileKindFiletest},
		{"explicit_source", std.MemFile{Name: "foo.gno", Kind: std.KindPackageSource}, FileKindPackageSource},
		{"explicit_xtest", std.MemFile{Name: "foo_test.gno", Body: "package foo", Kind: std.KindXTest}, FileKindXTest},
		{"unknown_legacy_filetest_suffix", std.MemFile{Name: "foo_filetest.gno", Kind: std.KindUnknown}, FileKindFiletest},
		{"unknown_falls_back_to_name", std.MemFile{Name: "foo.gno", Kind: std.KindUnknown}, FileKindPackageSource},
		{"explicit_test_needs_body_parse", std.MemFile{Name: "foo_test.gno", Body: "package foo_test", Kind: std.KindTest}, FileKindXTest},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := GetMemFileKind(&tc.mf, nil)
			require.Equal(t, tc.want, out)
		})
	}
}
