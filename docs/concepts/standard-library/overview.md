---
id: overview
---

# Overview

Gno comes with a set of standard libraries which are included to ease development and provide extended functionality to the language. These include:
- standard libraries as we know them in classic Golang, i.e. `encoding/binary`, `strings`, `testing`, etc.
- a special `std` package, which contains types and APIs created to handle blockchain-related functionality.

Standard libraries are distinguishable from on-chain [packages](../packages.md) by having  a domain-like structure as the
first part of their import path. For instance: 
- `import "encoding/binary"` refers to a standard library
- `import "gno.land/p/demo/avl"` refers to an on-chain package.

To see concrete implementation details & API references, see the [reference](../../reference/standard-library.md) section.

## Accessing documentation

Apart from the official documentation you are currently reading, you can also access documentation for the standard
libraries in several other different ways. You can obtain a list of all the available standard libraries with the following commands:

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

All the packages have automatically generated documentation through the use of
`gno doc` command, which has similar functionality and features to `go doc`:

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