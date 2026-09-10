# ADR: Bypass the gas meter for the ante-handler VM params read

## Context

The gno.land ante wrapper in `NewAppWithOptions` reads `vmk.GetParams(ctx)` to
build the store `GasConfig`. This runs before the auth ante handler calls
`SetGasMeter`, which installs the per-transaction meter
(`gno.land/pkg/gnoland/app.go`, `tm2/pkg/sdk/auth/ante.go`).

`vmk.GetParams` charges each field read against the current meter. At the ante
site that meter is the passthrough meter `runTx` installs for `DeliverTx`
(`tm2/pkg/sdk/baseapp.go`), whose head limit is the gas remaining in the block.
`SetGasMeter` then replaces it and `runTx` sets `ctx = newCtx`, so the
passthrough meter is discarded. As long as the read fits in the remaining block
gas the charge reaches neither the fee payer's `GasWanted` nor the block gas
meter. When it does *not* fit, the read out-of-gases the passthrough and the
charge does reach the block gas meter — as the whole remaining block gas. That
case is the user-visible half of the defect; see Consequences.

Two facts bound the impact.

First, no transaction ever paid for this read, before or after the change. The
same params are read again in the begin-tx hook, after `ctx = newCtx`, under the
real per-transaction meter (`SetBeginTxHook` to `newGnoTransactionStore`) — but
that read is a **cache hit and costs nothing**. `runTx` cache-wraps the store
once for the whole transaction (ante and messages share one layer), and
`cacheStore.Get` populates the cache on both the metered and the nil-`GasContext`
branch, so the ante-site read always warms the keys the hook then re-reads;
`cacheStore` charges no gas on a cache hit (`tm2/pkg/store/cache/store.go`). The
same reasoning is already recorded at the `checkCodePolicy` call site, which
notes that its re-read of these keys would be "cache hits and cost no gas".
Measured: a metered `GetParams` on a cache already warmed by a nil-meter
`GetParams` charges 0 gas.

Second, the backing-store read for the `vm:p:` keys happens once per block,
because the block-scoped cache absorbs later transactions.

So the ante-site charge was work that escaped accounting, its loss changes no
user-visible gas, and there is no per-transaction I/O amplification.

Note what this means for the PR #5629 policy that `vm.GetParams` reads carry a
real gas signal: on the transaction path that signal was already absent, because
the ante-site charge was discarded and the begin-tx hook read is a cache hit.
`VMKeeper.GetParams` is genuinely metered only on the query path (the
`newGnoTransactionStore` throwaway stores in `gno.land/pkg/sdk/vm/keeper.go`,
each preceded by its own query gas meter) and at genesis. This change does not
remove a signal from the transaction path; it removes a charge that was thrown
away.

The other config reads in the same closure already avoid the meter:
`acck.GetParams` reads with `ctx.WithGasMeter(nil)` and `gpk.LastGasPrice`
passes a nil `GasContext`. (`bank.GetParams` uses the same bypass but is not
called from this closure at all.) `vmk.GetParams` was the only metering read on
the non-genesis path. The height-0 branch further down the same closure —
`acck.GetAccount`, `acck.SetAccount`, `bankk.MintCoins` — still meters and still
has its charge discarded; that is harmless today because `InitChain` installs an
infinite block gas meter and `SetGasMeter` returns an infinite meter at height
0, and it is left alone here.

## Decision

Bypass the meter at the ante-site call only.

```go
vmParams := vmk.GetParams(ctx.WithGasMeter(nil))
vmParams.ApplyToGasConfig(&gasCfg)
```

`VMKeeper.GetParams` itself is unchanged and stays metered for its other
callers, which is what preserves the #5629 gas signal where that signal actually
exists — the query path. The change aligns the ante-phase config reads on one
bypass. The stale comment claiming the read "DOES meter" for a real gas signal
is corrected, since that charge was always discarded before `SetGasMeter`, and
so is the later claim that the ctx is "on the infinite meter until
auth.SetGasMeter runs" — on `DeliverTx` it is on the passthrough meter, which is
precisely why the read can out-of-gas.

