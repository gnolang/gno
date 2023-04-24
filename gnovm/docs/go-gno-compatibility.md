# Go<>Gno compatibility

**WIP: does not reflect the current state yet.**

## Native keywords

| keyword     | status |
|-------------|--------|
| break       |        |
| case        |        |
| chan        |        |
| const       |        |
| continue    |        |
| default     |        |
| defer       |        |
| else        |        |
| fallthrough |        |
| for         |        |
| func        |        |
| go          |        |
| goto        |        |
| if          |        |
| import      |        |
| interface   |        |
| map         |        |
| package     |        |
| range       |        |
| return      |        |
| select      |        |
| struct      |        |
| switch      |        |
| type        |        |
| var         |        |

## Stdlibs

<!-- generated with: find . -name "*.go" | grep -v _test.go | grep -v internal/ | grep -v vendor/ | xargs dirname | sort | uniq -->

| package                                     | status   |
|---------------------------------------------|----------|
| archive/tar                                 |          |
| archive/zip                                 |          |
| arena                                       |          |
| bufio                                       |          |
| builtin                                     |          |
| bytes                                       |          |
| cmd/addr2line                               |          |
| cmd/api                                     |          |
| cmd/api/testdata/src/issue21181/dep         |          |
| cmd/api/testdata/src/issue21181/indirect    |          |
| cmd/api/testdata/src/issue21181/p           |          |
| cmd/api/testdata/src/pkg/p1                 |          |
| cmd/api/testdata/src/pkg/p2                 |          |
| cmd/api/testdata/src/pkg/p3                 |          |
| cmd/api/testdata/src/pkg/p4                 |          |
| cmd/asm                                     |          |
| cmd/buildid                                 |          |
| cmd/cgo                                     |          |
| cmd/compile                                 |          |
| cmd/covdata                                 |          |
| cmd/covdata/testdata                        |          |
| cmd/cover                                   |          |
| cmd/cover/testdata                          |          |
| cmd/cover/testdata/html                     |          |
| cmd/cover/testdata/pkgcfg/a                 |          |
| cmd/cover/testdata/pkgcfg/b                 |          |
| cmd/cover/testdata/pkgcfg/main              |          |
| cmd/dist                                    |          |
| cmd/distpack                                |          |
| cmd/doc                                     |          |
| cmd/doc/testdata                            |          |
| cmd/doc/testdata/merge                      |          |
| cmd/doc/testdata/nested                     |          |
| cmd/doc/testdata/nested/empty               |          |
| cmd/doc/testdata/nested/nested              |          |
| cmd/fix                                     |          |
| cmd/go                                      |          |
| cmd/gofmt                                   |          |
| cmd/go/testdata                             |          |
| cmd/link                                    |          |
| cmd/link/testdata/pe-binutils               |          |
| cmd/link/testdata/pe-llvm                   |          |
| cmd/link/testdata/testBuildFortvOS          |          |
| cmd/link/testdata/testHashedSyms            |          |
| cmd/link/testdata/testIndexMismatch         |          |
| cmd/link/testdata/testRO                    |          |
| cmd/nm                                      |          |
| cmd/objdump                                 |          |
| cmd/objdump/testdata                        |          |
| cmd/objdump/testdata/testfilenum            |          |
| cmd/pack                                    |          |
| cmd/pprof                                   |          |
| cmd/pprof/testdata                          |          |
| cmd/test2json                               |          |
| cmd/trace                                   |          |
| cmd/vet                                     |          |
| cmd/vet/testdata/asm                        |          |
| cmd/vet/testdata/assign                     |          |
| cmd/vet/testdata/atomic                     |          |
| cmd/vet/testdata/bool                       |          |
| cmd/vet/testdata/buildtag                   |          |
| cmd/vet/testdata/cgo                        |          |
| cmd/vet/testdata/composite                  |          |
| cmd/vet/testdata/copylock                   |          |
| cmd/vet/testdata/deadcode                   |          |
| cmd/vet/testdata/directive                  |          |
| cmd/vet/testdata/httpresponse               |          |
| cmd/vet/testdata/lostcancel                 |          |
| cmd/vet/testdata/method                     |          |
| cmd/vet/testdata/nilfunc                    |          |
| cmd/vet/testdata/print                      |          |
| cmd/vet/testdata/rangeloop                  |          |
| cmd/vet/testdata/shift                      |          |
| cmd/vet/testdata/structtag                  |          |
| cmd/vet/testdata/tagtest                    |          |
| cmd/vet/testdata/testingpkg                 |          |
| cmd/vet/testdata/unmarshal                  |          |
| cmd/vet/testdata/unsafeptr                  |          |
| cmd/vet/testdata/unused                     |          |
| compress/bzip2                              |          |
| compress/flate                              |          |
| compress/gzip                               |          |
| compress/lzw                                |          |
| compress/zlib                               |          |
| container/heap                              |          |
| container/list                              |          |
| container/ring                              |          |
| context                                     |          |
| crypto                                      |          |
| crypto/aes                                  |          |
| crypto/boring                               |          |
| crypto/cipher                               |          |
| crypto/des                                  |          |
| crypto/dsa                                  |          |
| crypto/ecdh                                 |          |
| crypto/ecdsa                                |          |
| crypto/ed25519                              |          |
| crypto/elliptic                             |          |
| crypto/hmac                                 |          |
| crypto/md5                                  |          |
| crypto/rand                                 |          |
| crypto/rc4                                  |          |
| crypto/rsa                                  |          |
| crypto/sha1                                 |          |
| crypto/sha256                               |          |
| crypto/sha512                               |          |
| crypto/subtle                               |          |
| crypto/tls                                  |          |
| crypto/tls/fipsonly                         |          |
| crypto/x509                                 |          |
| crypto/x509/pkix                            |          |
| database/sql                                |          |
| database/sql/driver                         |          |
| debug/buildinfo                             |          |
| debug/dwarf                                 |          |
| debug/elf                                   |          |
| debug/gosym                                 |          |
| debug/gosym/testdata                        |          |
| debug/macho                                 |          |
| debug/pe                                    |          |
| debug/plan9obj                              |          |
| embed                                       |          |
| encoding                                    |          |
| encoding/ascii85                            |          |
| encoding/asn1                               |          |
| encoding/base32                             |          |
| encoding/base64                             |          |
| encoding/binary                             |          |
| encoding/csv                                |          |
| encoding/gob                                |          |
| encoding/hex                                |          |
| encoding/json                               |          |
| encoding/pem                                |          |
| encoding/xml                                |          |
| errors                                      |          |
| expvar                                      |          |
| flag                                        |          |
| fmt                                         |          |
| go/ast                                      |          |
| go/build                                    |          |
| go/build/constraint                         |          |
| go/build/testdata/alltags                   |          |
| go/build/testdata/cgo_disabled              |          |
| go/build/testdata/directives                |          |
| go/build/testdata/doc                       |          |
| go/build/testdata/multi                     |          |
| go/build/testdata/non_source_tags           |          |
| go/build/testdata/other                     |          |
| go/build/testdata/other/file                |          |
| go/constant                                 |          |
| go/doc                                      |          |
| go/doc/comment                              |          |
| go/doc/testdata                             |          |
| go/doc/testdata/examples                    |          |
| go/doc/testdata/pkgdoc                      |          |
| go/format                                   |          |
| go/importer                                 |          |
| go/parser                                   |          |
| go/parser/testdata/goversion                |          |
| go/parser/testdata/issue42951               |          |
| go/parser/testdata/issue42951/not_a_file.go |          |
| go/printer                                  |          |
| go/printer/testdata                         |          |
| go/scanner                                  |          |
| go/token                                    |          |
| go/types                                    |          |
| go/types/testdata                           |          |
| go/types/testdata/local                     |          |
| hash                                        |          |
| hash/adler32                                |          |
| hash/crc32                                  |          |
| hash/crc64                                  |          |
| hash/fnv                                    |          |
| hash/maphash                                |          |
| html                                        |          |
| html/template                               |          |
| image                                       |          |
| image/color                                 |          |
| image/color/palette                         |          |
| image/draw                                  |          |
| image/gif                                   |          |
| image/jpeg                                  |          |
| image/png                                   |          |
| index/suffixarray                           |          |
| io                                          |          |
| io/fs                                       |          |
| io/ioutil                                   |          |
| log                                         |          |
| log/internal                                |          |
| log/slog                                    |          |
| log/slog/internal                           |          |
| log/syslog                                  |          |
| maps                                        |          |
| math                                        |          |
| math/big                                    |          |
| math/bits                                   |          |
| math/cmplx                                  |          |
| math/rand                                   |          |
| mime                                        |          |
| mime/multipart                              |          |
| mime/quotedprintable                        |          |
| net                                         |          |
| net/http                                    |          |
| net/http/cgi                                |          |
| net/http/cookiejar                          |          |
| net/http/fcgi                               |          |
| net/http/httptest                           |          |
| net/http/httptrace                          |          |
| net/http/httputil                           |          |
| net/http/internal                           |          |
| net/http/pprof                              |          |
| net/mail                                    |          |
| net/netip                                   |          |
| net/rpc                                     |          |
| net/rpc/jsonrpc                             |          |
| net/smtp                                    |          |
| net/textproto                               |          |
| net/url                                     |          |
| os                                          |          |
| os/exec                                     |          |
| os/signal                                   |          |
| os/user                                     |          |
| path                                        |          |
| path/filepath                               |          |
| plugin                                      |          |
| reflect                                     |          |
| regexp                                      |          |
| regexp/syntax                               |          |
| runtime                                     |          |
| runtime/asan                                |          |
| runtime/cgo                                 |          |
| runtime/coverage                            |          |
| runtime/coverage/testdata                   |          |
| runtime/coverage/testdata/issue56006        |          |
| runtime/debug                               |          |
| runtime/metrics                             |          |
| runtime/msan                                |          |
| runtime/pprof                               |          |
| runtime/pprof/testdata/mappingtest          |          |
| runtime/race                                |          |
| runtime/race/testdata                       |          |
| runtime/testdata/testexithooks              |          |
| runtime/testdata/testfaketime               |          |
| runtime/testdata/testprog                   |          |
| runtime/testdata/testprogcgo                |          |
| runtime/testdata/testprogcgo/windows        |          |
| runtime/testdata/testprognet                |          |
| runtime/testdata/testwinlib                 |          |
| runtime/testdata/testwinlibsignal           |          |
| runtime/testdata/testwinlibthrow            |          |
| runtime/testdata/testwinsignal              |          |
| runtime/trace                               |          |
| slices                                      |          |
| sort                                        |          |
| strconv                                     |          |
| strings                                     |          |
| sync                                        |          |
| sync/atomic                                 |          |
| syscall                                     |          |
| syscall/js                                  |          |
| testing                                     |          |
| testing/fstest                              |          |
| testing/iotest                              |          |
| testing/quick                               |          |
| text/scanner                                |          |
| text/tabwriter                              |          |
| text/template                               |          |
| text/template/parse                         |          |
| time                                        |          |
| time/tzdata                                 |          |
| unicode                                     |          |
| unicode/utf16                               |          |
| unicode/utf8                                |          |
| unsafe                                      |          |



## Tooling (`gno` binary)

| go command          | gno command        | comment                               |
|---------------------|--------------------|---------------------------------------|
| `go bug`            |                    |                                       |
| `go build`          | `gno build`        | same intention, limited compatibility |
| `go clean`          |                    |                                       |
| `go doc`            |                    |                                       |
| `go env`            |                    |                                       |
| `go fix`            |                    |                                       |
| `go fmt`            |                    |                                       |
| `go generate`       |                    |                                       |
| `go get`            |                    |                                       |
| `go help`           |                    |                                       |
| `go install`        |                    |                                       |
| `go list`           |                    |                                       |
| `go mod`            |                    |                                       |
| + `go mod download` | `gno mod download` | same behavior                         |
|                     | `gno precompile`   |                                       |
| `go work`           |                    |                                       |
|                     | `gno repl`         |                                       |
| `go run`            | `gno run`          |                                       |
| `go test`           | `gno test`         | limited compatibility                 |
| `go tool`           |                    |                                       |
| `go version`        |                    |                                       |
| `go vet`            |                    |                                       |
