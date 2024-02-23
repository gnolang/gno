---
id: gnopher-hole-stdlib
---

# Gnopher Hole

## Native bindings

Gno has support for "natively-defined" functions  exclusively within the standard
libraries. These are functions which are _declared_ in Gno code, but only _defined_
in Go. There are generally three reasons why a function should be natively
defined:

1. It relies on inspecting the Gno Virtual Machine itself, i.e. `std.AssertOriginCall`
   or `std.CurrentRealmPath`.
2. It relies on `unsafe`, or other features which are not planned to be
   available in the GnoVM, i.e. `math.Float64frombits`.
3. Its native Go performance significantly outperforms the Gno counterpart by
   several orders of magnitude, and it is used in crucial code or hot paths in
   many programs, i.e. `sha256.Sum256`.

Native bindings are made to be a special feature which can be
help overcome pure Gno limitations, but it is not a substitute for writing
standard libraries in Gno.

There are three components to a natively bound function in Gno:

1. The Gno function declaration, which must be a top-level function with no body
   (and no brackets), i.e. `crypto/sha256/sha256.gno`.
2. The Go function definition, which must be a top-level function with the same
   name and signature, i.e. `crypto/sha256/sha256.go`.
3. When the two above are present and valid, the native binding can be created
   by executing the code generator: either by executing `go generate` from the
   `stdlibs` directory, or run `make generate` from the `gnovm` directory.
   This generates the `native.go` file available in the `stdlibs` directory,
   which provides the binding itself to then be used by the GnoVM.

The code generator in question is available in the `misc/genstd` directory.
There are some quirks and features that must be kept in mind when writing native
bindings, which are the following:

- Unexported functions (i.e. `func sum256(b []byte)`) must have their
  Go counterpart prefixed with `X_` in order to make the functions exported (i.e.
  `func X_sum256(b []byte)`).
- The Go function declaration may specify as the first argument
  `m *gno.Machine`, where `gno` is an import for
  `github.com/gnolang/gno/gnovm/pkg/gnolang`. This gives the function access to
  the Virtual Machine state, and is used by functions like `std.AssertOriginCall()`.
- The Go function may change the type of any parameter or result to
  `gno.TypedValue`, where `gno` is an import for the above import path. This
  means that the `native.go` generated code will not attempt to automatically
  convert the Gno value into the Go value, and can be useful for unsupported
  conversions like interface values.
- A small set of named types are "linked" between their Gno version and Go
  counterpart. For instance, `std.Address` in Gno is
  `(".../tm2/pkg/crypto").Bech32Address` in Go. A list of these can be found in
  `misc/genstd/mapping.go`.
- Not all type literals are currently supported when converting from their Gno
  version to their Go counterpart, i.e. `struct` and `map` literals. If you intend to use these,
  modify the code generator to support them.
- The code generator does not inspect any imported packages from the Go native code
  to determine the default package identifier (i.e. the `package` clause).
  For example, if a package is in `foo/bar`, but declares `package xyz`, when importing
  foo/bar the generator will assume the name to be `bar` instead of `xyz`.
  You can add an identifier to the import to fix this and use the identifier
  you want/need, such as `import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"`.

## Adding new standard libraries

New standard libraries may be added by simply creating a new directory (whose
path relative to the `stdlibs` directory will be the import path used in Gno
programs). Following that, the suggested approach for adding a Go standard
library is to copy the original files from the Go source tree, and renaming their
extensions from `.go` to `.gno`.

:::note
As a small aid, this bash one-liner can be useful to convert all the file
extensions:
```sh
for i in *.go; do mv $i "$(echo $i | sed 's/\.go$/.gno/')"; done
```
:::  

Following that, the suggested approach is to iteratively try running `gno test .`,
while fixing any errors that may come out of trying to test the package.

Some things to keep in mind:

- Gno doesn't support assembly functions and build tags. Some Go packages may
  contain assembly versions for different architecture and a `generic.go` file
  containing the architecture-independent version. The general approach is that
  of removing everything architecture/os-specific except for the `generic.go` file.
- Gno doesn't support reflection at the time of writing, which means that for
  now many packages which rely heavily on reflection have to be delayed or
  reduced while we figure out the details on how to implement reflection.
  Aside from the `reflect` package itself, this also translates to very common
  packages still not available in Gno, such as `fmt` or `encoding/json`.
- In the package documentation, specify the Go version from which the library
  was taken.
- All changes from the Go standard libraries must be explicitly marked, possibly
  with `// XXX` comments as needed.

If you intend to create a PR to add a new standard library, remember to update
[Go<\>Gno compatibility](../../reference/go-gno-compatibility.md) accordingly.


