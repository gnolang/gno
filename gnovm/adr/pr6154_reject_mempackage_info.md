# ADR: Refuse a submitter-supplied `MemPackage.Info`

## Status

Proposed

## Context

`std.MemPackage.Info` is amino field 5, typed `any`. `UnmarshalAnyBinary2`
decodes it into any concrete type the codec has registered, so a
`MsgAddPackage` or a `MsgRun` arriving off the wire can carry an arbitrary
value there. `std.MemFile` is registered, which is enough on its own: a
`MsgAddPackage` with a 4 KB `MemFile` in `Info` encodes to 4353 bytes where
the same message without it encodes to 221.

Nothing in the tree ever puts a value in the field. `ReadMemPackage`, the
genesis loaders, gnokey, gnoclient, gnodev and the gnovm test harness all
leave it zero, and the three assignments that exist
(`MemPackageFilter.FilterMemPackage`, `splitProdAllButProd`,
`GetMemPackageAll`) copy whatever the struct already holds. Nothing reads it.

A submitted value nevertheless reached merkleized state. `AddPackage`
overwrites `Type` and leaves `Info` alone, `MsgAddPackage.ValidateBasic`,
`std.MemPackage.ValidateBasic` and `ValidateMemPackageAny` never looked at
it, and `AddInertPackage` and `AddMemPackage` amino-marshal the whole struct
into the `inert_pkg:` and `pkg:` blobs.

Those bytes were also priced far below the source they travel with.
`chargePreprocessGas` sums `.gno`, `gnomod.toml` and `gno.mod` bodies only,
so a source byte pays `PreprocessGasPerByte` (1250 by default) on top of the
ante's `TxSizeCostPerByte` (10, levied over the whole encoded tx), while an
`Info` byte pays that 10 plus `amino.GasEncodePerByte` (3) for each blob it
is encoded into, plus the iavl write behind it. Two orders of magnitude per
byte, for bytes that land in the same merkleized blob.

The gap is a discount, not an unbounded hole: `Block.MaxTxBytes` is 1 MB
(`MaxBlockTxBytes`, refused in `CListMempool.CheckTxWithInfo` with
`TxTooLargeError`), so neither a megabyte of `Info` nor a megabyte of `.gno`
fits in one transaction. Gas is not what stops the source: a megabyte of it
would stay under `Block.MaxGas` (3e9), at 1.25e9 gas, if it could get there
at all.

## Decision

`ValidateMemPackageAny` returns an error wrapping `ErrMemPackageInfo` when
`Info` is non-nil.

That function is the one shared gate both submit paths reach: `AddPackage`
calls it directly and `Run` calls it through `ValidateMemPackage`. It is also
what the store's write path (`AddMemPackage`) and the genesis loader
(`LoadPackage`) call, so no route puts an `Info` value into a stored blob.

`EnablePackage` now calls it too, on the blob it reads back out of
`inert_pkg:`, before the type check and before `RunMemPackage`. That is the
one place a payload parked by an older binary is still met, and without the
call it arrives at `AddMemPackage`, which re-validates at the end of
`RunMemPackage` and *panics* rather than returning: `doRecover` turns that
into `VM panic: <error: *errors.errorString>`, an untyped error carrying no
diagnostic at all. With the call, the enable returns `ErrInvalidPackage`
naming the field and writes no live blob.

A parked blob outlives the binary that validated it, so the call is worth
more than the `Info` case: every stored-blob failure `ValidateMemPackageAny`
*returns* now reaches the approver as a typed error. It is not a complete
enable-time validation, and should not be read as one.
`ValidateMemPackageAny` raises some failures as a panic of its own rather
than returning them, `mptype.Validate(mpkg.Path)` on a type/path mismatch
above all, and it runs that before the `Info` check. So a parked `MPUserAll`
blob whose path stops satisfying `IsUserlib` under a later binary still
arrives as the same opaque recovered panic. Turning those into errors is a
change inside the validator, not here.

