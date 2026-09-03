# VM assigned GRC20 token IDs

## Context

GRC20 callers supplied a local sequence ID. Two tokens created with the same
realm, symbol, and sequence ID produced identical event identifiers even when
their ledgers were independent. Package initialization also could not call
`runtime.NewRealmID` because only `MsgCall` enabled issuance. A fully opaque
token ID would also force indexers to observe `NewToken` before they could map
later `Transfer` or `Approval` events to an origin realm.

## Decision

`runtime.NewRealmID` returns `<realm-path>:<realm-time>` for the current
persistent realm. Valid package paths cannot contain `:`, so consumers may
split on that delimiter and recover both the verified origin realm and numeric
realm time. `grc20.NewToken` stores this value directly as its canonical ID.
The token also keeps its origin realm separately for registry authorization
and display metadata. `NewToken` does not emit a redundant `realm` attribute
because every event already carries the canonical token ID.

Realm ID issuance is enabled while `AddPackage` and `EnablePackage` execute
package initialization. The shared Gno test context enables the same behavior
so realm filetests exercise production token construction.

## Alternatives considered

Keeping caller managed sequence IDs would preserve the old API but would not
prevent collisions. Adding a `realm` attribute to every token event would keep
the ID opaque, but duplicate the same metadata and expand every event. Keeping
that attribute only on `NewToken` would require indexers to retain the complete
creation history. Prefixing GRC20 IDs with the origin realm while retaining an
opaque VM ID would repeat the realm identity in two encodings and add extra
delimiters without improving uniqueness.

## Consequences

Every successful token construction consumes the issuing realm's persistent
time. Token IDs are unique within the realm and make every token event
independently attributable to its verified origin realm. The numeric suffix is
not contiguous across ID calls because object finalization advances the same
realm time. The `<realm-path>:<realm-time>` representation becomes a public
format for `runtime.NewRealmID`. Existing callers must remove the sequence ID
argument; symbol remains token metadata rather than part of the ID.
