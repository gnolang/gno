package packages

// Package packages provides utility functions to statically analyze Gno
// packages using the Go go/* libraries.  Since Gno currently uses the Go
// parser and type-checker before the Gno preprocessor, if the Go AST provides
// everything you need (such as in pkg/gnolang/gnomod.go ReadPkgListFromDir(),
// this maybe the package you want. For pure Gno & std.MemPackage logic see
// pkg/gnolang/mempackage.go.
