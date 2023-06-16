// PKGPATH: gno.land/r/demo/showcase_test
package showcase_test

import "gno.land/r/demo/showcase"

func main() {
	println(showcase.Render(""))
}

// Output:
// # Showcase
//
//
// Package showcase provides a framework for creating small modules that showcase numerous details and expose various usages.
//
//
// The primary goal is to facilitate easy demonstration and exploration of small features. Modules self-register, expose commands, and demonstrate different aspects. This framework encourages experimentation and creativity, allowing developers to easily add new modules while keeping the main files untouched.
