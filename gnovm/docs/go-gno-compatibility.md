# Go<>Gno compatibility

**WIP: does not reflect the current state yet.**

## Native keywords

Legend: full, partial, missing, TBD.

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

## Native types

| type                                          | usage                  | persistency                                                |
|-----------------------------------------------|------------------------|------------------------------------------------------------|
| `bool`                                        | full                   | full                                                       |
| `byte`                                        | full                   | full                                                       |
| `float32`, `float64`                          | full                   | full                                                       |
| `int`, `int8`, `int16`, `int32`, `int64`      | full                   | full                                                       |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | full                   | full                                                       |
| `string`                                      | full                   | full                                                       |
| `rune`                                        | full                   | full                                                       |
| `interface{}`                                 | full                   | full                                                       |
| `[]T` (slices)                                | full                   | full*                                                      |
| `map[T1]T2`                                   | full                   | full*                                                      |
| `func (T1...) T2...`                          | full                   | full (needs more tests)                                    |
| `*T` (pointers)                               | full                   | full*                                                      |
| `chan T` (channels)                           | missing (after launch) | missing (after launch)                                     |

**\*:** depends on `T`/`T1`/`T2`

Additional native types:

| type     | comment                                                                                    |
|----------|--------------------------------------------------------------------------------------------|
| `bigint` | Based on `math/big.Int`                                                                    |
| `bigdec` | Based on https://github.com/cockroachdb/apd, (see https://github.com/gnolang/gno/pull/306) |


## Stdlibs

<!-- generated with: find . -name "*.go" | grep -v _test.go | grep -v internal/ | grep -v vendor/ | xargs dirname | sort | uniq -->

