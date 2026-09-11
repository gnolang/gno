# PR5478: Fix duplicate validator entry when delete and re-add in same block

## Context

When a validator is removed (voting power set to 0) and re-added (voting power > 0)
in the same block, the node crashes with:

```
Error changing validator set: duplicate entry Validator{...} in [...]
```

The root cause is that the validator realm (`r/sys/validators/v2`) has no
`UpdateValidator()` — power changes require a remove + re-add. The `poc.gno`
proposal callback calls `removeValidator()` then `addValidator()` for the same
address, and `saveChange()` blindly appends both entries. When these reach
`processChanges()` in tm2, it rejects them as duplicates.

This is critical because `processChanges()` is strict for a reason:
[`applyUpdates()`](https://github.com/gnolang/gno/blob/master/tm2/pkg/bft/types/validator_set.go#L437)
and [`applyRemovals()`](https://github.com/gnolang/gno/blob/master/tm2/pkg/bft/types/validator_set.go#L490)
assume no duplicate addresses across both lists. If the same address ends up in
both, `applyUpdates` adds the validator, then `applyRemovals` removes it — the
wrong outcome. Therefore tm2's strict rejection is correct and should not be
weakened.

## Decision

Fix at the realm layer — deduplicate and reject duplicate addresses before they
reach tm2.

### 1. `saveChange()` dedup (handles cross-tx duplicates)

`saveChange()` in `validators.gno` now checks if a change for the same address
already exists in the current block. If so, it overwrites (last-writer-wins)
instead of appending. Since all txs in a block run consecutively, tx2 sees tx1's
writes — this handles the cross-tx case where two separate proposals modify the
same validator in the same block.

### 2. `NewPropRequest()` guards (handles within-proposal duplicates)

Reject duplicate addresses in `NewPropRequest()` in
`examples/gno.land/r/sys/validators/v2/poc.gno`:

- **At proposal creation**: panic if `changesFn()` returns duplicate addresses.
- **At proposal execution**: same check in the callback, before applying changes.

### 3. tm2 `processChanges()` — unchanged

The tm2 `processChanges()` duplicate rejection stays as-is — it's a correct
safety net that catches programming errors.

## Alternatives considered

1. **Fix at tm2 level (dedup in `UpdateWithABCIValidatorUpdates`)**: masks the
   real problem — the realm is producing bad input. Downstream code
   (`applyUpdates`/`applyRemovals`) assumes no duplicates, so silently
   deduplicating hides a correctness issue. The tm2 layer's strict check is
   correct and should remain strict.

2. **Defense-in-depth (realm + tm2)**: adds unnecessary complexity to tm2 when
   the root cause is squarely at the realm level. The existing `processChanges()`
   rejection already serves as the tm2 safety net.

3. **Add `UpdateValidator()` to the realm**: cleanest long-term fix so power
   changes emit a single entry. Left as a follow-up.

## Key files

- `examples/gno.land/r/sys/validators/v2/poc.gno` — duplicate address guards
- `examples/gno.land/r/sys/validators/v2/validators_test.gno` — unit tests
- `gno.land/pkg/integration/testdata/validator_duplicate_address.txtar` — integration test

## Consequences

- Proposals with duplicate validator addresses are rejected at creation and execution time.
- `processChanges()` is untouched — no behavior change for tm2.
- Future work: add `UpdateValidator()` to the realm so power changes don't
  require remove + re-add.
