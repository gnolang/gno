package main

import (
	// pkg name is actually quux, should require an identifier
	"github.com/gnolang/gno/_test/bar"
)

func main() {
	println("Hello", bar.Quux()) // bar should not be a known symbol.
}

// Error:
// main/files/pkgname1.gno:5:2: package name for "github.com/gnolang/gno/_test/bar" ("quux") doesn't match its expected identifier "bar"; the import declaration must specify an identifier

// TypeCheckError:
// main/files/pkgname1.gno:9:19: undefined: bar; main/files/pkgname1.gno:5:2: "github.com/gnolang/gno/_test/bar" imported as quux and not used
