# Standard Libraries

Gno comes with a set of standard libraries which are included whenever you
execute Gno code. These are distinguishable from imports of packages from the
chain by not referring to a "domain" as the first part of their import path. For
instance, `import "encoding/binary"` refers to a standard library, while
`import "gno.land/p/demo/avl"` refers to an on-chain package.

Standard libraries packages follow the same semantics as on-chain packages (ie.
they don't persist state like realms do) and come as a part of the Gno
programming language rather than with the Gno.land chain.

Many standard libaries are near-identical copies of the equivalent Go standard
libraries; in fact, you can check the current status of implementation of each
Go standard libarary on [Go\<\>Gno compatibility](go-gno-compatibility.md).

## Gathering documentation

At the time being, there is no "list" of the available standard libraries
available from Gno tooling or documentation, but you can obtain a list of all
the available packages with the following commands:

```console
$ cd gnovm/stdlibs # go to correct directory
$ find -type d
./testing
./math
./crypto
./crypto/chacha20
./crypto/chacha20/chacha
./crypto/chacha20/rand
./crypto/sha256
./crypto/cipher
...
```

All of the packages have automatic, generated documentation through the use of
`gno doc`, which has similar functionality and features to `go doc`:

```console
$ gno doc encoding/binary
package binary // import "encoding/binary"

Package binary implements simple translation between numbers and byte sequences
and encoding and decoding of varints.

[...]

var BigEndian bigEndian
var LittleEndian littleEndian
type AppendByteOrder interface{ ... }
type ByteOrder interface{ ... }
$ gno doc -u -src encoding/binary littleEndian.AppendUint16
package binary // import "encoding/binary"

func (littleEndian) AppendUint16(b []byte, v uint16) []byte {
        return append(b,
                byte(v),
                byte(v>>8),
        )
}
```

`gno doc` will work automatically when used within the Gno repository or any
repository which has a `go.mod` dependency on `github.com/gnolang/gno`, which
can be a simple way to set up your Gno repositories to automatically support
`gno` commands (aside from `doc`, also `test`, `run`, etc.).

Another alternative is setting your enviornment variable `GNOROOT` to point to
where you cloned the Gno repository. You can set this in your `~/.profile` file
to be automatically set up in your console:

```sh
export GNOROOT=$HOME/gno
```

## Test standard libraries

There are some additional standard library functions and packages which are
currently available only in `_test.gno` and `_filetest.gno` files. At the time
of writing, these are only some additions in the `std` package to support
changing some values in test functions.

`gno doc` currently doesn't support reading from the test standard libraries,
though support is planned to be added. For now, you can inspect the directory
`gnovm/tests/stdlibs`.

## Adding new standard libraries

New standard libraries may be added by simply creating a new directory (whose
path relative to the `stdlibs` directory will be the import path used in Gno
programs). Following that, the suggested approach for adding a Go standard
libary is to copy the original files from the Go source tree, and renaming their
extensions from `.go` to `.gno`.

> As a small aid, this bash one-liner can be useful to convert all the file
> extensions:
>
> ```sh
> for i in *.go; do mv $i "$(echo $i | sed 's/\.go$/.gno/')"; done
> ```

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
- All changes from the Go standard libaries must be explicitly marked, possibly
  with `// XXX` comments as needed.

If you intend to create a PR to add a new standard library, remember to update
[Go\<\>Gno compatibility](go-gno-compatibility.md) accordingly.

## Native bindings

Gno has support for "natively-defined functions" exclusively within the standard
libaries. These are functions which are _declared_ in Gno code, but only _defined_
in Go. There are generally three reasons why a function should be natively
defined:

1. It relies on inspecting the Gno Virtual Machine itself.\
   For example: `std.AssertOriginCall`, `std.CurrentRealmPath`.
2. It relies on `unsafe`, or other features which are not planned to be
   available in the GnoVM.\
   For example: `math.Float64frombits`.
3. Its native Go performance significantly outperforms the Gno counterpart by
   several orders of magnitude, and it is used in crucial code or hot paths in
   many programs.\
   For example: `sha256.Sum256`.

The takeaway here is that native bindings are a special feature which can be
useful to overcome pure Gno limitations, but it is not a substitute for writing
standard libaries in Gno.

There are three components to a natively bound function in Gno:

1. The Gno function declaration, which must be a top-level function with no body
   (and no brackets).\
   For example: `crypto/sha256/sha256.gno`.
2. The Go function definition, which must be a top-level function with the same
   name and signature.\
   For example: `crypto/sha256/sha256.go`.
3. When the two above are present and valid, the native binding can be created
   by executing the code generator: either execute `go generate` from the
   `stdlibs` directory, or run `make generate` from the `gnovm` directory.\
   This generates the `native.go` file available in the `stdlibs` directory,
   which provides the binding itself to then be used by the GnoVM.

The code generator in question is available in the `misc/genstd` directory.
There are some quirks and features that must be kept in mind when writing native
bindings, which are the following:

- Unexported functions (for instance, `func sum256(b []byte)`) must have their
  Go counterpart prefixed with `X_` in order to make the functions exported (ie.
  `func X_sum256(b []byte)`).
- The Go function declaration may specify as the first argument
  `m *gno.Machine`, where `gno` is an import for
  `github.com/gnolang/gno/gnovm/pkg/gnolang`. This gives the function access to
  the Virtual Machine state, and is used by functions like `std.AssertOriginCall()`.
- The Go function may change the type of any parameter or result to
  `gno.TypedValue` (where `gno` is an import for the above import path). This
  means that the `native.go` generated code will not attempt to automatically
  convert the Gno value into the Go value, and can be useful for unsupported
  conversions like interface values.
- A small set of named types are "linked" between their Gno version and Go
  counterpart. For instance, `std.Address` in Gno is
  `(".../tm2/pkg/crypto").Bech32Address` in Go. A list of these can be found in
  `misc/genstd/mapping.go`.
- Not all type literals are currently supported when converting from their Gno
  version to their Go counterpart. Notable omissions at the time of writing
  include struct and map literals. If you intend to use these, modify the code
  generator to support them.
- The code generator does not inspect any imported packages from the Go native code
  to determine the default package identifier (ie. the `package` clause).
  Ie. if a package is in `foo/bar`, but declares `package xyz`, when importing
  foo/bar the generator will assume the name to be `bar` instead of `xyz`.
  You can add an identifier to the import to fix this and use the identifier
  you want/need, ie.: `import gno "github.com/gnolang/gno/gnovm/pkg/gnolang"`.
