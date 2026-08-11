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
			// Shape of examples/gno.land/p/onbloc/json/LICENSE.
			name:    "MIT under a markdown heading",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("# MIT License\n\nCopyright (c) 2019 The Authors\n")},
			want:    License{Kind: "MIT", FileName: "LICENSE"},
		},
		{
			name:    "SPDX in the title block takes precedence over signature",
			files:   []string{"LICENSE.md"},
			content: map[string][]byte{"LICENSE.md": []byte("SPDX-License-Identifier: Apache-2.0\n\nThe MIT License text ...")},
			want:    License{Kind: "Apache-2.0", FileName: "LICENSE.md"},
		},
		{
			// The one input that separates title-SPDX-first from signatures-first:
			// both lines sit in the title block and disagree.
			name:    "SPDX in the title block outranks a signature in it",
			files:   []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte("The MIT License\nSPDX-License-Identifier: Apache-2.0\n")},
			want:    License{Kind: "Apache-2.0", FileName: "LICENSE"},
		},
		{
			name:  "SPDX quoted below the title block loses to the title",
			files: []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte(
				"The MIT License\n" +
					"Copyright (c) 2024 Example\n" +
					"Permission is hereby granted, free of charge, to any person obtaining\n" +
					"a copy of this software and associated documentation files, to deal in\n" +
					"the Software without restriction. Portions are also available under\n" +
					"SPDX-License-Identifier: GPL-3.0\n")},
			want: License{Kind: "MIT", FileName: "LICENSE"},
		},
		{
			// ristretto's z/LICENSE: a file name for a title.
			name:  "SPDX below the title block wins when no signature matched",
			files: []string{"LICENSE"},
			content: map[string][]byte{"LICENSE": []byte(
				"bbloom.go\n" +
					"Copyright 2014 The Authors\n" +
					"All rights reserved.\n" +
					"Licensed under the terms below.\n" +
					"\n * SPDX-License-Identifier: Apache-2.0\n")},
			want: License{Kind: "Apache-2.0", FileName: "LICENSE"},
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

	// The body quotes "GNU General Public License" and another "version 3", the
	// pair GPL-3.0 looks for. Title scoping cuts it before "License".
	const lgpl3 = "                   GNU LESSER GENERAL PUBLIC LICENSE\n" +
		"                       Version 3, 29 June 2007\n\n" +
		"  This version of the GNU Lesser General Public License incorporates\n" +
		"the terms and conditions of version 3 of the GNU General Public\n" +
		"License, supplemented by the additional permissions listed below.\n\n" +
		"  0. Additional Definitions.\n\n" +
		"  As used herein, \"this License\" refers to version 3 of the GNU Lesser\n" +
		"General Public License, and the \"GNU GPL\" refers to version 3 of the GNU\n" +
		"General Public License.\n"

	// Clause 1.12 lists the GNU family, which title scoping keeps out of reach.
	const mpl2 = "Mozilla Public License Version 2.0\n" +
		"==================================\n\n" +
		"1.12. \"Secondary License\"\n" +
		"    means either the GNU General Public License, Version 2.0, the GNU\n" +
		"    Lesser General Public License, Version 2.1, the GNU Affero General\n" +
		"    Public License, Version 3.0, or any later versions of those\n" +
		"    licenses.\n"

	// btcsuite's header. ISC is title-scoped, so its own title must resolve.
	const isc = "ISC License\n\n" +
		"Copyright (c) 2013-2023 The Authors\n" +
		"All rights reserved.\n\n" +
		"Permission to use, copy, modify, and distribute this software for any\n" +
		"purpose with or without fee is hereby granted.\n"

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
		{"ISC", isc, "ISC"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := deriveLicense([]string{"LICENSE"}, fileContentFn(map[string][]byte{"LICENSE": []byte(tc.body)}))
			require.Equal(t, License{Kind: tc.want, FileName: "LICENSE"}, got)
		})
	}
}
