package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestPackageName(t *testing.T) {
	tests := []struct {
		name  string
		files []*std.MemFile
		want  string
	}{
		{
			name: "derives from first gno file",
			files: []*std.MemFile{
				{Name: "gnomod.toml", Body: "module = \"gno.land/r/x\"\n"},
				{Name: "x.gno", Body: "package x\n\nfunc F() {}"},
			},
			want: "x",
		},
		{
			name: "skips non-gno files",
			files: []*std.MemFile{
				{Name: "README.md", Body: "# not gno"},
				{Name: "foo.gno", Body: "package foo"},
			},
			want: "foo",
		},
		{
			name:  "no gno files yields empty",
			files: []*std.MemFile{{Name: "gnomod.toml", Body: "module = \"x\""}},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, packageName(tt.files))
		})
	}
}
