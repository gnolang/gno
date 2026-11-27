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
64 pending proposals, 128 retained proposals (older finished ones are
dropped; an executed call that has not been consumed is kept), 1024
safes in the factory, 8 watched denoms, 8 watched tokens, 30-day max
delay, 366-day max lifetime, 280-byte memo, 512-byte payload. A denom
has to match the chain's denom grammar before it is stored or passed
to `GetCoin`.

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

`v0` is unaudited by a third party. The review below is the author's
pass over the paths a genesis treasury would depend on. It is not a
substitute for that audit. Do not point genesis or GovDAO at this
realm on the strength of this note.

## Security review

Threat model. The attacker can create safes, call every exported
function, deploy a realm that calls this factory, pass a `cur` they
control into a signer they have compromised, send arbitrary coins and
denoms to any address, and fill storage up to the caps. They cannot
forge `cur.Previous()`, mint a banker for an address they do not hold,
or make `Execute` call a realm of their choosing. Signers are assumed
honest about proposals they approve and careless about realms they
pass `cur` into. The chain is assumed to run this code as written;
there is no admin key.

### Signer check

`auth` requires `rlm.IsCurrent()` and then takes
`rlm.Previous().Address()`. That is the realm that was current when
this frame was entered, not `OriginCaller`. A callee the user called
first does not inherit the user's vote. `maketx run` is an ephemeral
realm: `IsUser()` would accept it, and this code does not use
`IsUser()`. Deposit uses `IsUserCall()`, which is false for `maketx
run` and for a realm in the middle of the stack. A realm may itself be
a signer; that is intentional, and a signer who forwards `cur` into
untrusted code can be induced to approve. `Execute` does not call out,
so the safe does not add a second way to do that.

### Sub-realm addresses

`Bind` stores `chain.PackageAddress(host + "#" + subpath)`, the same
derivation as `cur.Sub(subpath).Address()`. Coin and GRC20 sends go to
that address. `applyCoin` and `applyGRC20` derive the sub again and
panic if it does not match the stored address. With `Bind` as the only
writer, the two cannot diverge; the panic is a tripwire. Two safes
have different subpaths, so one safe's sub-token cannot send the
other's coins. The factory's own address is not a wallet. `MsgCall
-send` credits the called realm's primary address, which is why
`Deposit` exists and why it refuses every caller that is not the
signing user.

### GRC20 teller

The registry key is `path.symbol`. `canAfford` loads it with
`grc20reg.Get` and reads `BalanceOf` on the safe's address. A token
that was never registered fails here, and the proposal stays pending.
`applyGRC20` builds `RealmTeller` with the sub-token, refuses anything
`IsCanonicalTeller` rejects, and `Transfer`s. `Transfer` does not call
the recipient. A teller minted for a different realm cannot pass the
canonical check. Tokens outside `grc20reg` cannot be spent; that is a
v0 limit, not a bypass.

### Consume

`Execute` of a call stores the payload and returns. It does not enter
the target, so a callback cannot reenter on the safe's `cur` or on a
banker minted from it. `Consume` requires `Previous().PkgPath()` to
equal the proposal's target, refuses a second pull, and refuses to run
while the safe is paused. The target package has to be a realm path,
and it cannot be the factory or a `host#...` subpath. The last rule
matters if a later version ever calls out: a target of the factory
would otherwise be an authorization to act as the factory.

### Pause

Pause blocks coin, GRC20, and call execution, and it blocks `Consume`.
Signer changes, threshold, delay, and unpause still run, so a paused
safe can still be repaired. Pause does not freeze a timelock that
already started. After unpause, a spend whose delay has elapsed can
execute in a later message. It cannot execute in the same message as
the unpause, because the unpause is the proposal that ran.

### Render

Names, memos, addresses, denoms, and token keys go through
`sanitize.InlineText` before they are written into the page. Paths
that do not parse render a fixed "unknown" string. Balances use
`GetCoin` for one denom. Watched denoms are checked with `ValidDenom`
again before `GetCoin`, so a denom that somehow failed the write-time
check cannot panic the page.

### DoS and storage

