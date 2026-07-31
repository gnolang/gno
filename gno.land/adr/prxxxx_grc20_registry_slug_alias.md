# ADR: separate GRC20 registry aliases from token identity

## Context

Issues #5988 and #6026 identified that keying the GRC20 registry by
`realm.symbol` prevents a realm from registering distinct tokens that share a
symbol. `Token.ID()` already identifies tokens as `realm.symbol.id`, while
registry callers also supply a slug.

## Decision

The registry requires a valid slug and stores each entry under `realm.slug`.
That key is unique within the registry and cannot overwrite an existing entry.
`Token.ID()` remains the token identity and must still prove that the token
originates from the registering realm and symbol.

The registering realm is responsible for unique Token ID allocation and for
whether one token is registered under multiple aliases. The registry adds no
reverse Token ID index or global ID generator.

Register events retain `token_path`, `pkgpath`, `slug`, and `symbol`.
`token_path` now holds `realm.slug`, and the new `token_id` field holds
`realm.symbol.id`.

## Alternatives considered

- Keep `realm.symbol`: rejected because same-symbol tokens would still collide.
- Key by `Token.ID()`: rejected because the caller-provided lookup alias would
  remain unusable.
- Add a reverse ID index or global ID generator: rejected because ID allocation
  and alias reuse belong to the registering realm.

## Consequences

Registry consumers use `realm.slug`; token-event consumers continue to use
`Token.ID()`. One token may have multiple aliases, and duplicate Token IDs are
not rejected by the registry. Testnet state will be reset, so no stored-state
migration or old-key compatibility layer is added.
