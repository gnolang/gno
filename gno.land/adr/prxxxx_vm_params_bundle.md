# ADR: Bundle VM governance parameters into one KV read

## Status

Proposed.

## Context

Every transaction loads the VM module parameters before applying their governed
store-gas configuration. The parameters were stored only as one JSON-encoded KV
entry per `Params` field, and `ParamsKeeper.GetStruct` therefore performed one
store read per field. Later reads in the same transaction were cache hits, but
the first load paid for every key.

The original investigation measured 13 fields on the former IAVL-backed main
store. The current `master` baseline differs in two important ways:

- `Params` has 14 fields after `PreprocessGasPerByte` was added.
- The main store is B+32 with a fast index. The ante handler loads VM params
  before installing their governed fixed depths, so the first load uses the
  B+ tree's live depth estimate, not `FixedGetReadDepth100`.

For a tree containing `N` entries, the current B+ estimator uses
`d100 = max(100, bits.Len64(N) * 20)`. With `ReadCostFlat = 59,000`, replacing
14 cold field reads with one cold bundle read removes the flat component

`13 * d100 * 59,000 / 100`.

For `DefaultParams`, the 14 JSON field values occupy 162 bytes while the amino
binary struct occupies 144 bytes, removing another `18 * 17 = 306` read gas.
The resulting parameter-load saving is 767,306 gas at the one-read depth floor
and approximately 4,142,106 gas around 100 million entries (`d100 = 540`). It
is no longer the size-independent 7,080,255 gas measured on the older
13-field/IAVL baseline.

There was also a gas-meter lifecycle defect in the old path. `gnoland` loaded
VM params while BaseApp's temporary pre-ante meter was active, then the auth
ante handler replaced that meter with the transaction meter. The storage reads
were visible in a trace but were discarded from the reported `GasUsed`. The
bundle alone therefore produced no end-to-end gas reduction against the
unmodified binary, even though it reduced physical reads.

This change fixes that lifecycle at the same activation point as the bundle.
It is consensus-affecting: transaction gas changes can change out-of-gas
outcomes, and adding the bundle changes the application hash.

## Decision

Store an amino-binary encoding of the complete `vm.Params` value under the
chain-internal root params key `_vm_params`.

The key is deliberately unprefixed. A key such as `vm:_params` would be
addressable through the generic governance parameter writer and would invoke
`VMKeeper.WillSetParam` recursively. The unprefixed key follows the existing
`_realmmeta_` precedent: only chain code can address it, and it is not exposed
as a module parameter.

### Reads

`VMKeeper.GetParams` reads and decodes `_vm_params` when present. The bundle is
authoritative. A malformed present bundle fails loudly instead of falling back
to potentially stale individual values.

When the bundle is absent, `GetParams` falls back to `GetStruct("vm:p")` so an
upgraded binary can decode legacy state. The three specialized chain-domain and
system-package getters try the bundle first and otherwise perform their
original single-leaf read; they do not fan out over the whole legacy struct.

The fallback is value-compatible, not gas-compatible: probing the absent bundle
costs one additional cold read. This is why the binary rollout itself still
requires the coordinated upgrade described below.

### Gas-meter lifecycle

`auth.NewAnteHandler` now installs the bounded transaction meter and its
out-of-gas recovery before invoking the optional `AnteOptions.PrepareGasMeter`
callback. The callback receives that final meter, loads VM params once, and
applies the resulting depth configuration to the returned context. Existing
callers that do not provide the callback retain the previous behavior. Genesis
and replay modes continue to use the meter semantics selected by `SetGasMeter`
(including their existing infinite-meter paths).

This is a coordinated repricing, not merely an optimization. On the measured
`foo20` GRC20 transfer at the current test tree depth, the unmodified master
reported `6,148,884` gas for both layouts because the parameter-load charge was
discarded. With the lifecycle fix, the correctly metered legacy layout reports
`7,803,638` gas (14 reads, `1,654,754` gas) and the bundled layout reports
`6,269,332` gas (one read, `120,448` gas). The bundle therefore saves exactly
`1,534,306` gas against the correctly metered legacy schedule, while appearing
`120,448` gas more expensive than unmodified master because master was
undercharging that one read.

### Writes and synchronization

`SetParams` validates and amino-encodes the struct before changing state, then
writes both representations. The individual keys remain available to the
generic params query and governance paths; the bundle is written last.

