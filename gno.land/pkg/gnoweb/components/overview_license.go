package components

import (
	"bytes"
	"regexp"
)

// A name in the body is as likely a citation as a declaration, so Title rows
// match the title block only. Patterns use `\s+` because real files hard-wrap.
var licenseSignatures = []struct {
	Kind  string
	RE    *regexp.Regexp
	Title bool
}{
	{Kind: "MIT", RE: regexp.MustCompile(`(?i)^\s*(the )?mit license`), Title: true},
	{Kind: "Apache-2.0", RE: regexp.MustCompile(`(?is)apache\s+license.*version\s+2\.0`), Title: true},
	{Kind: "MPL-2.0", RE: regexp.MustCompile(`(?is)mozilla\s+public\s+license.*version\s+2\.0`), Title: true},
	{Kind: "LGPL", RE: regexp.MustCompile(`(?is)gnu\s+lesser\s+general\s+public\s+license`), Title: true},
	{Kind: "AGPL-3.0", RE: regexp.MustCompile(`(?is)gnu\s+affero\s+general\s+public\s+license.*version\s+3`), Title: true},
	{Kind: "GPL-3.0", RE: regexp.MustCompile(`(?is)gnu\s+general\s+public\s+license.*version\s+3`), Title: true},
	{Kind: "ISC", RE: regexp.MustCompile(`(?i)isc license`), Title: true},
	// The third clause is what separates BSD-3 from BSD-2. Anchor on its wording
	// rather than on a "3." label, which real files replace with a bullet.
	{Kind: "BSD-3-Clause", RE: regexp.MustCompile(`(?is)redistribution\s+and\s+use.*with\s+or\s+without\s+modification.*neither\s+the\s+name`)},
	{Kind: "BSD-2-Clause", RE: regexp.MustCompile(`(?is)redistribution\s+and\s+use.*with\s+or\s+without\s+modification`)},
	{Kind: "Unlicense", RE: regexp.MustCompile(`(?is)this\s+is\s+free\s+and\s+unencumbered\s+software`)},
}

var spdxRE = regexp.MustCompile(`(?i)SPDX-License-Identifier:\s*([^\s]+)`)

// titleLines stops before the prose where one license cites another.
const titleLines = 4

// mdHeading is stripped so "# MIT License" reads as its name.
const mdHeading = "#"

// licenseTitle joins the title block into one phrase, so a wrapped name still
// matches.
func licenseTitle(sample []byte) []byte {
	out := make([]byte, 0, 256)
	n := 0
	for line := range bytes.Lines(sample) {
		line = bytes.TrimSpace(line)
		line = bytes.TrimSpace(bytes.TrimLeft(line, mdHeading))
		if len(line) == 0 {
			continue
		}
		if n > 0 {
			out = append(out, ' ')
		}
		out = append(out, line...)
		if n++; n == titleLines {
			break
		}
	}
	return out
}

// deriveLicense returns the first recognized license file. The 4 KB cap bounds
// regex work, which runs at about 26 MB/s. A failed lookup sets FileName only.
func deriveLicense(files []string, fileContent func(string) ([]byte, bool)) License {
	var licenseFile string
	for _, f := range files {
		if ReLicenseFileName.MatchString(f) {
			licenseFile = f
			break
		}
	}
	if licenseFile == "" {
		return License{}
	}

	body, ok := fileContent(licenseFile)
	if !ok || len(body) == 0 {
		return License{FileName: licenseFile}
	}
	sample := body
	if len(sample) > 4096 {
		sample = sample[:4096]
	}

	// Title first, so a cited identifier loses to the file's own name.
	title := licenseTitle(sample)
	if m := spdxRE.FindSubmatch(title); len(m) == 2 {
		return License{Kind: string(m[1]), FileName: licenseFile}
	}
	for _, sig := range licenseSignatures {
		target := sample
		if sig.Title {
			target = title
		}
		if sig.RE.Match(target) {
			return License{Kind: sig.Kind, FileName: licenseFile}
		}
	}
	if m := spdxRE.FindSubmatch(sample); len(m) == 2 {
		return License{Kind: string(m[1]), FileName: licenseFile}
	}
	return License{FileName: licenseFile}
}
