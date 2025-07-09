package gnomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalTomlHelper(t *testing.T) {
	cases := []struct {
		name      string
		tomlStr   string
		expectErr bool
	}{
		{
			name: "valid",
			tomlStr: `
module = "gno.land/r/test"
develop = 0
AddPkg = {
  Creator = "addr1"
  Height = 42
}
`,
			expectErr: false,
		},
		{
			name:      "invalid",
			tomlStr:   `not a toml`,
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := parseTomlBytes("gnomod.toml", []byte(tc.tomlStr))
			// XXX
			_ = f
			_ = err
		})
	}
}

func TestMarshalTomlHelper(t *testing.T) {
	cases := []struct {
		name     string
		file     *File
		expected string
	}{
		{
			name: "minimal",
			file: func() *File {
				file := File{}
				file.Module = "gno.land/r/test"
				file.Gno = "0.9"
				return &file
			}(),
			expected: "module = \"gno.land/r/test\"\ngno = \"0.9\"\n",
		},
		{
			name: "post upload",
			file: func() *File {
				file := File{}
				file.Module = "gno.land/r/test"
				file.Gno = "0.9"
				file.AddPkg.Creator = "addr1"
				file.AddPkg.Height = 42
				return &file
			}(),
			expected: "module = \"gno.land/r/test\"\ngno = \"0.9\"\n\n[addpkg]\n  creator = \"addr1\"\n  height = 42\n",
		},
		{
			name: "full",
			file: func() *File {
				file := File{}
				file.Module = "gno.land/r/test"
				file.Ignore = true
				file.Draft = true
				file.Private = true
				file.Replace = []Replace{
					{Old: "gno.land/r/test", New: "gno.land/r/test/v2"},
					{Old: "gno.land/r/test/v3", New: "../.."},
				}
				file.Gno = "0.9"
				file.AddPkg.Creator = "addr1"
				file.AddPkg.Height = 42
				return &file
			}(),
			expected: "module = \"gno.land/r/test\"\ngno = \"0.9\"\nignore = true\ndraft = true\nprivate = true\n\n[[replace]]\n  old = \"gno.land/r/test\"\n  new = \"gno.land/r/test/v2\"\n\n[[replace]]\n  old = \"gno.land/r/test/v3\"\n  new = \"../..\"\n\n[addpkg]\n  creator = \"addr1\"\n  height = 42\n",
		},
		{
			name:     "empty",
			file:     &File{},
			expected: "module = \"\"\ngno = \"\"\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.file.WriteString()
			require.NotEmpty(t, out)
			assert.Equal(t, tc.expected, out)
		})
	}
}
