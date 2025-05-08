// Package modfile defines the structure for Gno package's `gno.toml` module
// files and provides helpers to parse and interact with these files.
//
// The core `Modfile` struct captures metadata such as package path, usage
// constraints, generated runtime metadata, and eventually versioning.
// This structure is designed to be used both on-chain for package management
// within the GnoVM, and off-chain by developer tools for tasks like package
// discovery and metadata inspection.
package modfile
