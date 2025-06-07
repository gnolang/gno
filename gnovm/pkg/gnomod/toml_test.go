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
[module]
path = "gno.land/r/test"
develop = 0
[UploadMetadata]
Uploader = "addr1"
Height = 42
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
				file.Module.Path = "gno.land/r/test"
				file.Gno.Version = "0.9"
				return &file
			}(),
			expected: `
[module]
  path = "gno.land/r/test"

[gno]
  version = "0.9"
`,
		},
		{
			name: "post upload",
			file: func() *File {
				file := File{}
				file.Module.Path = "gno.land/r/test"
				file.Gno.Version = "0.9"
				file.UploadMetadata.Uploader = "addr1"
				file.UploadMetadata.Height = 42
				return &file
			}(),
			expected: `
[module]
  path = "gno.land/r/test"

[gno]
  version = "0.9"

[upload_metadata]
  uploader = "addr1"
  height = 42
`,
		},
		{
			name: "full",
			file: func() *File {
				file := File{}
				file.Module.Path = "gno.land/r/test"
				file.Module.Draft = true
				file.Module.Private = true
				file.Develop.Replace = []Replace{
					{Old: "gno.land/r/test", New: "gno.land/r/test/v2"},
					{Old: "gno.land/r/test/v3", New: "../.."},
				}
				file.Gno.Version = "0.9"
				file.UploadMetadata.Uploader = "addr1"
				file.UploadMetadata.Height = 42
				return &file
			}(),
			expected: `
[module]
  path = "gno.land/r/test"
  draft = true
  private = true

[develop]

  [[develop.replace]]
    old = "gno.land/r/test"
    new = "gno.land/r/test/v2"

  [[develop.replace]]
    old = "gno.land/r/test/v3"
    new = "../.."

[gno]
  version = "0.9"

[upload_metadata]
  uploader = "addr1"
  height = 42
`,
		},
		{
			name: "empty",
			file: &File{},
			expected: `
[module]
  path = ""

[gno]
  version = ""
`,
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
