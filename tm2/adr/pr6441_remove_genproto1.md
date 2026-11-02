# Remove genproto (genproto1) and dead code around it

## Status

Proposed, in PR #6441.

## Context

Amino had two protobuf generators. genproto (genproto1,
`tm2/pkg/amino/genproto`, driven by `misc/genproto`) wrote `.proto` schemas
and "pbbindings": `ToPBMessage`/`FromPBMessage` methods that convert a Go
value to a protoc-generated struct, so `amino.Marshal` could encode through
`proto.Marshal`. genproto2 (ADR-001) replaced the bindings with direct
encoders (`pb3_gen.go`) and became the generator gno.land uses; it still
wrote `.proto` schemas by calling genproto1's `WriteProto3Schema`.

genproto1's bindings had been dead for some time:

- `pbbindings.go` and the protoc `*.pb.go` it needs are gitignored. The CI
  `genproto` job generated them and then only diffed the tracked tree, which
  ignores them, so no build or test ever compiled them.
- `amino.RegisterPbbindingsType` is only called from the `init()` of a
  generated `pbbindings.go`, so the codec's pbbindings path (`usePBBindings`,
  `HasNativePbbindings`) never ran.
- In the property fuzz (`TestCodec*`), the pbbindings checks ended in
  `if !ok { continue }` for a type without bindings, which was every type,
  and so also skipped the genproto2 checks after them.
- `HARDENING.md` and the `misc/genproto` banner kept genproto1 "for future
  projects that need protobuf3-compatible generated code". None has appeared
  since genproto2 landed.

`misc/genproto` had one live job: regenerating
`tm2/pkg/bft/privval/upstream/upstreampb/upstream.pb.go`, which is protoc
output used in production for tmkms compatibility, not amino output.

## Decision

- Delete the bindings generator (`bindings.go`, `camelcase.go`), the protoc
  helpers (`MakeProtoFolder`, `RunProtoc`), the example driver and its stale
  `.proto` files, `HARDENING.md`, `misc/genproto`,
  `tm2/pkg/amino/tests/proto3` and `tm2/pkg/amino/tests/cmd`.
- Move the schema writer (`P3Context`, the `P3*` types, `WriteProto3Schema`
  and its helpers) into `tm2/pkg/amino/genproto2`, with its tests and the
  `example/submodule{,2}` fixtures, and trim `stringutil` to
  `ToLowerSnakeCase`. Regenerating with `make -C misc/genproto2` changes no
  `.proto` or `pb3_gen.go` file.
- Remove the codec's pbbindings path: `PBMessager`, `RegisterPbbindingsType`,
  `HasNativePbbindings`, `WithPBBindings`, `MarshalPBBindings`, the
  pbbindings stats counters, and `Object` with the `GetTypeURL` fast path
  that only generated bindings implemented.
- Move the `upstream.pb.go` recipe (protoc, strip version lines, goimports)
  into a Makefile next to `upstream.proto`. The CI `genproto` job becomes an
  `upstreampb` job that runs it and fails on a diff; the output is
  byte-identical. Drop the `protoc-gen-go-grpc` dev dependency, which only
  the deleted `compat.proto` step used.
- Remove `misc/genproto2/genproto2`, a 47 MB macOS build committed by
  mistake in #5282, and ignore it.
