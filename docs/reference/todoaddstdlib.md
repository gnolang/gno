---
id: standard-library
---

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
Go standard libarary on [Go<\>Gno compatibility](go-gno-compatibility.md).


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


