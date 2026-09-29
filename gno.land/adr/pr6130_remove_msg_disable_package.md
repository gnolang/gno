# Remove the unimplemented `MsgDisablePackage`

Closes gnolang/gno#6125.

## Context

The `inert` code-submission policy (#5888, extended in #6088) splits a
deployment in two: `MsgAddPackage` parks the source under `inert_pkg:`, and an
approver from `params.PkgApprovers` completes it with `MsgEnablePackage`, which
type-checks, runs `init()` and makes the package importable.

`MsgDisablePackage { Approver, PkgPath }` arrived in the same series as the
inverse operation — move a live package back to inert state — but it was never
implemented. Its keeper method checked the approver gate and then returned:

```go
// TODO: evict executed package objects from baseStore and move source back
// to inert_pkg key. Tracked in a follow-up PR.
return std.ErrUnknownRequest("disable_package is not yet implemented")
```

The follow-up never landed, and #6088's own review notes record why it is not a
small piece of work: it needs eviction of executed objects from the base store,
and the storage deposit taken at enable would have to be released. Neither has
a design.

Meanwhile the message was wired through every layer as if it worked:

- registered in the amino codec as `m_disable_pkg`, with generated `pb3_gen.go`
  marshal/size/unmarshal methods and an `m_disable_pkg` message in `vm.proto`;
- routed by `vmHandler.Process` to `handleMsgDisablePackage`;
- covered by `txCarriesCode`, so a tx containing it needs a verified signature
  even on the simulate path;
- denied to session accounts by `sessionAlwaysDenied`;
- matched by gnokey's `txNeedsSimulationSignature`, so `maketx` would sign a
  second transaction to simulate it;
- listed in `params.PkgApprovers`' doc comment as one of the two messages
  approvers may send.

So the whole surface existed — a type on the wire, a route, ante rules, client
support — for a message whose only possible outcome is an "unknown request"
error. Tests pinned that error rather than any behaviour:
`TestVMKeeperDisablePackageNotImplemented`,
`TestVMKeeperGenesisReplayDisablePassesTheApproverGate` and two subtests of
`TestVmHandlerProcessRoutesInertMsgs` asserted only that the gate was reached
and the stub refused.

Beyond being dead, the intended semantics are unclear. `MsgRejectPackage`
(#6088) already covers the operationally needed case — clearing a package that
is parked awaiting approval, either by an approver declining it or the creator
withdrawing it. What is left for disable is un-deploying *live* code, which
raises questions this series never answered: what happens to callers that
imported the package, to its realm state, to its storage deposit, and to the
addresses that paid it.

## Decision

Remove `MsgDisablePackage` and everything wired to it, rather than leave a
committed message type that cannot succeed.

Removed:

- `MsgDisablePackage` and its `std.Msg` implementation (`msgs.go`);
- `VMKeeper.DisablePackage` (`keeper_inert.go`);
- the `Process` case and `handleMsgDisablePackage` (`handler.go`);
- the `m_disable_pkg` amino registration (`package.go`), and with it the
  generated codec in `pb3_gen.go` and the `m_disable_pkg` message in
  `vm.proto` (regenerated with `make -C misc/genproto2`);
- the message from `txCarriesCode` and `sessionAlwaysDenied` (`app.go`) and
  from `txNeedsSimulationSignature` (`tm2/.../maketx.go`);
- the stub-only tests listed above, and the `disable_package` rows in
  `TestTxCarriesCode`, `TestSessionAlwaysDeniedMatrix`,
  `TestTxNeedsSimulationSignature` and the codec parity table.

Also fixed, because they only made sense while the stub existed:

- `VMKeeper.RejectPackage`'s doc comment (`keeper_inert.go`), which had a stray
  `// DisablePackage moves an active package back to inert state.` header glued
  above its own, and the matching `MsgDisablePackage` comment block stranded
  above `MsgRejectPackage` in `msgs.go`;
- `EnablePackage`'s note that "nothing evicts them in the meantime — see
  DisablePackage", which stopped being true when `MsgRejectPackage` landed;
- `Params.PkgApprovers`' doc, which listed enable/disable as the approver's two
  messages;
- the `txCarriesCode` doc block in `app.go` and `EstimateGas`' doc in
  `gno.land/pkg/gnoclient/client_txs.go`, both of which enumerated the message;
- a comment in `keeper.go` asserting that nothing reaches the
  full-deposit-refund branch "while `DisablePackage` is a stub". The branch
  predates the stub, so it stays; the comment now gives the reason that does
  not depend on it (a realm's storage includes its own package objects, which
  no message frees).

Nothing outside that set is touched. In particular `MsgEnablePackage` and
`MsgRejectPackage` are unchanged, including their tests: a first pass folded
`TestVmHandlerProcessRoutesReject` into `TestVmHandlerProcessRoutesInertMsgs`
(with disable gone the two are near-duplicates) and corrected several
pre-existing stale comments nearby, and all of that was reverted — it is
unrelated to this removal and belongs in its own change. Nothing here changes
behaviour: the message could not succeed, so no reachable code path loses a
capability.

## Alternatives Considered

- **Implement it.** Rejected for now: it needs base-store object eviction and a
  storage-deposit release policy, and the semantics for live-package
  un-deployment (importers, realm state, deposit refunds) are undecided. That
  is a design discussion, not a follow-up commit — and it does not need a
  placeholder message on the wire to start.
- **Keep the type, drop the route.** Rejected. A registered amino type with no
  handler is worse than either extreme: it stays decodable and estimable by
  clients but fails as "unrecognized vm message type", i.e. a routing error
  reported as a message-type error.
- **Keep the stub as a marker for the future work.** Rejected. #6125 is the
  marker, and it costs nothing on the wire. A stub, by contrast, is a codec
  entry, an ante rule and a client branch that all have to be kept correct.
- **Repurpose the name for a live-package operation later.** Deliberately not
  reserved. If a disable message is designed, it should pick its own fields and
  amino name then, and reusing `m_disable_pkg` for different semantics would
  make old and new transactions decode into the same type.

## Consequences

The `inert` flow now has exactly three messages: submit (`MsgAddPackage` under
`inert`), `MsgEnablePackage`, `MsgRejectPackage`. A package can still be enabled
only once, and no message releases the deposit that enable locked — ordinary
storage frees still refund proportionally, but to whoever triggers them rather
than to the original payer. That was already true, and now the tree does not
suggest otherwise.

`vm/disable_package` becomes an unknown message type: a transaction carrying it
is rejected by the codec rather than by the handler. Because the message never
had a success path, nothing that worked stops working.

There is one migration edge, and it is not purely hypothetical. The
`chain/pearl` release ref (testnet `pearl-1`, genesis 2026-08-27) registers
`m_disable_pkg` in its codec, so a `vm/disable_package` transaction *decodes*
on a pearl node and reaches the keeper, where it fails. That is enough: a
failed message is still committed into a block. (Pearl's own genesis does not
enable the `inert` policy — that lands on the pending
`tbruyelle/chain/pearl-with-inert-pkgs` branch — but the policy is irrelevant
here, since the message fails either way and is committed either way.)

A pearl node moved to a post-removal binary would therefore be unable to
amino-decode any block containing such a transaction, on replay at restart and
in anything that decodes history (indexers, gnoweb). Whether pearl's history
actually contains one is an empirical question about that chain's tx index, not
something this diff can settle; if pearl is upgraded across this change rather
than restarted from a fresh genesis, check first. No mainnet is affected —
there is none.

`PkgApprovers` now gates `MsgEnablePackage` and the approver half of
`MsgRejectPackage`.

Prior ADRs (`pr5888_phase2_inert_packages.md`,
`pr6088_msgrun_allowlist_and_inert_charging.md`,
`pr6088_param_delegation.md`) still describe the stub. They are left as written:
they record the state of the tree at their own PRs, and this ADR is the record
of the removal.
