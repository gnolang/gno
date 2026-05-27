# ADR: Client-Only CLA Signing Helper in gnokey

## Context

When CLA enforcement is enabled and a user tries to deploy a package without
signing, they get a terse error:

```
address g1xxx has not signed the required CLA
```

No guidance on what a CLA is, where to find the document, or how to sign it.
The goal is to provide actionable feedback with a copy-pasteable signing command.

## Decision

All CLA helper logic lives in the gnokey client (`gno.land/pkg/keyscli`).
The keeper and VM handler remain unchanged.

### How it works

1. `gnokey maketx addpkg` broadcasts a transaction.
2. If the transaction fails, gnokey checks whether the error message contains
   `"has not signed the required CLA"`.
3. If it does, gnokey makes follow-up queries to the chain:
   - `params/vm:p:syscla_pkgpath` to get the CLA realm path
   - `vm/qeval` with `<realm>.requiredHash` and `<realm>.claURL`
4. If the queries succeed, gnokey formats a helpful message with the signing
   command and prints it before returning the error.

The helper includes:
- An explanation of what the CLA is
- A link to the CLA document (if available)
- A ready-to-use `gnokey maketx call` command with the correct hash

### Graceful degradation

If any query fails (realm not deployed, network issue, variable renamed),
no helper is printed. The original error is always returned regardless.

## Alternatives Considered

### Keeper-side: CLAUnsignedError + InfoKV + OnTxError

We considered having the VM keeper return a structured `CLAUnsignedError`
carrying CLA metadata (hash, URL), populating the ABCI `Info` field with
key-value pairs, and adding an `OnTxError` callback to tm2 for gnokey to
parse. This was rejected because:

- It required changes across four layers: keeper, handler, tm2, and gnokey.
- It introduced a new error type, proto definition, and amino registration.
- It required an `OnTxError` callback in tm2's `BaseOptions` — a new concept
  in the broadcast layer just for CLA helpers.
- The ABCI `Info` field needed a custom key-value format with newline
  sanitization to prevent injection.
- The keeper already knows the CLA result (pass/fail) but had to spin up
  additional VM machines to read realm state for the error metadata.

The client-only approach achieves the same user experience with changes
confined to a single package (`keyscli`), using existing query infrastructure.

### Exported getter functions on the CLA realm

Adding `GetRequiredHash()` and `GetCLAURL()` to the CLA realm would let
any client call them. Rejected because:
- gnokey can already read unexported variables via `vm/qeval` (which evaluates
  expressions in the realm's package scope).
- Adding public getters just for error reporting would be unnecessary API surface.

## Consequences

- Users get clear, actionable feedback when CLA signing is required.
- Zero changes to the keeper, handler, tm2, or proto definitions.
- The helper depends on string-matching the error message
  `"has not signed the required CLA"`, which is fragile if the message
  changes. This is acceptable because the integration test already
  depends on the same string.
- Two extra network round-trips on the error path (param query + realm query).
  This is acceptable since it only happens on failure and the user is about
  to read and act on the output anyway.
- Other clients (gnoweb, wallets) would need to implement similar logic
  independently. If this becomes a pattern, a structured error approach
  could be revisited.
