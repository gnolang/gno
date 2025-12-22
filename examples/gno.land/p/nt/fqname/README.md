# `fqname` - Fully Qualified Name Parser

A utility package for parsing and handling fully qualified identifiers in Gno. Helps split package paths from names in identifiers like `gno.land/p/nt/avl.Tree`.

## Features

- **Package path extraction**: Separate package path from identifier name
- **Robust parsing**: Handles various identifier formats
- **Slash-aware**: Correctly processes paths with forward slashes
- **Dot notation**: Supports standard Go-style package.Name syntax

## Usage

```go
import "gno.land/p/nt/fqname"

// Parse a fully qualified name
pkgpath, name := fqname.Parse("gno.land/p/nt/avl.Tree")
// pkgpath: "gno.land/p/nt/avl"
// name: "Tree"

// Handle simple package names
pkgpath, name = fqname.Parse("fmt.Sprintf")
// pkgpath: "fmt"  
// name: "Sprintf"

// Handle just package path (no name)
pkgpath, name = fqname.Parse("gno.land/p/nt/avl")
// pkgpath: "gno.land/p/nt/avl"
// name: ""
```

## API

```go
// Parse splits a fully qualified identifier into package path and name
func Parse(fqname string) (pkgpath, name string)
```

## Parsing Rules

The parser follows these rules:

1. **With slashes**: Finds the last slash, then looks for a dot in the remaining part
   - `gno.land/p/nt/avl.Tree` → `("gno.land/p/nt/avl", "Tree")`
   - `github.com/user/repo/pkg.Function` → `("github.com/user/repo/pkg", "Function")`

2. **Without slashes**: Looks for the last dot
   - `fmt.Sprintf` → `("fmt", "Sprintf")`
   - `strings.Builder` → `("strings", "Builder")`

3. **No dot found**: Returns entire string as package path
   - `mypackage` → `("mypackage", "")`
   - `gno.land/p/demo/hello` → `("gno.land/p/demo/hello", "")`

## Examples

```go
// Gno standard library
pkgpath, name := fqname.Parse("std.CurrentRealm")
// pkgpath: "std", name: "CurrentRealm"

// Gno package with path
pkgpath, name = fqname.Parse("gno.land/p/demo/avl.Tree")  
// pkgpath: "gno.land/p/demo/avl", name: "Tree"

// Go standard library style
pkgpath, name = fqname.Parse("encoding/json.Marshal")
// pkgpath: "encoding/json", name: "Marshal"

// Complex package path
pkgpath, name = fqname.Parse("gno.land/r/user/app/types.User")
// pkgpath: "gno.land/r/user/app/types", name: "User"

// Just package name
pkgpath, name = fqname.Parse("myutil")
// pkgpath: "myutil", name: ""
```

## Use Cases

- **Import analysis**: Understanding which packages are being referenced
- **Documentation generation**: Separating package paths from symbol names  
- **Code reflection**: Analyzing fully qualified identifiers at runtime
- **Package management**: Processing dependency information
- **Tooling**: Building development tools that work with Gno packages

This package is essential for any tooling or libraries that need to work with fully qualified identifiers in the Gno ecosystem.
