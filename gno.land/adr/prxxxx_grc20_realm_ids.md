# VM assigned GRC20 token IDs

## Context

GRC20 callers supplied a local sequence ID. Two tokens created with the same
realm, symbol, and sequence ID produced identical event identifiers even when
their ledgers were independent. Package initialization also could not call
`runtime.NewRealmID` because only `MsgCall` enabled issuance.

## Decision

`grc20.NewToken` obtains its canonical ID directly from
`runtime.NewRealmID`. The returned string remains opaque and is stored without
prefixing or parsing. The token keeps its origin realm separately for registry
authorization and display metadata.

Realm ID issuance is enabled while `AddPackage` and `EnablePackage` execute
package initialization. The shared Gno test context enables the same behavior
so realm filetests exercise production token construction.

## Alternatives considered

Keeping caller managed sequence IDs would preserve the old API but would not
prevent collisions. Combining the realm path and symbol with the VM ID would
duplicate metadata and create another delimiter format for consumers to parse.
Parsing the VM ID to recover provenance would turn an opaque value into a
public encoding contract.

## Consequences

Every successful token construction consumes the issuing realm's persistent
counter. Token IDs are unique within the realm and carry an unforgeable VM
assigned package identifier. Existing callers must remove the sequence ID
argument. Consumers that parsed the old dotted ID must treat the new ID as an
opaque key and use token metadata for the origin realm and symbol.
