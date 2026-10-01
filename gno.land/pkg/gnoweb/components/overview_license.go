package components

import "regexp"

// licenseSignatures orders detections most-specific-first to avoid false matches.
var licenseSignatures = []struct {
	Kind string
	RE   *regexp.Regexp
}{
	{"MIT", regexp.MustCompile(`(?i)^\s*(the )?mit license`)},
	{"Apache-2.0", regexp.MustCompile(`(?i)apache license\s*,?\s*version 2\.0`)},
	{"AGPL-3.0", regexp.MustCompile(`(?i)gnu affero general public license.*version 3`)},
	{"GPL-3.0", regexp.MustCompile(`(?i)gnu general public license.*version 3`)},
	{"LGPL", regexp.MustCompile(`(?i)gnu lesser general public license`)},
	{"BSD-3-Clause", regexp.MustCompile(`(?i)redistribution and use.*with or without modification[\s\S]*3\.\s*neither`)},
	{"BSD-2-Clause", regexp.MustCompile(`(?i)redistribution and use.*with or without modification`)},
	{"ISC", regexp.MustCompile(`(?i)isc license`)},
	{"MPL-2.0", regexp.MustCompile(`(?i)mozilla public license.*version 2\.0`)},
	{"Unlicense", regexp.MustCompile(`(?i)this is free and unencumbered software`)},
}

var spdxRE = regexp.MustCompile(`(?i)SPDX-License-Identifier:\s*([^\s]+)`)

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
	for _, sig := range licenseSignatures {
		if sig.RE.Match(sample) {
			return License{Kind: sig.Kind, FileName: licenseFile}
		}
	}
	return License{FileName: licenseFile}
}