| package                                     | status   |
|---------------------------------------------|----------|
| archive/tar                                 | TBD      |
| archive/zip                                 | TBD      |
| arena                                       | TBD      |
| bufio                                       | TBD      |
| builtin                                     | TBD      |
| bytes                                       | TBD      |
| cmd/addr2line                               | TBD      |
| cmd/api                                     | TBD      |
| cmd/api/testdata/src/issue21181/dep         | TBD      |
| cmd/api/testdata/src/issue21181/indirect    | TBD      |
| cmd/api/testdata/src/issue21181/p           | TBD      |
| cmd/api/testdata/src/pkg/p1                 | TBD      |
| cmd/api/testdata/src/pkg/p2                 | TBD      |
| cmd/api/testdata/src/pkg/p3                 | TBD      |
| cmd/api/testdata/src/pkg/p4                 | TBD      |
| cmd/asm                                     | TBD      |
| cmd/buildid                                 | TBD      |
| cmd/cgo                                     | TBD      |
| cmd/compile                                 | TBD      |
| cmd/covdata                                 | TBD      |
| cmd/covdata/testdata                        | TBD      |
| cmd/cover                                   | TBD      |
| cmd/cover/testdata                          | TBD      |
| cmd/cover/testdata/html                     | TBD      |
| cmd/cover/testdata/pkgcfg/a                 | TBD      |
| cmd/cover/testdata/pkgcfg/b                 | TBD      |
| cmd/cover/testdata/pkgcfg/main              | TBD      |
| cmd/dist                                    | TBD      |
| cmd/distpack                                | TBD      |
| cmd/doc                                     | TBD      |
| cmd/doc/testdata                            | TBD      |
| cmd/doc/testdata/merge                      | TBD      |
| cmd/doc/testdata/nested                     | TBD      |
| cmd/doc/testdata/nested/empty               | TBD      |
| cmd/doc/testdata/nested/nested              | TBD      |
| cmd/fix                                     | TBD      |
| cmd/go                                      | TBD      |
| cmd/gofmt                                   | TBD      |
| cmd/go/testdata                             | TBD      |
| cmd/link                                    | TBD      |
| cmd/link/testdata/pe-binutils               | TBD      |
| cmd/link/testdata/pe-llvm                   | TBD      |
| cmd/link/testdata/testBuildFortvOS          | TBD      |
| cmd/link/testdata/testHashedSyms            | TBD      |
| cmd/link/testdata/testIndexMismatch         | TBD      |
| cmd/link/testdata/testRO                    | TBD      |
| cmd/nm                                      | TBD      |
| cmd/objdump                                 | TBD      |
| cmd/objdump/testdata                        | TBD      |
| cmd/objdump/testdata/testfilenum            | TBD      |
| cmd/pack                                    | TBD      |
| cmd/pprof                                   | TBD      |
| cmd/pprof/testdata                          | TBD      |
| cmd/test2json                               | TBD      |
| cmd/trace                                   | TBD      |
| cmd/vet                                     | TBD      |
| cmd/vet/testdata/asm                        | TBD      |
| cmd/vet/testdata/assign                     | TBD      |
| cmd/vet/testdata/atomic                     | TBD      |
| cmd/vet/testdata/bool                       | TBD      |
| cmd/vet/testdata/buildtag                   | TBD      |
| cmd/vet/testdata/cgo                        | TBD      |
| cmd/vet/testdata/composite                  | TBD      |
| cmd/vet/testdata/copylock                   | TBD      |
| cmd/vet/testdata/deadcode                   | TBD      |
| cmd/vet/testdata/directive                  | TBD      |
| cmd/vet/testdata/httpresponse               | TBD      |
| cmd/vet/testdata/lostcancel                 | TBD      |
| cmd/vet/testdata/method                     | TBD      |
| cmd/vet/testdata/nilfunc                    | TBD      |
| cmd/vet/testdata/print                      | TBD      |
| cmd/vet/testdata/rangeloop                  | TBD      |
| cmd/vet/testdata/shift                      | TBD      |
| cmd/vet/testdata/structtag                  | TBD      |
| cmd/vet/testdata/tagtest                    | TBD      |
| cmd/vet/testdata/testingpkg                 | TBD      |
| cmd/vet/testdata/unmarshal                  | TBD      |
| cmd/vet/testdata/unsafeptr                  | TBD      |
| cmd/vet/testdata/unused                     | TBD      |
| compress/bzip2                              | TBD      |
| compress/flate                              | TBD      |
| compress/gzip                               | TBD      |
| compress/lzw                                | TBD      |
| compress/zlib                               | TBD      |
| container/heap                              | TBD      |
| container/list                              | TBD      |
| container/ring                              | TBD      |
| context                                     | TBD      |
| crypto                                      | TBD      |
| crypto/aes                                  | TBD      |
| crypto/boring                               | TBD      |
| crypto/cipher                               | TBD      |
| crypto/des                                  | TBD      |
| crypto/dsa                                  | TBD      |
| crypto/ecdh                                 | TBD      |
| crypto/ecdsa                                | TBD      |
| crypto/ed25519                              | TBD      |
| crypto/elliptic                             | TBD      |
| crypto/hmac                                 | TBD      |
| crypto/md5                                  | TBD      |
| crypto/rand                                 | TBD      |
| crypto/rc4                                  | TBD      |
| crypto/rsa                                  | TBD      |
| crypto/sha1                                 | TBD      |
| crypto/sha256                               | TBD      |
| crypto/sha512                               | TBD      |
| crypto/subtle                               | TBD      |
| crypto/tls                                  | TBD      |
| crypto/tls/fipsonly                         | TBD      |
| crypto/x509                                 | TBD      |
| crypto/x509/pkix                            | TBD      |
| database/sql                                | TBD      |
| database/sql/driver                         | TBD      |
| debug/buildinfo                             | TBD      |
| debug/dwarf                                 | TBD      |
| debug/elf                                   | TBD      |
| debug/gosym                                 | TBD      |
| debug/gosym/testdata                        | TBD      |
| debug/macho                                 | TBD      |
| debug/pe                                    | TBD      |
| debug/plan9obj                              | TBD      |
| embed                                       | TBD      |
| encoding                                    | TBD      |
| encoding/ascii85                            | TBD      |
| encoding/asn1                               | TBD      |
| encoding/base32                             | TBD      |
| encoding/base64                             | TBD      |
| encoding/binary                             | partial  |
| encoding/csv                                | TBD      |
| encoding/gob                                | TBD      |
| encoding/hex                                | TBD      |
| encoding/json                               | TBD      |
| encoding/pem                                | TBD      |
| encoding/xml                                | TBD      |
| errors                                      | TBD      |
| expvar                                      | TBD      |
| flag                                        | TBD      |
| fmt                                         | TBD      |
| go/ast                                      | TBD      |
| go/build                                    | TBD      |
| go/build/constraint                         | TBD      |
| go/build/testdata/alltags                   | TBD      |
| go/build/testdata/cgo_disabled              | TBD      |
| go/build/testdata/directives                | TBD      |
| go/build/testdata/doc                       | TBD      |
| go/build/testdata/multi                     | TBD      |
| go/build/testdata/non_source_tags           | TBD      |
| go/build/testdata/other                     | TBD      |
| go/build/testdata/other/file                | TBD      |
| go/constant                                 | TBD      |
| go/doc                                      | TBD      |
| go/doc/comment                              | TBD      |
| go/doc/testdata                             | TBD      |
| go/doc/testdata/examples                    | TBD      |
| go/doc/testdata/pkgdoc                      | TBD      |
| go/format                                   | TBD      |
| go/importer                                 | TBD      |
| go/parser                                   | TBD      |
| go/parser/testdata/goversion                | TBD      |
| go/parser/testdata/issue42951               | TBD      |
| go/parser/testdata/issue42951/not_a_file.go | TBD      |
| go/printer                                  | TBD      |
| go/printer/testdata                         | TBD      |
| go/scanner                                  | TBD      |
| go/token                                    | TBD      |
| go/types                                    | TBD      |
| go/types/testdata                           | TBD      |
| go/types/testdata/local                     | TBD      |
| hash                                        | TBD      |
| hash/adler32                                | TBD      |
| hash/crc32                                  | TBD      |
| hash/crc64                                  | TBD      |
| hash/fnv                                    | TBD      |
| hash/maphash                                | TBD      |
| html                                        | TBD      |
| html/template                               | TBD      |
| image                                       | TBD      |
| image/color                                 | TBD      |
| image/color/palette                         | TBD      |
| image/draw                                  | TBD      |
| image/gif                                   | TBD      |
| image/jpeg                                  | TBD      |
| image/png                                   | TBD      |
| index/suffixarray                           | TBD      |
| io                                          | TBD      |
| io/fs                                       | TBD      |
| io/ioutil                                   | TBD      |
| log                                         | TBD      |
| log/internal                                | TBD      |
| log/slog                                    | TBD      |
| log/slog/internal                           | TBD      |
| log/syslog                                  | TBD      |
| maps                                        | TBD      |
| math                                        | partial      |
| math/big                                    | TBD      |
| math/bits                                   | TBD      |
| math/cmplx                                  | TBD      |
| math/rand                                   | TBD      |
| mime                                        | TBD      |
| mime/multipart                              | TBD      |
| mime/quotedprintable                        | TBD      |
| net                                         | TBD      |
| net/http                                    | TBD      |
| net/http/cgi                                | TBD      |
| net/http/cookiejar                          | TBD      |
| net/http/fcgi                               | TBD      |
| net/http/httptest                           | TBD      |
| net/http/httptrace                          | TBD      |
| net/http/httputil                           | TBD      |
| net/http/internal                           | TBD      |
| net/http/pprof                              | TBD      |
| net/mail                                    | TBD      |
| net/netip                                   | TBD      |
| net/rpc                                     | TBD      |
| net/rpc/jsonrpc                             | TBD      |
| net/smtp                                    | TBD      |
| net/textproto                               | TBD      |
| net/url                                     | TBD      |
| os                                          | TBD      |
| os/exec                                     | TBD      |
| os/signal                                   | TBD      |
| os/user                                     | TBD      |
| path                                        | TBD      |
| path/filepath                               | TBD      |
| plugin                                      | TBD      |
| reflect                                     | TBD      |
| regexp                                      | TBD      |
| regexp/syntax                               | TBD      |
| runtime                                     | TBD      |
| runtime/asan                                | TBD      |
| runtime/cgo                                 | TBD      |
| runtime/coverage                            | TBD      |
| runtime/coverage/testdata                   | TBD      |
| runtime/coverage/testdata/issue56006        | TBD      |
| runtime/debug                               | TBD      |
| runtime/metrics                             | TBD      |
| runtime/msan                                | TBD      |
| runtime/pprof                               | TBD      |
| runtime/pprof/testdata/mappingtest          | TBD      |
| runtime/race                                | TBD      |
| runtime/race/testdata                       | TBD      |
| runtime/testdata/testexithooks              | TBD      |
| runtime/testdata/testfaketime               | TBD      |
| runtime/testdata/testprog                   | TBD      |
| runtime/testdata/testprogcgo                | TBD      |
| runtime/testdata/testprogcgo/windows        | TBD      |
| runtime/testdata/testprognet                | TBD      |
| runtime/testdata/testwinlib                 | TBD      |
| runtime/testdata/testwinlibsignal           | TBD      |
| runtime/testdata/testwinlibthrow            | TBD      |
| runtime/testdata/testwinsignal              | TBD      |
| runtime/trace                               | TBD      |
| slices                                      | TBD      |
| sort                                        | TBD      |
| strconv                                     | TBD      |
| strings                                     | TBD      |
| sync                                        | TBD      |
| sync/atomic                                 | TBD      |
| syscall                                     | TBD      |
| syscall/js                                  | TBD      |
| testing                                     | TBD      |
| testing/fstest                              | TBD      |
| testing/iotest                              | TBD      |
| testing/quick                               | TBD      |
| text/scanner                                | TBD      |
| text/tabwriter                              | TBD      |
| text/template                               | TBD      |
| text/template/parse                         | TBD      |
| time                                        | TBD      |
| time/tzdata                                 | TBD      |
| unicode                                     | TBD      |
| unicode/utf16                               | TBD      |
| unicode/utf8                                | TBD      |
| unsafe                                      | TBD      |



## Tooling (`gno` binary)

| go command        | gno command      | comment                                                               |
|-------------------|------------------|-----------------------------------------------------------------------|
| go bug            |                  | see https://github.com/gnolang/gno/issues/733                         |
| go build          | gno build        | same intention, limited compatibility                                 |
| go clean          | gno clean        | same intention, limited compatibility                                 |
| go doc            | gno doc          | limited compatibility; see https://github.com/gnolang/gno/issues/522  |
| go env            |                  |                                                                       |
| go fix            |                  |                                                                       |
| go fmt            |                  |                                                                       |
| go generate       |                  |                                                                       |
| go get            |                  |                                                                       |
| go help           |                  |                                                                       |
| go install        |                  |                                                                       |
| go list           |                  |                                                                       |
| go mod            |                  |                                                                       |
| + go mod download | gno mod download | same behavior                                                         |
|                   | gno precompile   |                                                                       |
| go work           |                  |                                                                       |
|                   | gno repl         |                                                                       |
| go run            | gno run          |                                                                       |
| go test           | gno test         | limited compatibility                                                 |
| go tool           |                  |                                                                       |
| go version        |                  |                                                                       |
| go vet            |                  |                                                                       |
