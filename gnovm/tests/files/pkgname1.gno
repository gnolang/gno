package main

import (
	// pkg name is actually quux, should require an identifier
	"filetests/extern/bar"
)

func main() {
	println("Hello", bar.Quux()) // bar should not be a known symbol.
}

// Error:
// main/pkgname1.gno:5:2-24: package name for "filetests/extern/bar" ("quux") doesn't match its expected identifier "bar"; the import declaration must specify an identifier

// TypeCheckError:
// main/pkgname1.gno:9:19: undefined: bar; main/pkgname1.gno:5:2: "filetests/extern/bar" imported as quux and not used
