package importer

// This file contains "definitions"; it attempts to centralize some common
// answers to common questions like "Is this a gno file?", "What is the import
// path to the gno repository?", "Is this import path of a realm?".

const (
	// RepoImport is the import path to the Gno repository.
	RepoImport = "github.com/gnolang/gno"

	// GnolangImport is the import path to the gnolang package.
	GnolangImport = RepoImport + "/gnovm/pkg/gnolang"
)

// IsGnoFile determines whether the given files matches all of the given patterns,
// with the same matching rules as [MatchPatterns].
//
// It is essentially a helper for MatchPatterns, implicitly adding the patterns
// "*.gno" and "!.*".
//
// IsGnoFile assumes its patterns to be syntactically well-formed; if not, it
// will panic. To test for the correctness of patterns, try passing them with
// any input to MatchPatterns.
func IsGnoFile(name string, patterns ...string) bool {
	m, err := MatchPatterns(name, append(patterns, "*.gno", "!.*")...)
	if err != nil {
		panic(err)
	}
	return m
}