- Remove other dead code found by `deadcode -test` over a `go.work` that
  joins the root module with every `contribs/*` and `misc/*` module (so a
  caller in another module counts), cross-checked with staticcheck U1000,
  build-tag, cgo and GOOS variants, and repository-wide greps. Only code
  that is internal, test-only, app-internal or broken goes:
  - `gnovm/pkg/benchops/gno/**`, `.gno` fixtures orphaned when
    `cmd/benchops` was deleted (#5291);
  - `gnovm/cmd/benchstore/{amino_test,values}.go`, behind a `genproto2`
    build tag nothing passes, and no longer compiling under it;
  - unused test tables and mocks in `gnovm/pkg/gnomod`, `tm2/pkg/p2p`,
    `gno.land/pkg/gnoland` and `gno.land/pkg/sdk/vm`;
  - the unused `verify` filter methods and `TestImportError` in
    `gnovm/pkg/test`;
  - `tm2/pkg/store/firstlast.go`, `DiffStores` and `InclusiveEndBytes`,
    which `gno.land/adr/gas_refactor.md` already lists as zero-caller code
    that iterates without a gas context;
  - unused `contribs/gnodev` node methods and its zap logger file;
  - `softfloat.Ftrunc64`/`Ftrunc32` and their helpers, hand-written, so a
    regeneration of the runtime copy does not restore them.
- Remove exported library code that nothing in any module of this
  repository calls, checked the same way, plus amino registration (no
  registered type and no method a registered type needs is removed):
  - whole packages `tm2/pkg/store/prefix` (unused since #276),
    `tm2/pkg/crypto/xchacha20poly1305` (never imported) and
    `tm2/pkg/amino/gengo` (an experimental generator nothing ran);
  - gnolang AST builders (`helpers.go`), `KindOf`, `ParseDeclaredTypeID`,
    the `Span`/`Location` parsers, `WriteToMemPackage`/`WriteToMemFile`,
    `NewValuePath*`, `NewBigdecFrom*` and `GenGnoMod*`;
  - iavl debug printers (`WriteDotGraphv2`, `PrintTree`), which drops the
    `github.com/emicklei/dot` dependency from every module;
  - rootmulti `MultiStoreProof`, bft evidence and state error types that are
    never constructed, `tm2/pkg/bft/node/id.go`, `tm2/pkg/os` process and
    path helpers, gnomod `SubPkg`, the rpc server's typed `GetParam*`
    getters and websocket options, low-level amino varint helpers, std
    `ParseTxs`, auth `GetSignBytes`, bank `NewMsgMultiSend`, and unused
    `tm2/pkg/colors` and `tm2/pkg/random` functions.

## Alternatives considered

- Keep the schema writer in a slimmed `tm2/pkg/amino/genproto`: less churn
  and no import change for genproto2, but it keeps a package named after
  the generator being removed, and genproto2 would still depend on it.
- Keep genproto1 deprecated in-tree, as `HARDENING.md` proposed: it costs a
  CI job, a protoc install in CI, and code that every amino change has to
  keep compiling, for a use nobody has.
- Keep exported library code that has no caller in this repository, for
  projects outside it that may import tm2 and gnovm packages: rejected,
  since nothing in the repository tests or maintains that code, and an
  importer can vendor what it needs from git history.
- Also remove the unexported dead code kept on purpose
  (`codaGotoLoopDefines`, marked `//nolint:unused` "in case needed") and
  `gnovm/pkg/benchops`'s exporter, whose removal edits the VM's
  `machine.go`: left for their owners.

## Consequences

- Breaking for importers of `tm2/pkg/amino`: the API listed above is gone,
  and `tm2/pkg/amino/genproto` no longer exists; its schema API is in
  `genproto2`. Code generated by genproto1 outside this repository no longer
  compiles. Nothing in this repository used any of it.
- Breaking for importers of the tm2, gnovm and gno.land packages that lose
  exported identifiers above, and of `tm2/pkg/store/prefix`,
  `tm2/pkg/crypto/xchacha20poly1305` and `tm2/pkg/amino/gengo`.
- No encoding changes: the pbbindings path never ran, and the schema writer
  produces the same `.proto` files.
- The property fuzz now checks genproto2 byte parity and roundtrip for every
  type, as #6399 also arranged; `go test ./tm2/pkg/amino` takes about 40 s
  longer.
- `make -C tm2/pkg/bft/privval/upstream/upstreampb` regenerates
  `upstream.pb.go`; it needs protoc and `protoc-gen-go`
  (`make -C misc/devdeps install`).
