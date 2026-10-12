# gocorpus: Go's test corpus as Gno filetests

`testdata/` holds files lifted verbatim from Go's `$GOROOT/test` (the
release in `VERSION`), each followed by a comment tail the harness writes
(`gnovm/pkg/test/filetest.go`). Nothing above the tail is edited.

## Per-file contract

Evidence blocks (raw, auto-written), then exactly one verdict line, last:

| block | meaning |
|---|---|
| `GnoOutput:` / `GoOutput:` | stdout of Gno and of `go run` (run files) |
| `GnoError:` | Gno's own errors, per line, wording kept |
| `GoTypeCheckError:` / `GoBuildError:` | lines the go/types guard / `go build` flagged; lines only, their wording follows the Go release |
| `GoTypesOnly:` | errorcheck markers only the guard caught: what Gno's checker still lacks |
| `GnoOverStrictError:` / `UncaughtError:` | Gno rejects what gc accepts / gc-invalid code neither checker caught |
| `GnoPreprocessError:` | Gno's errors on compile-only files |

| verdict | meaning |
|---|---|
| `Clean: run\|compile\|errorcheck` | Gno matches Go, nothing to bless |
| `Unsupported: <tag>: <reason>` | skipped, feature gap or gc-only test; tags in `errorcheck.go` (`unsupportedTags`) |
| `KnownDivergence: <category>: <reason>` | accepted difference |
| `KnownIssue: <category>: <reason>` | a Gno bug, pinned so it is greppable |

Verify (`go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus`) runs Gno only
and compares with the pinned Go side; it never shells out to `go`. One
dependency on the Go release remains: the go/types guard runs in-process, so
the *set of lines* in `GoTypeCheckError:` / `GoTypesOnly:` follows the Go that
built the test binary. Run with `GOTOOLCHAIN=$(cat VERSION)` to match the
goldens; `make -C gnovm _test.filetest` does not pin the toolchain unless
`go.mod` carries a `toolchain` line, so set it explicitly.

## Re-sync

    GOTOOLCHAIN=$(cat gnovm/tests/files/gocorpus/VERSION) \
      go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus -short -update-golden-tests

Sync runs `go run` / `go build` for the Go side; the pinned toolchain keeps
`GoOutput` stable. Bump `VERSION` and re-sync together.
