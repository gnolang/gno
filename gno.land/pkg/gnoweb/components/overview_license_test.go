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

func TestDeriveLicense_LineWrappedFiles(t *testing.T) {
	t.Parallel()
	// Headers copied verbatim from real LICENSE files, wrapping included. Every
	// signature has to survive a newline where the upstream text breaks a line.
	const (
		gpl3 = "                    GNU GENERAL PUBLIC LICENSE\n" +
			"                       Version 3, 29 June 2007\n\n" +
			" Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>\n"

		agpl3 = "                    GNU AFFERO GENERAL PUBLIC LICENSE\n" +
			"                       Version 3, 19 November 2007\n"

		bsd3 = "Copyright (c) 2017 The Libc Authors. All rights reserved.\n\n" +
			"Redistribution and use in source and binary forms, with or without\n" +
			"modification, are permitted provided that the following conditions are\n" +
			"met:\n\n" +
			"3. Neither the name of the copyright holder\n"

		bsd2 = "Redistribution and use in source and binary forms, with or without\n" +
			"modification, are permitted provided that the following conditions are met:\n"

		apache2 = "                                 Apache License\n" +
			"                           Version 2.0, January 2004\n"
	)

	// An LGPL body quotes "GNU General Public License" and, further down, another
	// "version 3". That pair is exactly what the GPL-3.0 signature looks for, so
	// LGPL has to be recognised before detection reaches it.
	const lgpl3 = "                   GNU LESSER GENERAL PUBLIC LICENSE\n" +
		"                       Version 3, 29 June 2007\n\n" +
		"  This version of the GNU Lesser General Public License incorporates\n" +
		"the terms and conditions of version 3 of the GNU General Public\n" +
		"License, supplemented by the additional permissions listed below.\n\n" +
		"  0. Additional Definitions.\n\n" +
		"  As used herein, \"this License\" refers to version 3 of the GNU Lesser\n" +
		"General Public License, and the \"GNU GPL\" refers to version 3 of the GNU\n" +
		"General Public License.\n"

	// The MPL body enumerates the whole GNU family under "Secondary License".
	// Detection keys off the title, so none of those names may win here.
	const mpl2 = "Mozilla Public License Version 2.0\n" +
		"==================================\n\n" +
		"1.12. \"Secondary License\"\n" +
		"    means either the GNU General Public License, Version 2.0, the GNU\n" +
		"    Lesser General Public License, Version 2.1, the GNU Affero General\n" +
		"    Public License, Version 3.0, or any later versions of those\n" +
		"    licenses.\n"

	// Real BSD-3 files bullet their third clause instead of numbering it.
	const bsd3Bullet = "Copyright (c) 2017 The Libc Authors. All rights reserved.\n\n" +
		"Redistribution and use in source and binary forms, with or without\n" +
		"modification, are permitted provided that the following conditions are\n" +
		"met:\n\n" +
		"   * Neither the names of the authors nor the names of the\n" +
		"contributors may be used to endorse products.\n"

	tests := []struct {
		name string
		body string
		want string
	}{
		{"GPL-3.0", gpl3, "GPL-3.0"},
		{"AGPL-3.0", agpl3, "AGPL-3.0"},
		{"LGPL", lgpl3, "LGPL"},
		{"MPL-2.0", mpl2, "MPL-2.0"},
		{"BSD-3-Clause", bsd3, "BSD-3-Clause"},
		{"BSD-3-Clause bulleted third clause", bsd3Bullet, "BSD-3-Clause"},
		{"BSD-2-Clause", bsd2, "BSD-2-Clause"},
		{"Apache-2.0", apache2, "Apache-2.0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := deriveLicense([]string{"LICENSE"}, fileContentFn(map[string][]byte{"LICENSE": []byte(tc.body)}))
			require.Equal(t, License{Kind: tc.want, FileName: "LICENSE"}, got)
		})
	}
}
