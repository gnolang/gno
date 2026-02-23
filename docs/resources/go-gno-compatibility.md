# Go - Gno compatibility

Gno is modeled after Go 1.17.

## Reserved keywords

| keyword     | support                |
|-------------|------------------------|
| break       | full                   |
| case        | full                   |
| const       | full                   |
| continue    | full                   |
| default     | full                   |
| defer       | full                   |
| else        | full                   |
| fallthrough | full                   |
| for         | full                   |
| func        | full                   |
| go          | missing (after launch) |
| goto        | full                   |
| if          | full                   |
| import      | full                   |
| interface   | full                   |
| package     | full                   |
| range       | full                   |
| return      | full                   |
| select      | missing (after launch) |
| struct      | full                   |
| switch      | full                   |
| type        | full                   |
| var         | full                   |

Generics are currently not implemented.

Note that Gno does not support shadowing of built-in types.
While the following built-in typecasting assignment would work in Go, this is not supported in Gno.

```go
rune := rune('a')
```

## Builtin types

| type                                          | usage                  | persistency                                                |
|-----------------------------------------------|------------------------|------------------------------------------------------------|
| `bool`                                        | full                   | full                                                       |
| `byte`                                        | full                   | full                                                       |
| `int`, `int8`, `int16`, `int32`, `int64`      | full                   | full                                                       |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | full                   | full                                                       |
| `float32`, `float64`                          | full                   | full                                                       |
| `complex64`, `complex128`                     | missing (TBD)          | missing                                                    |
| `uintptr`, `unsafe.Pointer`                   | missing                | missing                                                    |
| `string`                                      | full                   | full                                                       |
| `rune`                                        | full                   | full                                                       |
| `interface{}` / `any`                         | full                   | full                                                       |
| `[]T` (slices)                                | full                   | full\*                                                     |
| `[N]T` (arrays)                               | full                   | full\*                                                     |
| `map[T1]T2`                                   | full                   | full\*                                                     |
| `func (T1...) T2...`                          | full                   | full (needs more tests)                                    |
| `*T` (pointers)                               | full                   | full\*                                                     |
| `chan T` (channels)                           | missing (after launch) | missing (after launch)                                     |

**\*:** depends on `T`/`T1`/`T2`

Note: for determinism, converting a `string` to `[]byte` or `[]rune` produces a slice with `cap == len`.

Additional builtin types:

| type     | comment                                                                                    |
|----------|--------------------------------------------------------------------------------------------|
| `bigint` | Based on `math/big.Int`                                                                    |
| `bigdec` | Based on https://github.com/cockroachdb/apd, (see https://github.com/gnolang/gno/pull/306) |


## Stdlibs

Legend:

* `nondet`: the standard library in question would require non-deterministic
  behaviour to implement as in Go, such as cryptographical randomness or
  os/network access. The library may still be implemented at one point, with a
  different API.
* `gospec`: the standard library is very Go-specific -- for instance, it is used
  for debugging information or for parsing/build Go source code. A Gno version
  may exist at one point, likely with a different package name or semantics.
* `gnics`: the standard library requires generics.
* `test`: the standard library is currently available for use exclusively in
  test contexts, and may have limited functionality.
* `cmd`: the Go standard library is a command -- a direct equivalent in Gno
  would not be useful.
* `tbd`: whether to include the standard library or not is still up for
  discussion.
