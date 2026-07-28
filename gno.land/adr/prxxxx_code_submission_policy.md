# ADR: Code Submission Policy

## Context

Both `MsgAddPackage` and `MsgRun` feed user-supplied code through the Go
typechecker synchronously during transaction delivery. The typechecker has
superlinear performance on adversarial input and no meaningful gas bound, so
a single malicious transaction can consume excessive block-execution time.

Chain operators need a way to restrict code submission to a trusted set of
addresses while the permissionless path matures.

## Decision

Add two parameters to the `vm` module params (`vm:p`):

| Param | Type | Default | Description |
|---|---|---|---|
| `code_submission_policy` | string | `"permissionless"` | `"permissionless"` or `"permissioned"` |
| `code_submitters` | []address | `[]` | Allowlist of addresses; checked when policy is `"permissioned"` |

An ante-handler check (`checkCodeSubmissionPolicy`) runs **after** signature
verification. When `code_submission_policy` is `"permissioned"`, any
`MsgAddPackage` or `MsgRun` whose signer is not in `code_submitters` is
rejected with `ErrUnauthorized` before the typechecker is invoked.

Both `CheckTx` and `DeliverTx` paths are covered because the ante handler
runs at both stages, which also prevents unauthorized transactions from
entering the mempool.

## Consequences

- **Permissionless (default):** no behaviour change; existing chains and
  devnets are unaffected.
- **Permissioned:** governance (GovDAO or equivalent) manages the
  `code_submitters` allowlist via param proposals. Only listed addresses can
  call `MsgAddPackage` / `MsgRun`.
- The policy is a chain parameter, so it can be toggled without a hard-fork.

### Setting `code_submitters` (repeated / string-array param)

`code_submitters` is a **repeated** param (`[]crypto.Address`). It is stored as
a JSON array and `GetParams` decodes it element-wise back into the typed field.
It must therefore be set through the *strings* param path — from governance,
`params.NewSysParamStringsPropRequest("vm", "p", "code_submitters", []string{...})`
(→ `ParamsKeeper.SetStrings`). Each entry is validated verbatim (a valid,
non-duplicate, non-zero bech32 address; no trimming) so that the value always
round-trips: an entry that passed validation but failed to decode would make
every subsequent `GetParams` (and thus every ante check) panic. A
comma-separated single string is **not** supported.

### Ordering caveat (avoid self-lockout)

Because `MsgRun` is itself gated, flipping the policy to `"permissioned"`
before adding the intended submitters would prevent those submitters from
issuing the governance `maketx run` transactions needed to fix it. Always add
addresses to `code_submitters` **first** (while still permissionless), then set
`code_submission_policy = "permissioned"`.

### Testing

- `gno.land/pkg/sdk/vm/code_submission_policy_test.go` — param validation,
  governance setter parsing, and the genesis/governance storage round-trip.
- `gno.land/pkg/gnoland/code_submission_policy_test.go` — the ante-handler
  enforcement matrix (policy modes, message types, multi-signer/multi-message).
- `gno.land/pkg/integration/testdata/code_submission_policy.txtar` — end-to-end
  through a real node: govdao param changes, then authorized vs unauthorized
  `addpkg`/`run`.

## Alternatives Considered

- **Handler-level check** (inside `handleMsgAddPackage` / `handleMsgRun`):
  rejected because the check would happen *after* typechecking, not before.
- **ValidateBasic extension**: cannot read chain state (params).
- **Phase 2 — oracle-based permissionless** (packages stored inert, activated
  off-chain): deferred; this ADR implements Phase 1 only.