The refusal carries an identity. `AddPackage`, `Run` and `EnablePackage` map
it through one `errInvalidMemPackage` helper: `errors.Is(err,
gno.ErrMemPackageInfo)` becomes `ErrInvalidPackage`, and every other
validation failure keeps the `ErrInvalidPkgPath` the submit paths have always
returned. Without the sentinel a submitter reads "invalid package path" for a
field that is not a path, and correcting the type later would need a second
fork, since the type URL is hashed into the tx result.

The field itself stays on the type. `getMemPackage` and `GetInertPackage`
read blobs with `amino.MustUnmarshal`, and the generated decoder ends in
`return fmt.Errorf("unknown field number %d for MemPackage", fnum)`, so
removing field 5 would panic on any legacy blob that carries a value, turning
a dormant payload into an unreadable package. The struct doc also still
describes the extension the field was reserved for.

## Alternatives considered

- **Clear it in the keeper**, `memPkg.Info = nil` beside the existing
  `memPkg.Type = gno.MPUserAll`. Rejected: the tx would then succeed while
  storing bytes other than the ones the submitter signed, and it needs the
  same line in `AddPackage` and in `Run`. A refusal says what happened.
- **Reject it in `MsgAddPackage.ValidateBasic` and `MsgRun.ValidateBasic`.**
  Rejected: two copies of one rule, and it leaves the store's write path, the
  genesis loader and the enable path ungated. It would buy one thing the
  shipped guard does not, and only one: `validateBasicTxMsgs` runs in every
  mode while `handler.Process` is skipped for `RunTxModeCheck`, so the
  shipped guard first fires at DeliverTx and a payload still passes the
  mempool into block data. A junk `.gno` file behaves exactly the same way,
  so that is completeness rather than a defect.
- **Reject it in `std.MemPackage.ValidateBasic`.** Same coverage, since
  `ValidateMemPackageAny` is its only caller, but the sibling
  submitter-authored field `Type` is already checked in
  `ValidateMemPackageAny`, and `tm2/pkg/std` holds the wire type rather than
  the chain's admission policy.
- **Widen `PackageContentHash` to cover `Info`.** A separate question: what an
  approval binds to. It would leave the pricing gap open, since a payload an
  approver signs for is still stored at encode-and-store rates.

## Consequences

- A `MsgAddPackage` or `MsgRun` carrying a non-nil `Info` is now refused, and
  a `MsgEnablePackage` over a parked blob that carries one is refused with a
  typed error instead of a recovered panic. On the old binary the same
  `MsgAddPackage` succeeds and writes `inert_pkg:` or `pkg:` blobs containing
  the value, so a validator that has not upgraded computes a different app
  hash for such a block: this is a hardfork change, not a node-local one.
- Existing chain state needs no repair. Read paths are untouched and nothing
  reads `Info`, so a blob that does carry a payload still loads and executes
  exactly as before. `EnablePackage` is the one path that re-validates a
  stored blob, and it now returns an error there rather than panicking.
  Replay is the exception, and it is a fork concern rather than a state one:
  `AddPackage` has no `auth.IsGenesisReplay` exemption (only the inert
  lifecycle messages do), so a fork replaying a source chain that accepted a
  payload now fails that `MsgAddPackage` during InitChain. `deliverGenesisTx`
  logs the delivery error and records it in the replay report, then the
  outcome is the operator's: the default `PanicOnFailingTxResultHandler`
  aborts the boot, and `-skip-failing-genesis-txs`
  (`NoopGenesisTxResultHandler`) continues and the fork comes up without that
  package. No chain has accepted one, so neither outcome exists today.
- Genesis packages are built by `ReadMemPackage` and no `genesis_txs.jsonl`
  in the tree sets `info`, so the genesis loader is unaffected in practice. A
  genesis file that did set it would now fail to load, which is the intended
  outcome: those bytes have no reader.

## AI assistance

Written with AI assistance (Claude Code). The assistance traced every writer
and reader of the field across the tree, drove the fix test-first, and
corrected two claims that did not survive checking the constants: the gas
ceiling is not what bounds a large payload, and the pre-fork enable path
fails as an opaque recovered panic rather than an internal error. The human
author reviewed and owns the change.