The ante site is the only caller that runs before a real meter, so bypassing it
closes the leak.

## Alternatives considered

- **Bypass inside `VMKeeper.GetParams` itself** (proposed fix). On the
  transaction path this is behaviourally identical to the narrow fix, since no
  transaction pays for the read either way. It differs on the query path, where
  the read *is* charged, and it breaks
  `gno.land/pkg/sdk/vm/gas_test.go`. Rejected as broader than the defect: the
  narrow fix leaves the query-path charge, and the queries are where the #5629
  signal is real.
- **Transfer the discarded charge to the new meter.** Requires changing the
  baseapp meter-handoff, which is invasive and consensus-sensitive. It would
  also be a genuine gas increase for every transaction rather than a no-op,
  since the begin-tx hook read is a cache hit and nothing currently pays this
  cost. Rejected.

## Consequences

- No change to user-visible gas for transactions with adequate block headroom
  (verified: identical `GasUsed` before and after). The ante-site charge was
  already discarded, and the begin-tx hook read was already a free cache hit.
- State-machine-breaking at the block-gas boundary, so this needs a coordinated
  upgrade. When a block's remaining gas is below the ante-site read cost, the
  pre-change passthrough meter OOGs during the read and `runTx` charges the
  whole remaining block gas to the block meter; after the change the transaction
  succeeds. That changes both the `ResponseDeliverTx` and the block-gas input to
  `UpdateGasPrice`. Because `DeductFees` runs after `SetGasMeter`, the
  pre-change failure also burned the block tail without collecting a fee.
- That read cost is **not a constant**. `store.DefaultGasConfig` leaves
  `Min*Depth100` and `Fixed*Depth100` at 0, and the ante-site read runs before
  the governed config is applied, so `cacheStore` prices it from the store's own
  depth estimate. The params keeper is mounted on `mainKey`, the
  `storebptree.FastStoreConstructor` store shared with all VM state, whose
  `expectedDepth100` is `max(100, bits.Len64(size)*20)`. Measured cost of one
  ante-site `GetParams`:

  | keys in `mainKey` | gas |
  |---|---|
  | 0 (fresh chain, and what a flat store measures) | 1_183_978 |
  | 100 | 1_655_978 |
  | 10_000 | 3_307_978 |
  | 500_000 | 4_487_978 |

  So the dead zone at the tail of every block is a few million gas on a chain of
  any real size, not the ~1.2M a fresh-chain measurement suggests.
  `measureVMParamsReadGas` in the regression test measures against a flat
  `dbadapter` store, which is why it reports the first row; that is a lower
  bound, which is the safe direction for the test's budget.
- `WithGasMeter(nil)` only suppresses charging; the returned value is identical.
  This is already accepted for the auth config read in this closure.

## Verification

- `gno.land/pkg/gnoland/app_gas_test.go` adds two ante-level regression tests that
  fail before the change and pass after. One drives a `MaxGas=1` block, the other
  a funded transaction whose block headroom covers its own cost but not the leaked
  ante read. Deleting the ante-site read outright (rather than un-metering it)
  fails the second test, so the pair also covers the governed-params plumbing.
- `gno.land/pkg/sdk/vm/params_test.go` `TestGetParamsReadsBackingStoreOncePerBlock`
  documents the once-per-block caching invariant. It is not a fix guard: it passes
  on unfixed code by design.
- The gas triad from `AGENTS.md`:
  - `go test ./gno.land/pkg/sdk/vm/ -run Gas` passes.
  - `go test ./gno.land/pkg/integration/ -run TestTestdata` passes.
  - `go test ./gnovm/pkg/gnolang/ -run Files -test.short` — no gas or allocation
    filetest is affected (this change touches no `gnovm/` code).
