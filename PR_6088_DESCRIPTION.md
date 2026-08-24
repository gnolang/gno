Builds on #5888 and contains it, so this supersedes that PR — please review this
one instead. Base is master, which #5888 predates by about a month.

Under the inert policy a package is parked when submitted and only runs once an
approver enables it. Storage gets written at both points, and this fixes who
pays and when:

- The deposit ceiling is recorded at submit time and honored at enable, so a
  later change to the chain default cannot change what a submitter already
  agreed to.
- The namespace and CLA checks moved before the type-check. They ran after the
  object cache was cleared, so the creator was charged nothing.
- Deleting the old package moved below the inert branch. The two orderings
  disagreed on stored state, which is a fork.
- Re-submitting a parked package is now limited to its original creator.
  Overwrites are silent, so anyone could clobber a pending submission.
- Parking is skipped at genesis, where there is no approver and the chain
  would fail to boot — and skipped during genesis **replay**, where the fork's
  new policy would otherwise re-route a previous chain's history and boot with
  its own realms parked instead of deployed.
- `MsgEnablePackage` is valid only while the policy is `inert`, and re-applies
  the full gnomod rule set. Without either, a package parked during an `inert`
  era stayed activatable forever under any later policy.
- A submission under `inert` may not carry `Send`. The coins would move at
  submit while `init()` runs at enable, so a payable `init()` panicked on the
  origin-send recipient mismatch — the same source deployed under
  `permissionless` and failed under `inert`.

Other changes:

- **MsgRun allowlist.** New `run_submitters` param. **Empty means today's
  behavior: anyone may send it.** Listing one address turns the gate on. This
  is deliberately the opposite reading from `code_submitters`, which is only
  consulted once `code_submission_policy` has been explicitly moved to
  `permissioned` and so can safely treat empty as "nobody". `run_submitters`
  has no such switch in front of it, so a fail-closed empty value would disable
  MsgRun — and with it GovDAO proposal creation, which is MsgRun-only — on
  every chain that upgrades without editing genesis.
- **Policy enforcement.** The check that decides whether a transaction carries
  code missed `MsgEnablePackage` and `MsgDisablePackage`. It is now an explicit
  list that fails closed for message types added later.
- **Default deposit 600M → 100M ugnot.** This is a ceiling, not a charge. The
  largest real deploy among all 321 genesis packages is 276,098 bytes, so the
  new ceiling clears the worst case by about 3.6x.
- **Simulation now verifies signatures on code-bearing transactions.** It did
  not, so a forged signature naming any address bought a full type-check and
  `init()` run off one unauthenticated request.
- **Parameter delegation.** GovDAO can hand one parameter to one other realm
  and take it back at any time. The delegate may only remove entries it added
  itself, and may never take the list to zero — an empty `run_submitters` means
  the gate is off, so emptying it would let the delegate unilaterally revoke the
  restriction GovDAO voted for. `r/nt/commondao` leaves quarantine, since it is
  the intended delegate.
- **gpao fails closed.** It ignored whether the transaction it read actually
  succeeded, so a rejected submission could still be approved.

Merging master also surfaced a determinism bug: `EnablePackage` type-checked
with a test-file getter, which reads a test-stdlib overlay off the node's local
disk. #6025 removed exactly that from `AddPackage` for the same reason. Fixed
here.

Consensus-breaking: new and changed vm params, the storage-charging order, and
the genesis package set (`r/nt/commondao` leaves quarantine). The pinned
multistore hash was re-derived by running the test.

ADRs are included for the inert charging work and the delegation design.
AI-assisted, reviewed by me before pushing.
