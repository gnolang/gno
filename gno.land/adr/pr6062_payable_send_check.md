# Reject coins nobody looked at

## Status

Implemented. Verified against the full test suite.

## The problem

You can attach coins to a call. They are moved into the address of the
realm you called, before that realm's code runs. We call them the
envelope.

Nothing required the realm to notice. If its code never looked at the
envelope, the coins simply sat in its address. There was no error, no
refund, and often no way to get them back. The caller believed they had
paid for something. The realm had no idea it had been paid.

There is no way to work this out ahead of time. Gno has interface calls
and function values, so whether a given call will end up reading the
envelope cannot be decided by reading the code. Any safe approximation
would mark most of the ecosystem as accepting payment.

## The decision

Work it out while the call runs, and use the only definition that is
actually true: **the code read the envelope**.

Two places in the code that runs on chain make the envelope visible:

1. `unsafe.OriginSend()`, which returns it directly.
2. The spending limit check inside an OriginSend banker.

Both now set a flag. After a `MsgCall` finishes, if coins were attached
and the flag was never set, the whole message fails and the coins go back.

The second place matters. A realm that simply forwards the payment on
never calls `unsafe.OriginSend()` — it only ever touches the envelope
through the banker. Marking in one place only would have rejected the most
ordinary payment-handling realm there is.

A third place reads it, but only in tests: `testing.GetContext()`
hands the whole context, envelope included, to test code. It deliberately
does not set the flag, and it does not need to — the payable check only
runs on the real chain, and the test harness is not reachable from there.

It is left alone for a practical reason. Every `testing.Set…` helper works
by reading the context, changing one field, and writing it back. Blanking
or refusing the envelope on the way out would make `SetHeight` silently
erase a payment a test had just set up. The cost is a small fidelity gap:
a test that reads the envelope this way is not exercising the payable
path, and cannot tell.

### Only the realm that was paid may vouch

The first version of this marked the flag whenever anyone read the
envelope, at any depth. That was wrong.

Any realm can read the envelope, not just the one the coins went to. So a
realm that never looked could still have its call accepted, because
something it called happened to look. The coins stranded in the first
realm while the second one believed it had been paid. We reproduced this.

Now the flag is only set when the realm doing the reading is the realm the
coins went to. The runtime already tracks which realm is executing, so
this is a single comparison — no stack walking, and no change to gas
costs.

The banker path is deliberately different. It marks without checking who
is running, because by that point it has already proved the coins being
spent are the ones that were paid in. The owner of a banker may hand it to
another realm to spend within the same message, so the realm running at
that moment is not necessarily the one that was paid. Building a banker
over your own envelope is itself proof you noticed the payment.

## Scope

`MsgCall` only.

`MsgAddPackage` is exempt: coins attached there land in the new package's
own address, and a realm can spend from its own address later, so they are
not lost.

That exemption was too broad, and it hid a second bug. A pure library
package has no realm identity, so it can never obtain a banker and can
never spend from its address. Deploying one with coins attached used to
succeed, and those coins were gone permanently.

**Fixed separately, in the commit that follows this one.** `MsgAddPackage`
now rejects a non-empty payment when the path is not a realm. Realms still
accept coins on deploy, because a realm can spend from its own address
later. See `addpkg_unspendable_send.txtar`.

`MsgRun` is exempt: the coins are moved from the caller to the caller, so
nothing actually moves.

## Alternatives considered

**Mark a function as accepting payment, like `payable` in other
languages.** Cannot be checked. See above.

**Track whether the coins were spent instead of whether they were seen.**
Wrong test. A realm that takes a deposit and holds it never spends the
envelope. It only checks the amount and keeps it. That is a normal, honest
realm and this would have rejected it.

**Return the envelope as empty to realms that were not paid, instead of
tracking who looked.** Dangerous. A realm asking "was I paid?" would
silently read zero and take its "no payment" branch while the coins sat
elsewhere. We tested a version of this and it quietly changed the meaning
of five existing payment tests without any of them failing.

**Leave it alone.** Coins keep getting stranded with no signal.

## A known limitation, in both directions

"Which realm is running" is answered by the realm the VM is currently
operating in. That is normally the realm whose code is executing, but not
always: when a realm calls a method on an object owned by a *different*
realm, the VM switches to the owner for the duration. That is deliberate
and correct for storage, but it means the answer can be the owning realm
rather than the calling one.