Single-field governance changes pass through `VMKeeper.WillSetParam` immediately
before `ParamsKeeper` writes the individual leaf. The hook constructs and
validates the complete candidate value, then rewrites an already-active bundle.
Both writes live in BaseApp's transaction cache and commit or roll back
together. `CommitGnoTransactionStore` is not involved: it flushes Gno object
state, while BaseApp owns params-store commit.

On legacy state, `WillSetParam` does not create the bundle. Bundle creation is a
state-schema migration and must not happen as an incidental side effect of an
otherwise ordinary governance proposal.

### Migration and coordinated activation

`VMKeeper.InitGenesis` calls `SetParams`, so fresh chains and the repository's
supported export/replay hardfork flow create the bundle before replaying any
transactions. This is the explicit migration path in this change.

For a live chain, operators must use the governance halt and minimum-version
mechanism (`node:p:halt_height` and `node:p:halt_min_version`) and activate this
binary only after the coordinated halt. Old and new binaries must not execute
the same post-upgrade block:

- even without a bundle, the new binary charges the additional absent-key
  probe;
- on bundle-active state, the new binary rewrites the bundle when a VM
  parameter changes while the old binary writes only the individual leaf,
  producing different app hashes.

On legacy same-database state, neither a fallback read nor an ordinary
single-parameter governance change activates the bundle. Such activation is
unsupported by this change and requires the deterministic migration described
below.

This change does not add an in-place same-database migration hook because the
current B+ main store is itself supported only for fresh chains and
export/import forks. If a future release needs an in-place rollout, it must add
a deterministic one-time migration at the scheduled upgrade height before any
transaction executes. It must not write the bundle during process startup or
as read-repair in `GetParams`, because those paths are not consensus state
transitions and queries must remain read-only.

Historical replay under `GasReplayMode="source"` bypasses the new metering as
designed. Strict replay uses the new gas schedule and may report gas deltas.

## Alternatives considered

### Cache params in keeper memory

An in-memory cache could remove all persistent reads after startup, but it adds
mutable keeper state, invalidation requirements, and replay/query concurrency
risk. A state-backed bundle retains ordinary transaction-cache semantics.

### Move params out of the depth-estimated store

This attacks the gas multiplier but changes the storage/gas model for more than
VM configuration and requires a broader migration.

### Stop charging depth for configuration reads

Special-casing config keys would make the gas layer aware of module semantics.
Bundling instead reduces the real storage work and the corresponding gas.

### Store only the bundle

Removing individual keys would simplify synchronization, but it would break the
generic params query and governance writers. Keeping both representations
preserves those interfaces.

### Use `vm:_params`

This keeps the key visually within the module namespace but exposes the raw
blob to generic module-parameter writes unless a new internal-write sentinel is
added. An unprefixed internal key is simpler and safer.

## Consequences

- A transaction with an active bundle pays one cold VM-parameter read instead
  of 14; repeated `GetParams` and specialized getter calls hit the per-tx cache.
- Genesis and every active-bundle governance parameter update write one extra
  key. Governance updates remain more expensive than before but are rare.
- Individual parameter queries and governance writes remain compatible.
- The bundle adds a committed key and therefore changes genesis/app hashes.
- Binary amino encoding is deterministic for this flat, map-free struct.
  Appending a field lets an upgraded decoder read older bundles, but an older
  decoder rejects a newer bundle containing the unknown field. Any schema
  extension therefore requires a coordinated binary upgrade, and existing
  fields must not be reordered or repurposed.
- The regression test disables per-byte gas and pins read depth to one operation
  so it directly proves that all current VM parameter call sites collapse to
  one cold read. A separate fallback assertion dynamically tracks the number of
  `Params` fields.
- The VM gas regression performs the same cold GRC20 transfer against active
  bundle and legacy layouts. It asserts one versus 14 parameter reads,
  identical non-parameter gas, equal live B+ depth, and the exact saving
  derived from the 13 removed reads plus the encoding-byte delta. In its
  isolated harness (`d100 = 200`), total gas falls from 6,652,703 to
  5,118,397: a 1,534,306 gas saving for unchanged transfer work. A matching
  gnodev run verifies the final BaseApp `GasUsed` numbers recorded above.

## AI assistance

This change was implemented with AI assistance. The human author is responsible
for reviewing and owning the contribution.
