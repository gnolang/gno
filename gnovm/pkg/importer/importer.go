// Package importer allows to match an import path to a directory, and select
// a set of files within that directory which are Gno files.
package importer

import (
	"sync"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
)

// Config defines the configuration for the importer.
type Config struct {
	// RootDir should point to the directory of the gno repository.
	// This is used to resolve import paths of packages and realms in
	// "examples", and to resolve standard libraries (gnovm/stdlibs and
	// gnovm/tests/stdlibs).
	RootDir string

	// GnoHome is the path to the "home" directory.
	// This is used to resolve imports to on-chain packages.
	GnoHome string

	// If set to true, this will enable the usage of "testing standard
	// libraries"; ie., some stdlibs packages will be additionally resolved
	// to a second directory in gnovm/tests/stdlibs/.
	Test bool
}

var defaultConfig = sync.OnceValue(func() Config {
	return Config{
		RootDir: gnoenv.RootDir(),
		GnoHome: gnoenv.HomeDir(),
	}
})
