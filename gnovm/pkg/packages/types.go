package packages

// ported from https://cs.opensource.google/go/go/+/refs/tags/go1.23.2:src/cmd/go/internal/load/pkg.go
type Package struct {
	Dir              string   `json:",omitempty"` // directory containing package sources
	ImportPath       string   `json:",omitempty"` // import path of package in dir
	Name             string   `json:",omitempty"` // package name
	Root             string   `json:",omitempty"` // Gno root, Gno path dir, or module root dir containing this package
	Module           Module   `json:",omitempty"` // info about package's module, if any
	Match            []string `json:",omitempty"` // command-line patterns matching this package
	GnoFiles         []string `json:",omitempty"` // .gno source files (excluding TestGnoFiles, FiletestGnoFiles)
	Imports          []string `json:",omitempty"` // import paths used by this package
	Deps             []string `json:",omitempty"` // all (recursively) imported dependencies
	TestGnoFiles     []string `json:",omitempty"` // _test.gno files in package
	TestImports      []string `json:",omitempty"` // imports from TestGnoFiles
	FiletestGnoFiles []string `json:",omitempty"` // _filetest.gno files in package
	FiletestImports  []string `json:",omitempty"` // imports from FiletestGnoFiles
	Errors           []error  `json:",omitempty"` // error loading this package (not dependencies)
}

// ported from https://cs.opensource.google/go/go/+/refs/tags/go1.23.2:src/cmd/go/internal/modinfo/info.go
type Module struct {
	Path   string `json:",omitempty"` // module path
	Dir    string `json:",omitempty"` // directory holding local copy of files, if any
	GnoMod string `json:",omitempty"` // path to gno.mod file describing module, if any
}
