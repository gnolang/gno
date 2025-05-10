package gnolang

const (
	GnoVersion = "0.9" // Gno 0.9 is the current version.
)

const (
	// gno.mod files assumed in testing/default contexts.
	gnomodTesting = `go 0.9` // when gno.mod is missing while testing.
	gnomodDefault = `go 0.0` // when gno.mod is missing in general.
)