* `todo`: the standard library is to be added, and
  [contributions are welcome.](https://github.com/gnolang/gno/issues/1267)
* `part`: the standard library is partially implemented in Gno, and contributions are
  welcome to add the missing functionality.
* `full`: the standard library is fully implemented in Gno.

<!-- generated with: env -C /usr/lib/go/src find -name '*.go' | grep -v _test.go | grep -vE '(internal|vendor|testdata)/' | xargs dirname | sort | uniq | cut -d/ -f2 -->

| package                                     | status   |
|---------------------------------------------|----------|
| archive/tar                                 | `tbd`    |
| archive/zip                                 | `tbd`    |
| arena                                       | `improb` |
| bufio                                       | `full`   |
| builtin                                     | `full`[^1] |
| bytes                                       | `full`   |
| cmd/\*                                      | `cmd`    |
| compress/bzip2                              | `tbd`    |
| compress/flate                              | `tbd`    |
| compress/gzip                               | `tbd`    |
| compress/lzw                                | `tbd`    |
| compress/zlib                               | `tbd`    |
| container/heap                              | `tbd`    |
| container/list                              | `tbd`    |
| container/ring                              | `tbd`    |
| context                                     | `tbd`    |
| crypto                                      | `todo`   |
| crypto/aes                                  | `todo`   |
| crypto/boring                               | `tbd`    |
| crypto/cipher                               | `part`   |
| crypto/des                                  | `tbd`    |
| crypto/dsa                                  | `tbd`    |
| crypto/ecdh                                 | `tbd`    |
| crypto/ecdsa                                | `tbd`    |
| crypto/ed25519                              | `part`[^8] |
| crypto/elliptic                             | `tbd`    |
| crypto/hmac                                 | `todo`   |
| crypto/md5                                  | `test`[^2] |
| crypto/rand                                 | `nondet` |
| crypto/rc4                                  | `tbd`    |
| crypto/rsa                                  | `tbd`    |
| crypto/sha1                                 | `test`[^2] |
| crypto/sha256                               | `part`[^3] |
| crypto/sha512                               | `tbd`    |
| crypto/subtle                               | `tbd`    |
| crypto/tls                                  | `nondet` |
| crypto/tls/fipsonly                         | `nondet` |
| crypto/x509                                 | `tbd`    |
| crypto/x509/pkix                            | `tbd`    |
| database/sql                                | `nondet` |
| database/sql/driver                         | `nondet` |
| debug/buildinfo                             | `gospec` |
| debug/dwarf                                 | `gospec` |
| debug/elf                                   | `gospec` |
| debug/gosym                                 | `gospec` |
| debug/macho                                 | `gospec` |
| debug/pe                                    | `gospec` |
| debug/plan9obj                              | `gospec` |
| embed                                       | `tbd`    |
| encoding                                    | `full`   |
| encoding/ascii85                            | `todo`   |
| encoding/asn1                               | `todo`   |
| encoding/base32                             | `todo`   |
| encoding/base64                             | `full`   |
| encoding/binary                             | `part`   |
| encoding/csv                                | `todo`   |
| encoding/gob                                | `tbd`    |
| encoding/hex                                | `full`   |
| encoding/json                               | `todo`   |
| encoding/pem                                | `todo`   |
| encoding/xml                                | `todo`   |
| errors                                      | `part`   |
| expvar                                      | `tbd`    |
| flag                                        | `nondet` |
| fmt                                         | `test`[^4] |
| go/ast                                      | `gospec` |
| go/build                                    | `gospec` |
| go/build/constraint                         | `gospec` |
| go/constant                                 | `gospec` |
| go/doc                                      | `gospec` |
| go/doc/comment                              | `gospec` |
| go/format                                   | `gospec` |
| go/importer                                 | `gospec` |
| go/parser                                   | `gospec` |
| go/printer                                  | `gospec` |
| go/scanner                                  | `gospec` |
| go/token                                    | `gospec` |
| go/types                                    | `gospec` |
| hash                                        | `full`   |
| hash/adler32                                | `full`   |
| hash/crc32                                  | `todo`   |
| hash/crc64                                  | `todo`   |
| hash/fnv                                    | `todo`   |
| hash/maphash                                | `todo`   |
| html                                        | `full`   |
| html/template                               | `todo`   |
| image                                       | `tbd`    |
| image/color                                 | `tbd`    |
| image/color/palette                         | `tbd`    |
| image/draw                                  | `tbd`    |
| image/gif                                   | `tbd`    |
| image/jpeg                                  | `tbd`    |
| image/png                                   | `tbd`    |
| index/suffixarray                           | `tbd`    |
| io                                          | `full`   |
| io/fs                                       | `tbd`    |
| io/ioutil                                   | removed[^5] |
| log                                         | `tbd`    |
| log/slog                                    | `tbd`    |
| log/syslog                                  | `nondet` |
| maps                                        | `gnics`  |
| math                                        | `full`   |
| math/big                                    | `tbd`    |
| math/bits                                   | `full`   |
| math/cmplx                                  | `tbd`    |
| math/rand                                   | `full`[^9] |
| mime                                        | `tbd`    |
| mime/multipart                              | `tbd`    |
| mime/quotedprintable                        | `tbd`    |
| net                                         | `nondet` |
| net/http                                    | `nondet` |
| net/http/cgi                                | `nondet` |
| net/http/cookiejar                          | `nondet` |
| net/http/fcgi                               | `nondet` |
| net/http/httptest                           | `nondet` |
| net/http/httptrace                          | `nondet` |
| net/http/httputil                           | `nondet` |
| net/http/internal                           | `nondet` |
| net/http/pprof                              | `nondet` |
| net/mail                                    | `nondet` |
| net/netip                                   | `nondet` |
| net/rpc                                     | `nondet` |
| net/rpc/jsonrpc                             | `nondet` |
| net/smtp                                    | `nondet` |
| net/textproto                               | `nondet` |
| net/url                                     | `full`   |
| os                                          | `nondet` |
| os/exec                                     | `nondet` |
| os/signal                                   | `nondet` |
| os/user                                     | `nondet` |
| path                                        | `full`   |
| path/filepath                               | `nondet` |
| plugin                                      | `nondet` |
| reflect                                     | `todo`   |
| regexp                                      | `full`   |
| regexp/syntax                               | `full`   |
| runtime                                     | `gospec` |
| runtime/asan                                | `gospec` |
| runtime/cgo                                 | `gospec` |
| runtime/coverage                            | `gospec` |
| runtime/debug                               | `gospec` |
| runtime/metrics                             | `gospec` |
| runtime/msan                                | `gospec` |
| runtime/pprof                               | `gospec` |
| runtime/race                                | `gospec` |
| runtime/trace                               | `gospec` |
| slices                                      | `gnics`  |
| sort                                        | `part`[^6] |
| strconv                                     | `full`[^10] |
| strings                                     | `full`   |
| sync                                        | `tbd`    |
| sync/atomic                                 | `tbd`    |
| syscall                                     | `nondet` |
| syscall/js                                  | `nondet` |
| testing                                     | `part`   |
| testing/fstest                              | `tbd`    |
| testing/iotest                              | `tbd`    |
| testing/quick                               | `tbd`    |
| text/scanner                                | `todo`   |
| text/tabwriter                              | `todo`   |
| text/template                               | `todo`   |
| text/template/parse                         | `todo`   |
| time                                        | `full`[^7] |
| time/tzdata                                 | `tbd`    |
| unicode                                     | `full`   |
| unicode/utf16                               | `full`   |
| unicode/utf8                                | `full`   |
| unsafe                                      | `nondet` |

[^1]: `builtin` is a "fake" package that exists to document the behaviour of
  some builtin functions. The "fake" package does not currently exist in Gno,
  but [all functions up to Go 1.17 exist](https://pkg.go.dev/builtin@go1.17),
  except for those relating to complex (real or imag) or channel types.
[^2]: `crypto/sha1` and `crypto/md5` implement "deprecated" hashing
  algorithms, widely considered unsafe for cryptographic hashing. Decision on
  whether to include these as part of the official standard libraries is still
  pending.
[^3]: `crypto/sha256` is currently only implemented for `Sum256`, which should
  still cover a majority of use cases. A full implementation is welcome.
[^4]: like many other encoding packages, `fmt` depends on `reflect` to be added.
  For now, package `gno.land/p/nt/ufmt` may do what you need. In test
  functions, `fmt` works.
[^5]: `io/ioutil` [is deprecated in Go.](https://pkg.go.dev/io/ioutil)
  Its functionality has been moved to packages `os` and `io`. The functions
  which have been moved in `io` are implemented in that package.
[^6]: `sort` has the notable omission of `sort.Slice`. You'll need to write a
  bit of boilerplate, but you can use `sort.Interface` + `sort.Sort`!
[^7]: `time.Now` returns the block time rather than the system time, for
  determinism. Concurrent functionality (such as `time.Ticker`) is not implemented.
[^8]: `crypto/ed25519` is currently only implemented for `Verify`, which should
  still cover a majority of use cases. A full implementation is welcome.
[^9]: `math/rand` in Gno ports over Go's `math/rand/v2`.
[^10]: `strconv` does not have the methods relating to types `complex64` and
  `complex128`.

## Tooling (`gno` binary)

| go command        | gno command                  | comment                                                               |
|-------------------|------------------------------|-----------------------------------------------------------------------|
| go bug            | gno bug                      | same behavior                                                         |
| go build          | gno tool transpile -gobuild  | same intention, limited compatibility                                 |
| go clean          | gno clean                    | same intention, limited compatibility                                 |
| go doc            | gno doc                      | limited compatibility; see https://github.com/gnolang/gno/issues/522  |
| go env            | gno env                      |                                                                       |
| go fix            |                              |                                                                       |
| go fmt            | gno fmt                      | gofmt (& similar tools, like gofumpt) works on gno code.              |
| go generate       |                              |                                                                       |
| go get            |                              | see `gno mod download`.                                               |
| go help           | gno $cmd --help              | ie. `gno doc --help`                                                  |
| go install        |                              |                                                                       |
| go list           |                              |                                                                       |
| go mod            | gno mod                      |                                                                       |
| + go mod init     | gno mod init                 | same behavior                                                         |
| + go mod download | gno mod download             | same behavior                                                         |
| + go mod tidy     | gno mod tidy                 | same behavior                                                         |
| + go mod why      | gno mod why                  | same intention                                                        |
|                   | gno tool transpile           |                                                                       |
| go work           |                              |                                                                       |
|                   | gno tool repl                |                                                                       |
| go run            | gno run                      |                                                                       |
| go test           | gno test                     | limited compatibility                                                 |
| go tool           |                              |                                                                       |
| go version        |                              |                                                                       |
| go vet            |                              |                                                                       |
| golint            | gno lint                     | same intention                                                        |
