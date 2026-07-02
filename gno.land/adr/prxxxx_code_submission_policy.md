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

## Alternatives Considered

- **Handler-level check** (inside `handleMsgAddPackage` / `handleMsgRun`):
  rejected because the check would happen *after* typechecking, not before.
- **ValidateBasic extension**: cannot read chain state (params).
- **Phase 2 — oracle-based permissionless** (packages stored inert, activated
  off-chain): deferred; this ADR implements Phase 1 only.
