# ADR: M-of-N multisig treasury

## Status

Proposed, with a v0 implementation under `examples/`. Unaudited.
Not wired to genesis or GovDAO.

## Context

[gnolang/gno#5798](https://github.com/gnolang/gno/issues/5798) recorded the
launch allocations and asked for multisigs to hold them. The issue is
closed. [thehowl's closing note](https://github.com/gnolang/gno/issues/5798#issuecomment-6876374799)
says the numbers were reconciled in `gnolang/independence-day`, vesting
landed, and the multisigs were split into purpose-specific addresses
before gnoland-1 launched. Those addresses are tm2 accounts
(`gnokey` multisig keys, `std` multisig pubkeys). They are not a realm.

What #5798 actually required of a multisig, from the issue body and the
linked launch items:

- GovDAO T1 holds Core Treasury (40M), Ecosystem Treasury (60M), and
  Validator Treasury (20M): 120M GNOT. [#5151](https://github.com/gnolang/gno/issues/5151)
  asked for the NT1, NT2, and GovDAO addresses so genesis could pay them.
- NT multisig 1 holds NT,LLC (332M) and Investors (300M): 632M GNOT.
- NT multisig 2 is the special case for the AIB airdrop slice.
- Vesting is a separate send-limit (`must_keep`), not a job for this
  realm. [#5566](https://github.com/gnolang/gno/pull/5566) and the later
  edit in #5798 set the schedule (4% at transferability, then 4% a month
  for 24 months, with 150M of the investor allocation unlocked at once).

[#4094](https://github.com/gnolang/gno/issues/4094) asked for a reusable
treasury whose spends are approved, with room for custom rules. That
landed as `r/gov/dao/treasury` plus `p/nt/treasury`: a payment rail that
only an allowlisted DAO realm may call. `r/nt/commondao` is the other
existing treasury. It is a council and a vote, with per-DAO sub-realm
addresses, not an M-of-N signer set.

Account multisig and a realm multisig answer different questions. A tm2
multisig signs a transaction. It can send coins and call one realm
function, and every signer has to assemble that transaction off-chain.
It cannot hold a GRC20 balance under its own policy, queue a proposal
for later, or show its state in `Render`. A realm can. It cannot, on
today's VM, impersonate an arbitrary transaction the way a pubkey
multisig can.

No open PR or issue is building this realm. Searches of `gnolang/gno`
and `faddat/gno-1` for an in-flight multisig realm, gnosafe, or m-of-n
treasury returned nothing to coordinate with. The nearby open issues
are different work: [#757](https://github.com/gnolang/gno/issues/757)
(no dynamic `std.Call` in v1) and the inert-approver cost discussion.
If another multisig realm shows up, review this factory instead of
starting a second one, and leave GovDAO's vote treasury alone until
someone explicitly wants a proposal kind that spends through a safe.

`cur.Sub` exists (see `gnovm/adr/pr5890_realm_sub.md` and the native
method on `.grealm`). A sub-token's address is
`chain.DerivePkgSubAddr(host, subpath)`. Sub-tokens may only construct
a `BankerTypeRealmSend` banker. The token itself cannot be stored.
`std.Call` / dynamic import was rejected for v1 in #757, so a proposal
cannot name an arbitrary function and invoke it.

`std.Emit` moved to `chain.Emit` when std was split. Events in this
change use `chain.Emit`.

## Decision

Namespace is `nt`, version `v0`, next to `p/nt/treasury`, `p/nt/grc20`,
and `p/nt/ownable`:

- `gno.land/p/nt/multisig/v0` is the state machine. It never keeps a
  realm value and never calls untrusted code.
- `gno.land/r/nt/multisig/v0` is the factory. `Create` is permissionless.

Each safe is `cur.Sub("s/<id>")`. Coins and GRC20 sent to that address
are the safe's funds. The factory's own address is not a wallet.
`Deposit` exists only for the `gnokey maketx call -send` habit: it
forwards the message envelope with `BankerTypeOriginSend`, and only
when `cur.Previous().IsUserCall()` so a realm in the middle cannot
spend the caller's envelope. A realm that already holds coins sends
them straight to `SafeAddress`.

Signer identity is `cur.Previous().Address()` on a live `cur`. An EOA
must `maketx call`. `maketx run` shows up as an ephemeral realm and is
not the user. A realm may be a signer; the immediate caller is what
counts, not `OriginCaller`, so a callee cannot inherit the user's
authority.

Proposals cover native coins, a GRC20 registry key, a one-shot call
authorization, signer add/remove, threshold, timelock, pause, and
unpause. MsgCall cannot pass a slice, so signers are a comma-separated
string.

There is no arbitrary cross-realm call. Execute of a call proposal
stores a payload. The target pulls it once with `Consume`, which
requires `cur.Previous().PkgPath()` to equal the proposal's target.
The safe never enters the target, so a callback cannot reenter with
the safe's `cur` or a banker minted from it. The realm passes the sub
token only to `grc20.RealmTeller` on a `*grc20.Token` from `grc20reg`,
then checks `IsCanonicalTeller`. A spend that the safe cannot afford
is rejected before the proposal is marked executed.

The timelock starts when the current signers first reach the threshold
(`armedAt + delay`). Revoking under the threshold clears `armedAt`, so
the delay starts over. `delay` on the safe applies to proposals created
afterwards; a pending proposal keeps the delay it was born with.
`expiresAt` must fall after `now+delay` and within 366 days. A proposal
that has reached the threshold cannot be cancelled; signers revoke
their own votes first. Any current signer may cancel one that has not.
Removing a signer deletes that address's approvals, including if the
same address is added again later. Pause blocks coin, GRC20, and call
execution. Signer changes and unpause still run. Proposal ids are never
reused, and `Nonce` counts successful executions.

`Render` lists safes, signers, `ugnot`, watched denoms, watched GRC20
keys, and proposals. User text goes through `sanitize.InlineText`.
Balances read `GetCoin` for one denom. They do not call `GetCoins`,
because a third party can stick extra denoms on the address and set
the gas cost.

Caps, chosen so one safe cannot be walked without a bound: 32 signers,
64 pending proposals, 8 watched denoms, 8 watched tokens, 30-day max
delay, 280-byte memo, 512-byte payload.

## Alternatives

- **Only the tm2 multisig.** Already shipped for genesis. It does not
  give GRC20, a queue, or a page. Kept as the launch mechanism; this
  realm does not replace those accounts.
- **Extend commondao.** That is a vote and a council. An M-of-N signer
  set is a smaller object. Mixing them would make every safe pay for
  DAO machinery, and commondao's execution closure can retain a
  `RealmSend` banker (see `pr6012_commondao_treasury_sub.md`). This
  factory never hands a banker to caller-supplied code.
- **One shared balance inside the factory.** Simpler, and one bug
  attributes a deposit to the wrong safe. Sub-realm addresses make the
  bank the ledger.
- **Dynamic call (`std.Call`).** Not available, and #757 says v1 will
  not add it. The pull authorization is the pattern that works now.
- **`p/demo` or a personal namespace.** `demo` is examples; a personal
  namespace hides a primitive other realms are meant to import. `nt` is
  the reusable set. Moving it after an audit is cheap because v0 is
  explicitly unaudited.

## Consequences

Anyone can create a safe and point coins at its address. GovDAO and the
NT accounts do not move by deploying this. Pointing 120M GNOT at a new
realm is a separate governance decision.

A signer who is a realm, or who passes `cur` into code they do not
trust, can still be induced to approve. That is the signer's bug, same
as handing a key to a program. The safe does not make it worse by
calling out during `Execute`.

`v0` is unaudited. Production use, and an upstream PR that genesis or
GovDAO would depend on, waits on review of the signer rules, the
sub-realm address match, the GRC20 teller path, and the pause switch.

## Testing note

`testing.SetRealm` updates the frame it is called from. A helper
between the test and `SetRealm` is the frame that gets the override,
and the test's `cur` stays unchanged. Call `SetRealm` directly in the
test function, then `cross(cur)`. Cross-realm panics are aborts;
`revive` observes them in filetests, `recover` does not.
