# ADR: Phase 2 — Inert Package Storage with Oracle Activation

Companion to [PR #5885](https://github.com/gnolang/gno/pull/5885) (Phase 1: permissioned
code submission policy).

## Context

Phase 1 adds a `permissioned` policy that restricts `MsgAddPackage` / `MsgRun` to an
allowlist. Phase 2 enables permissionless submission while keeping the Go typechecker
off the critical path: packages are stored in an **inert** state (no typechecking, no
execution) and activated later by a trusted approver — possibly an off-chain oracle.

## Decision

### New policy value

`code_submission_policy = "inert"`: any address may submit a package, but it is stored
in a separate key space (`inert_pkg:<path>` in iavlStore) that is invisible to the
normal package resolver. The package is not typechecked or executed at submission time.

### New param

| Param | Type | Default | Description |
|---|---|---|---|
| `pkg_approvers` | []address | `[]` | Who may call `MsgEnablePackage` / `MsgDisablePackage` |

### New messages

**`MsgEnablePackage { Approver, PkgPath }`**  
Approver must be in `pkg_approvers`. Chain retrieves the inert package, runs the
typechecker (oracle is untrusted for correctness), executes initialization, and moves
the package to the active store. After this point the package is importable.

**`MsgDisablePackage { Approver, PkgPath }`**  
Interface stub. Full disable requires evicting executed objects from the base store;
not yet implemented. Returns an error until a follow-up PR delivers it.

### Store layer

`gnovm.Store` gains three new methods:
- `AddInertPackage(mpkg)` — store at `inert_pkg:<path>` in iavlStore
- `GetInertPackage(path)` — read from `inert_pkg:<path>`
- `DelInertPackage(path)` — remove from `inert_pkg:<path>`

These keys are disjoint from `pkg:<path>` so normal `GetPackage` / `GetMemPackage`
never see inert packages.

## Testing

`gno.land/pkg/sdk/vm/keeper_inert_test.go` exercises the full oracle-activation
flow end-to-end:

- **`TestVMKeeperInertPackageLifecycle`** — policy `inert` + one approver: an
  untrusted user submits a package (stored inert, not resolvable, not callable),
  a non-approver is rejected, enabling an unknown path fails, then the approver
  (oracle) enables it — the chain typechecks + executes, the package becomes
  resolvable and callable, and the inert copy is removed.
- **`TestVMKeeperEnablePackageRejectsInvalidCode`** — "the oracle proposes, the
  chain enforces": ill-typed code is accepted inert but rejected on-chain at
  enable time, so it never becomes callable.
- **`TestVMKeeperDisablePackageNotImplemented`** — `MsgDisablePackage` is
  approver-gated but returns an error pending the follow-up PR.

## Consequences

- **Permissionless submission, deferred typechecking**: the DoS surface from the
  typechecker is removed from block execution time.
- **On-chain correctness guarantee**: the chain re-runs the typechecker at
  `MsgEnablePackage`; the oracle cannot activate a package that fails typecheck.
- **Default unaffected**: `code_submission_policy` still defaults to
  `"permissionless"`. Chains that don't opt in see no change.
- **Disable deferred**: MsgDisablePackage is stubbed; implementation requires a
  strategy for cleaning up executed objects from the base store.
