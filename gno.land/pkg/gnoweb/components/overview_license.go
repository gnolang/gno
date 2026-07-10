package components

import (
	"regexp"
	"strings"
)

// licenseSignatures orders detections most-specific-first to avoid false matches.
//
// A signature is matched against the file's title block when the license names
// itself there, and against the whole sample otherwise. Named licenses have to
// be title-scoped: a body routinely cites other licenses, so MPL-2.0 enumerates
// the whole GNU family under "Secondary License" and LGPL quotes the GPL, and a
// body-wide match would report the cited license instead of the real one.
//
// Real license files hard-wrap their text, so a pattern spanning more than one
// word matches its whitespace with `\s+`, and the "s" flag lets "." cross the
// remaining line breaks.
var licenseSignatures = []struct {
	Kind string
	RE   *regexp.Regexp
	// Title restricts the match to the title block rather than the whole sample.
	Title bool
}{
	{Kind: "MIT", RE: regexp.MustCompile(`(?i)^\s*(the )?mit license`), Title: true},
	{Kind: "Apache-2.0", RE: regexp.MustCompile(`(?is)apache\s+license.*version\s+2\.0`), Title: true},
	{Kind: "LGPL", RE: regexp.MustCompile(`(?is)gnu\s+lesser\s+general\s+public\s+license`), Title: true},
	{Kind: "AGPL-3.0", RE: regexp.MustCompile(`(?is)gnu\s+affero\s+general\s+public\s+license.*version\s+3`), Title: true},
	{Kind: "GPL-3.0", RE: regexp.MustCompile(`(?is)gnu\s+general\s+public\s+license.*version\s+3`), Title: true},
	{Kind: "MPL-2.0", RE: regexp.MustCompile(`(?is)mozilla\s+public\s+license.*version\s+2\.0`), Title: true},
	// The third clause is what separates BSD-3 from BSD-2. Anchor on its wording
	// rather than on a "3." label, which real files replace with a bullet.
	{Kind: "BSD-3-Clause", RE: regexp.MustCompile(`(?is)redistribution\s+and\s+use.*with\s+or\s+without\s+modification.*neither\s+the\s+name`)},
	{Kind: "BSD-2-Clause", RE: regexp.MustCompile(`(?is)redistribution\s+and\s+use.*with\s+or\s+without\s+modification`)},
	{Kind: "ISC", RE: regexp.MustCompile(`(?i)isc license`)},
	{Kind: "Unlicense", RE: regexp.MustCompile(`(?is)this\s+is\s+free\s+and\s+unencumbered\s+software`)},
}

var spdxRE = regexp.MustCompile(`(?i)SPDX-License-Identifier:\s*([^\s]+)`)

// titleLines is how many non-empty leading lines make up the title block. Four
// covers a wrapped name plus its version, date and copyright line, and stops
// before the prose where one license may cite another.
const titleLines = 4

// licenseTitle returns the first titleLines non-empty lines of sample, joined by
// a single space so a name wrapped across lines reads as one phrase.
func licenseTitle(sample []byte) string {
	var out []string
	for line := range strings.SplitSeq(string(sample), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
		if len(out) == titleLines {
			break
		}
	}
	return strings.Join(out, " ")
}

// deriveLicense returns the first recognized license file.
// Content is read up to 4 KB to bound regex work and avoid ReDoS surface.
// If the file exists but content lookup fails, FileName is set and Kind is empty.
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

	if m := spdxRE.FindSubmatch(sample); len(m) == 2 {
		return License{Kind: string(m[1]), FileName: licenseFile}
	}
	title := []byte(licenseTitle(sample))
	for _, sig := range licenseSignatures {
		target := sample
		if sig.Title {
			target = title
		}
		if sig.RE.Match(target) {
			return License{Kind: sig.Kind, FileName: licenseFile}
		}
	}
	return License{FileName: licenseFile}
}
