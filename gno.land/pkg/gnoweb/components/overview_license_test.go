package components

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveLicense(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		files   []string
		content map[string][]byte
		want    License
	}{
		{
			name:  "no license file",
			files: []string{"main.gno", "README.md"},
			want:  License{},
		},
		{
			name:    "MIT by content signature",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("The MIT License\n\nCopyright (c) 2024 ...")},
			want:    License{Kind: "MIT", FileName: "LICENSE"},
		},
		{
			name:    "SPDX identifier takes precedence over signature",
			files:   []string{"LICENSE.md"},
			content: map[string][]byte{"LICENSE.md": []byte("SPDX-License-Identifier: Apache-2.0\n\nThe MIT License text ...")},
			want:    License{Kind: "Apache-2.0", FileName: "LICENSE.md"},
		},
		{
			name:    "unknown license type still surfaces file name",
			files:   []string{"LICENSE.txt"},
			content: map[string][]byte{"LICENSE.txt": []byte("Some custom wording with no known signature")},
			want:    License{Kind: "", FileName: "LICENSE.txt"},
		},
		{
			name:    "file exists but content not fetched",
			files:   []string{"LICENSE"},
			content: nil,
			want:    License{FileName: "LICENSE"},
		},
		{
			name:    "bounded 4KB read ignores late signature",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": append(bytes.Repeat([]byte(" "), 5000), []byte("The MIT License")...)},
			want:    License{Kind: "", FileName: "LICENSE"},
		},
		{
			name:    "Apache detection",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("Apache License, Version 2.0\n\n...")},
			want:    License{Kind: "Apache-2.0", FileName: "LICENSE"},
		},
		{
			name:    "BSD-3-Clause detection",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:\n\n1. ...\n2. ...\n3. Neither the name of the copyright holder ...")},
			want:    License{Kind: "BSD-3-Clause", FileName: "LICENSE"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := deriveLicense(tc.files, fileContentFn(tc.content))
			require.Equal(t, tc.want, got)
		})
	}
}
