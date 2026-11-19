package components

import "strings"

// FileClass holds boolean classification flags for a package file name.
// IsGno is true for test files too (they end in ".gno").
type FileClass struct {
	IsGno     bool
	IsTest    bool
	IsReadme  bool
	IsLicense bool
}

// ClassifyFile categorizes a file name once so callers don't re-implement the
// suffix checks. ReLicenseFileName is the shared license-name matcher.
func ClassifyFile(name string) FileClass {
	return FileClass{
		IsGno:     strings.HasSuffix(name, ".gno"),
		IsTest:    strings.HasSuffix(name, "_test.gno") || strings.HasSuffix(name, "_filetest.gno"),
		IsReadme:  name == "README.md",
		IsLicense: ReLicenseFileName.MatchString(name),
	}
}