`GetCoin` panics on a denom the chain rejects. Storing one, from
`WatchDenom` or a coin proposal, would brick `Render` and `Balance`
for that safe. `ValidDenom` now uses the chain grammar (length 3 to
274, first byte `a-z` or `/`, remaining bytes the banker's set) and
both writers use it. Regression: a denom such as `NO` is rejected, and
`Render` skips a watched denom that fails the check.

Retained proposals used to grow without a bound. Loading a safe walks
that list, and so does stripping a removed signer's approvals. The cap
is 128. Past that, the oldest terminal proposal is dropped.
`ProposalForgotten` is emitted. An executed call that has not been
consumed is not terminal: the target can still pull it. If every
retained record is one of those, `Propose` returns `ErrHistory` until
one is consumed. Pending proposals are capped at 64 before the history
check, so a full pending set is `ErrTooManyPending`, not `ErrHistory`.
Ids stay monotonic and are not reused. After a drop, `View` of that id
is `ErrUnknownProposal`. Chain events are the record that outlives the
cap.

`Create` is permissionless. An unbounded factory list is a storage
sink every `Render` of the home page walks. The ceiling is 1024 safes.
A deposit that the creator can withdraw does not price this, so the
ceiling is the control. Filling it blocks later creates. That is an
accepted v0 limit.

One signer can occupy all 64 pending slots until they expire (at most
366 days). Any other current signer can cancel a proposal that has not
reached the threshold. A 1-of-1 safe can only fill its own queue.
Watched lists are capped at 8. Signer sets are capped at 32. Loops
that walk proposals, signers, or watches are bounded by those caps.

### Findings

| Severity | Issue | Disposition |
| --- | --- | --- |
| Medium | A denom `GetCoin` panics on could be stored and then panic `Render` | Fixed. `ValidDenom` before the write. Regression in the package and realm tests. |
| Medium | Finished proposals were retained forever, so one safe's stored list and its approval walk were unbounded | Fixed. 128 retained, oldest terminal proposal dropped, unconsumed calls kept. Regression: fill the cap, observe the oldest id disappear, observe `ErrHistory` while every retained call is unconsumed. |
| Medium | `Create` had no ceiling | Fixed. 1024 safes. Regression: a full list panics `ErrTooManySafes` without calling `New`. |
| Low | A call target of the factory, or `host#...`, could be authorized | Fixed. `validCallTarget` rejects both. The safe still does not call out. |
| Low | `New` accepted an invalid creator address | Fixed. `ErrInvalidAddress`. |
| Info | `armedAt == 0` means "not armed". A proposal armed at unix time 0 cannot execute | Accepted. Chain time is not 0. A test locks the current behavior. |
| Info | Second balance, teller, and address-match checks in `applyCoin` / `applyGRC20` sit behind `canAfford` and `Bind` | Kept. They are unreachable when the first check holds. They panic rather than mark the proposal executed. |
| Info | `expireIfDue` also marks a pending proposal expired, and `sweep` has already done that for every caller | Kept. A later caller that skips `sweep` still expires the proposal. |
| Info | Any current signer can cancel a proposal that is under the threshold | Accepted. An armed proposal cannot be cancelled; signers revoke their own votes first. |
| Info | A realm signer can be talked into approving by code it passed `cur` to | Accepted. Same class as a leaked key. `Execute` does not call out. |

Statement coverage, measured by instrumenting each block-level
statement with an event and running `gno test` (there is no `-cover`
flag): `p/nt/multisig/v0` 523/531 (98.5%), `r/nt/multisig/v0` 241/251
(96.0%), or 242/251 counting `init`, which runs before the test event
log starts. The statements not hit are the defensive branches in the
table above, plus `parseSigners` on an empty `strings.Split` result
(`Split` does not return an empty slice) and `rel` when the realm path
does not start with `gno.land`.

## Testing note

`testing.SetRealm` updates the frame it is called from. A helper
between the test and `SetRealm` is the frame that gets the override,
and the test's `cur` stays unchanged. Call `SetRealm` directly in the
test function, then `cross(cur)`. Cross-realm panics are aborts;
`revive` observes them in filetests, `recover` does not.