So there are two ways to get the wrong answer:

- **A realm gets credit it did not earn.** A realm deeper in the chain
  reads the envelope through a shared helper owned by the realm that was
  paid. The read is attributed to the owner, and the payment is accepted
  even though the owner never looked. The coins are still stranded, which
  is the thing we were trying to catch.
- **A realm is refused credit it did earn.** The realm that was paid reads
  its own envelope through a shared helper owned by *another* realm. The
  read is attributed to that other realm, the flag never gets set, and a
  perfectly good payment is rejected.

The second is worse, and it is worth being precise about why. This is a
safety net against losing coins, not a security boundary. Missing a
stranding means an accident goes uncaught — bad, but it is the situation
we were already in. Rejecting a real payment breaks a working system, and
does it with a message that blames the wrong thing.

That asymmetry is an argument for erring loose rather than strict, which
the earlier version of this check did — and which had its own problem, the
callee-vouching hole described above. There is no cheap setting that
avoids both. Worth revisiting deliberately rather than inheriting.

The refusal case is reproducible in a few lines. A `/p/` library exports a
shared object whose method reads the envelope; a realm that genuinely
requires payment calls that method instead of reading directly. Same
realm, same coins, same requirement — reading directly is accepted,
reading through the shared object is rejected.

**Do not read too much comfort into the fact that no `/p/` package in this
tree does that.** We checked all 609 of them and none reads the envelope,
which is why nothing breaks today. But `/p/` packages are user-deployable.
Anyone can publish a shared payment helper tomorrow, and a shared helper
that checks "was I paid" is a natural thing to write. The first person who
does it breaks payments in their own realms, with an error message that
points at the wrong thing.

So the honest statement is: nothing in the tree triggers it, and nothing
prevents it.

An exact answer needs to know whether the paid realm is anywhere on the
call stack, which a single comparison cannot tell. Doing it properly means
tracking that as the VM pushes and pops frames — more than this change was
scoped for, but the shape of the real fix.

We accepted this rather than walking the call stack to get an exact
answer, because the exact version costs a stack walk on every read and
would change the gas table.

Both directions were demonstrated with throwaway tests during review. They
are not checked in: they assert the buggy behavior, so they would fail the
day someone fixes it, and they need a `/p/` helper that reads the envelope
— which is exactly what does not exist. Rebuilding one is a few minutes'
work from the description above.

**If anyone ever writes a `/p/` helper that reads the envelope, revisit
this.** That is the trigger.

## Consequences

- Some calls that used to succeed now fail. This is the point, but it is a
  change in what counts as a valid transaction.
- Seventeen edits across nine test files. No realm source changed.
  - Fourteen were one pattern: using an attached payment to *fund* a
    realm's address rather than to pay it. Those realms can spend their
    own balance, so nothing was ever stranded — they just never looked.
    They now acknowledge the payment. Funding an address without calling
    it is still possible with an ordinary transfer, which does not go
    through this path at all.
  - Two dropped an attached payment that was never meaningful. One was a
    test fixture whose function has no notion of being paid. The other is
    the `valopers` registration test — see below.
  - One switched a call to send nothing, in a test that asserts nothing
    about balances and had copied the payment from elsewhere.
- The `valopers` realm reads the envelope only when a registration fee is
  configured, and that fee currently defaults to zero. Attaching a payment
  while the fee is off is now rejected. That is the correct outcome: those
  coins would have been stranded. Its test was updated to stop attaching
  one. If a test ever turns the fee on, the payment goes back.
- The app hash does not change **for this commit**. Everything here is Go,
  which is compiled into the node rather than stored on chain. (The banker
  fix in the preceding commit does move it, because it edits Gno standard
  library source, which is genesis state.)
- Gas costs do not change. The new check is one string comparison.
- This is a safety net against losing coins, not a permission check. A
  realm can opt in by reading the envelope and ignoring the result. That
  is fine — the goal is to catch accidents, not to police intent.

## What we checked

- `gno test ./examples/...` — 221 packages pass, none fail.
- The full integration suite passes.
- The VM, test-harness, and file-test suites pass.
- A new test pins the exact hole: a realm that never looks cannot be
  vouched for by one it calls.
- Reading the envelope from a deeper realm still returns the real number.
  This is about who may vouch, not about hiding anything.
