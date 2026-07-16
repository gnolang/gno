# Preserve pre-message governance-parameter gas charges

## Context

`BaseApp.runTx` installs a temporary passthrough meter before the ante handler
so that block-gas accounting is still protected while a transaction is being
decoded. The auth ante handler then installs the transaction's gas meter with
`GasWanted`. Gnoland's custom ante wrapper used to load VM governance
parameters before entering the auth ante handler. The parameter read therefore
charged the temporary meter, and replacing that meter discarded the charge from
the transaction's reported `GasUsed` (although store-gas traces still showed
the physical reads).

This made reported transaction gas depend on meter lifetime rather than on the
work performed. It also meant that the governed store-gas configuration was
applied only after the read that supplied it.

## Decision

Add an optional `auth.AnteOptions.PrepareGasMeter` callback. The auth ante
handler invokes it after the gas-wanted and mempool-fee checks, after installing
the final transaction meter, and before auth reads or writes. The existing
ante-handler recovery remains around the callback so an out-of-gas parameter
load reports the final meter's consumption.

Gnoland uses the callback to load VM parameters on the transaction meter and
then applies those parameters to the context's store gas configuration. The
custom wrapper no longer performs a pre-ante VM parameter read.

## Alternatives considered

1. Keep the pre-ante read and copy its consumed amount into the transaction
   meter. This would require a second meter protocol, risks double charging,
   and would still make callback failures difficult to report consistently.
2. Change `BaseApp` to share one meter between pre-ante and transaction phases.
   That would couple BaseApp's block-gas safety mechanism to every ante handler
   and change the SDK contract for applications that intentionally replace a
   meter.
3. Keep the read unmetered. This would hide consensus-configuration storage
   work from transaction gas and allow parameter changes to change execution
   cost without a corresponding charge.

## Consequences

- VM parameter loading is charged exactly once to the final transaction meter,
  and out-of-gas errors retain the consumed amount.
- On the current legacy VM-parameter layout, the callback performs the same
  fourteen field reads that the old wrapper performed, so integration gas
  fixtures increase by approximately 1,654,754 gas per VM transaction. The
  separate parameter-bundle change can reduce this physical work later; this
  PR fixes meter ownership independently of that optimization.
- Gas-wanted fixtures and exact gas assertions must include the newly reported
  work. They were recalibrated with the repository's
  `gno.land/pkg/integration/update_gas_wanted.sh` workflow and verified by the
  full integration testdata suite.
- The callback is optional and existing SDK applications that do not configure
  it retain their current ante behavior.
