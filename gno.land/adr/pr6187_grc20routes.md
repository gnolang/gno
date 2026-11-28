# ADR: grc20routes, a MsgCall route table for GRC20 tokens (PR #6187)

## Context

`grc20reg` lists tokens but cannot move a user's balance, so every user write
must go to the token's own realm. The call shape differs per realm
(`wugnot.Approve(cur, spender, amount)` vs
`grc20factory.Approve(cur, symbol, spender, amount)`) and is not discoverable,
so wallets fall back to `MsgRun`, which mainnet gates. `grc20factory` is live,
hosts most of the registry, and cannot be redeployed, so a table only token
realms can write to would miss most tokens.

## Decision

`r/nt/grc20routes/v0` resolves a registry key to `{pkg_path, prefix, funcs}`
with three sources, in precedence order:

1. **declared**: the token's own realm calls `Register`. The key is derived
   from `cur.Previous().PkgPath()` and must already exist in `grc20reg`.
2. **proven**: anyone shows by effect that a realm is keyed by registry
   symbol: two different tokens of that realm carry prover-chosen nonces as
   allowances to this realm's address. One proof covers the realm.
3. **convention**: canonical names, no prefix.

Review (davd-gzl, 2026-09) tightened the proven path:

- The prover must be a signing user (`IsUserCall`). A realm prover could set
  its own allowances on any token through `grc20reg.Approve` and mark any
  two-token realm keyed without calling it.
- Tokens registered under a sub identity (`<path>#<sub>.<SYM>`) get no proven
  or conventional route (`source: "none"`): their destination is the host
  realm, where the same symbol may name the host's own token. Keys are split at
  the last dot, since a sub path may contain one.
- `RefuseKeyedProof` lets a realm drop and permanently block proofs about
  itself; per-token `Register` cannot, since a proof covers the whole realm.
- No top-level `Func(tokenKey, op)` getter: a name without the prefix builds a
  call one argument short on a keyed realm.

## Alternatives considered

- **Record the leading argument the prover passed** as the prefix. Rejected:
  the proof sees allowances, not arguments, so that value would be
  prover-chosen and unverified. A prover could publish `BAR` as FOO's prefix,
  turning an abort into an approval of the wrong token.
- **Admin / governance removal of proofs.** Not added; the package is
  admin-free by design. It remains the only remedy for a pre-existing realm
  that is mis-marked and cannot call `RefuseKeyedProof`.
- **Declaration only.** Covers no already-deployed realm.

## Consequences

Residual false proofs (separately-named approves per token; a realm keyed by
a non-symbol identifier proven off-procedure; a MsgRun-enabled signer, whose
ephemeral realm shares their address) produce routes that abort. No funds
move. Realms that can call this one fix it with `RefuseKeyedProof` or
`Register`; pre-existing ones cannot.
