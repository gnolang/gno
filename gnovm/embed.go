// Package gnovm embeds the sources of the Gno standard libraries, so that a
// node writes at genesis the stdlibs its native bindings were built from,
// without reading a checkout at run time. It lives here rather than in
// gnovm/stdlibs because every binary links that package for the natives,
// while only the node needs the sources.
package gnovm

import (
	"embed"
	"io/fs"
)

// The patterns select every file gno.ReadMemPackage accepts in a stdlib
// package directory, at the three depths the tree uses (go:embed has no
// "**"): the .gno sources, test files included, gnomod.toml and the package
// READMEs. The Go sources next to them stay out of the binary.
//
//go:embed stdlibs/*/*.gno stdlibs/*/*/*.gno stdlibs/*/*/*/*.gno
//go:embed stdlibs/*/gnomod.toml stdlibs/*/*/gnomod.toml stdlibs/*/*/*/gnomod.toml
//go:embed stdlibs/*/*.md stdlibs/*/*/*.md stdlibs/*/*/*/*.md
var stdlibsFS embed.FS

// Stdlibs returns the stdlib sources, rooted at the stdlibs directory: a
// package's files are under its import path, such as "strings/builder.gno".
func Stdlibs() fs.FS {
	sub, err := fs.Sub(stdlibsFS, "stdlibs")
	if err != nil {
		panic(err) // "stdlibs" is a valid path
	}
	return sub
}
