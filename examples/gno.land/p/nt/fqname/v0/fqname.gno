// Package fqname provides utilities for handling fully qualified identifiers in
// Gno. A fully qualified identifier typically includes a package path followed
// by a dot (.) and then the name of a variable, function, type, or other
// package-level declaration.
package fqname

import (
	"strings"
)

// Parse splits a fully qualified identifier into its package path and name
// components. It handles cases with and without slashes in the package path.
//
//	pkgpath, name := fqname.Parse("gno.land/p/nt/avl/v0.Tree")
//	ufmt.Sprintf("Package: %s, Name: %s\n", id.Package, id.Name)
//	// Output: Package: gno.land/p/nt/avl/v0, Name: Tree
func Parse(fqname string) (pkgpath, name string) {
	// Find the index of the last slash.
	lastSlashIndex := strings.LastIndex(fqname, "/")
	if lastSlashIndex == -1 {
		// No slash found, handle it as a simple package name with dot notation.
		dotIndex := strings.LastIndex(fqname, ".")
		if dotIndex == -1 {
			return fqname, ""
		}
		return fqname[:dotIndex], fqname[dotIndex+1:]
	}

	// Get the part after the last slash.
	afterSlash := fqname[lastSlashIndex+1:]

	// Check for a dot in the substring after the last slash.
	dotIndex := strings.Index(afterSlash, ".")
	if dotIndex == -1 {
		// No dot found after the last slash
		return fqname, ""
	}

	// Split at the dot to separate the base and the suffix.
	base := fqname[:lastSlashIndex+1+dotIndex]
	suffix := afterSlash[dotIndex+1:]

	return base, suffix
}

// Construct a qualified identifier.
//
//	fqName := fqname.Construct("gno.land/r/demo/foo20", "Token")
//	fmt.Println("Fully Qualified Name:", fqName)
//	// Output: gno.land/r/demo/foo20.Token
func Construct(pkgpath, name string) string {
	// TODO: ensure pkgpath is valid - and as such last part does not contain a dot.
	if name == "" {
		return pkgpath
	}
	return pkgpath + "." + name
}

// RenderLink creates a formatted link for a fully qualified identifier.
// If the package path starts with "gno.land", it converts it to a markdown link.
// If the domain is different or missing, it returns the input as is.
func RenderLink(pkgPath, slug string) string {
	if strings.HasPrefix(pkgPath, "gno.land") {
		pkgLink := strings.TrimPrefix(pkgPath, "gno.land")
		if slug != "" {
			return "[" + pkgPath + "](" + pkgLink + ")." + slug
		}

		return "[" + pkgPath + "](" + pkgLink + ")"
	}

	if slug != "" {
		return pkgPath + "." + slug
	}

	return pkgPath
}
