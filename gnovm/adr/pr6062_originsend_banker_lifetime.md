# An OriginSend banker must not outlive its message

## Status

Implemented. Verified against the full test suite.

## The problem

When you send a transaction to a realm you can attach coins to it. Those
coins land in that realm's address. We call them the envelope.

A realm can ask for a "banker" to move coins. There are several kinds. The
`BankerTypeOriginSend` kind is meant to be the safe, limited one: it can
only spend the envelope that came with the current transaction. The other
kinds can spend everything the realm owns.

Two facts about the code did not fit together:

1. A banker was a plain value. You could store it and it would still be
   there in the next transaction. This was on purpose and was tested.
2. The spending limit was not stored inside the banker. It was looked up
   fresh every time, from whatever transaction happened to be running.

So a stored OriginSend banker did not stay limited to the envelope it was
made for. Every new transaction re-armed it against a new envelope. The
coins came out of the realm's own balance, not out of any envelope.

The check that a realm may create one of these bankers happens once, when
it is created. Nothing re-checked it later.

### Why this mattered

On its own this gave a realm nothing new. A realm can always ask for the
stronger `BankerTypeRealmSend` and spend its whole balance.

The problem was giving one away. Bankers are meant to be passed to other
realms — that is written in the docs. A realm handing out an OriginSend
banker believes it is handing out a limited, one-transaction permission.
It is not. The receiver can store it and reuse it in **any** later
transaction that carries coins, spending from the giver's balance up to
whatever that unrelated transaction happened to attach. The original code
never checked who the coins were paid to — only that the amount fit the
envelope of whatever message was running. Nothing in the type prevents
this, and the giver cannot revoke it.

We proved this end to end. Realm A hands a banker to realm B during a
legitimate call carrying a 1 coin envelope, and B stores it. In a later
message — carrying a 500,000 coin envelope that had nothing to do with A —
B uses the stored banker to take 500,000 coins out of A's own treasury.
The banker was armed for 1 coin and spent 500,000.

### Where it came from

This is a regression, not an original flaw.

The first version of this code (April 2022) stored the envelope and the
running total inside the banker itself. There was no lookup. The banker
could not be stored, and the docs said so.

A refactor in March 2024 removed a feature of the native-function
machinery. As part of that, the banker was reduced to a small value and
the limit check moved to the ambient lookup. Nothing in that change
suggests the security effect was noticed.

Later, storing bankers was documented as supported and given a test. But
that test only ever exercised the *other* banker kind. The support was
verified for one kind and then written up as applying to all of them.

## The decision

Make an OriginSend banker impossible to store.

The virtual machine already refuses to save realm handles — they are meant
to live only for one call. So we put a live realm handle inside the banker,
but only for the OriginSend kind. Any attempt to write such a banker into
realm storage now fails, and the whole transaction is rolled back before
any coins move.

The banker can still be passed to other realms *during* the same message,
which is the case the type exists for.

Other banker kinds get an empty field and are unaffected. They can still
be stored, exactly as before.

Underneath that we added a second, independent check: when an OriginSend
banker spends, the address it spends from must be the address the current
message's envelope was actually credited to.

We also added a fail-closed check for any banker that somehow arrives
without the handle set.

## Alternatives considered

**Only add the second check (spender must be the payee).** This was tried
first and it does not work. It does not stop a stored banker, it only puts
it to sleep. The banker wakes up at full value in any later message where
its own realm happens to be the one being called. We built a working
attack against it: a hostile add-on stored a banker from a vault, then
drained the vault every time an innocent user deposited. Keeping this
check as a second layer is still worthwhile, so we did.

**Give each message an identity token and store it in the banker.** This
works but it fails in the wrong direction. Any place that builds an
execution context and forgets to set the token would let everything
through. There are nine such places outside tests, five of them in the
keeper alone. The handle approach cannot be forgotten, because there is
nothing to set.

**Stop storing bankers entirely.** This would break a documented and
tested feature that other realms rely on for the other banker kinds.

**Leave it and document the hazard.** The whole point of this banker kind
is to be the safe one to hand out. A documented warning does not restore
that.

## Consequences

- The app hash changes. The banker's Gno source is part of genesis state,
  so editing it moves the hash. Expected and intentional.
- The banker value gains a field. On a chain that already has stored
  bankers this would need migration. Not an issue for a new chain.
- Test setup had to change. When a test says "pretend a user is calling",
  the harness was also saying "pretend that user received the envelope".
  A user address can never equal a realm address, so this pointed the
  payee at something no realm could match. Now only a *realm* override
  moves the payee.
- Three `banktest` filetests needed one line each, naming which realm the
  coins were sent to. These are tests that call a realm from a wrapper, so
  the payee is not obvious from context.
- `IsUserCall()` guards in realms are still needed and should not be
  removed. This change does not replace them.

## What we checked

- `gno test ./examples/...` — 221 packages pass, none fail.
- The full integration suite passes.
- The VM, test-harness, and file-test suites pass.
- The attack test fails at the moment the banker is handed over, before
  any coins move.
- Storing the other banker kinds still works, and delegating an
  OriginSend banker within one message still works.
