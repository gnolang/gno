# PRxxxx: meter the checker walks

## Context

The rule files added along the type-checker ladder walk types and
trees: type-shape and type-expression walks, `typeComplete`,
`tagsIdentical`, `inMapValue`, the unused-name walk, `carryIota`,
literal index checks, duplicate-case keys. Preprocess runs at deploy
under the transaction's gas meter; an unmetered walk is work a deployer
does not pay for and a DoS surface on pathological declarations.

## Decision

- Every checker walk bills the transaction's preprocess meter through
  `chargeCPUGas`, as `embedWalk` already does.
- One constant, `OpCPUSlopeCheckNode`, per node visited; an estimate,
  not yet fitted against measurements.
- A pin test asserts the charge grows per field.
- Filetests run unmetered, so no golden moves.

## Alternatives considered

- Fitting per-walk constants now: premature until the rule set settles;
  one slope keeps the accounting simple to revisit.
- Bounding the walks instead of metering them: a cap rejects valid
  programs, a meter charges them.

## Consequences

- Deploy gas rises slightly with declaration size; the amount is
  proportional to what the checker reads.
