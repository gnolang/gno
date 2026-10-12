# Sign a rendering with oversized strings replaced by their digests, so a Ledger can sign a large transaction

## Context

[pr6173](./pr6173_ledger_compatible_sign_doc.md) made the signature payload's
fee parse on the Ledger Cosmos app, which gno borrows for want of an app of its
own. It closed with an explicit exception:

> `addpkg` still fails, with `hidapi: unknown failure` at 8,840 bytes where a
> ~400-byte call succeeds. That is a transport or device-memory limit on large
> payloads and is a **separate, unfixed problem**: package upload appears to be
> beyond a Ledger regardless of this change. This ADR does not address it.

This is that problem. A realm deployment cannot be signed on a Ledger at all,
so the only keys that can deploy are keys held in software.

### The device will only sign what it parsed

There is no way to hand the device 32 bytes and get a signature over them. The
Cosmos app implements three APDU instructions and no more
(`app/src/apdu_handler.c`):

| instruction | what it does |
|---|---|
| `INS_GET_VERSION` | version |
| `INS_GET_ADDR_SECP256K1` | pubkey and bech32 address |
| `INS_SIGN_SECP256K1` | parses an amino sign doc, displays it, signs it |

`SignSECP256K1`'s mode byte selects `SIGN_MODE_LEGACY_AMINO_JSON` (0) or
`SIGN_MODE_TEXTUAL` (1) (`ledger-cosmos-go@v1.0.0/user_app.go:50`). Neither
skips the parse. So "hash signing" in the literal sense, the device showing a
hash and signing it, is not a mode we can ask for. It is not that blind signing
was judged undesirable; it is not on offer.

What the device signs is therefore always the whole rendering, and the
rendering has to fit:

| ceiling | value | source |
|---|---|---|
| payload bytes | `FLASH_BUFFER_SIZE` 16384, staged through `RAM_BUFFER_SIZE` 8192, or 7168 on Nano X | `app/src/common/tx.c` |
| JSON values | `MAX_NUMBER_OF_TOKENS` 768, 600 on Stax/Flex/Apex | `app/src/json/json_parser.h` |
| refusal | `APDU_CODE_TRANSACTION_DATA_EXCEEDS_BUFFER_CAPACITY` 0x6988 | `app/src/coin.h:90` |

A two-file realm with 20 KB of source renders to a 20,452-byte payload. In
practice the device gives up earlier than 16 KB: the measurement in pr6173 is
8,840 bytes failing from the transport, with nothing in the error about size.

### The one degree of freedom

`tx_validate.c` applies its closed-world key allowlist at the sign doc root and
inside `fee`, and **nowhere else** (`validate_allowed_keys`, called twice). The
contents of `msgs` are unconstrained and displayed generically. Shrinking what
is under `msgs` is the only lever available without shipping an app.

## Decision

A third payload rendering, alongside the two pr6173 established. One rule:

> In the signature payload, any string value longer than
> `std.MaxSignDocStringLen` (512) is replaced by the object
> `{"$sha256":"<hex>"}`, hashing the decoded bytes.

```
sign(digestStrings(sortJSON(aminoJSON(signDocPayload(SignDoc)))))
```

`GetSignaturePayloadDigest` produces it and returns the substitutions it made
alongside the bytes, because every caller that wants one wants the other.
`VerifySignaturePayload` accepts it as `PayloadRenderingDigest`. **Nothing new
goes on the wire**: the node holds the full `[]Msg` and recomputes the
rendering, exactly as it already recomputes the legacy one.

The fallback chain itself moved into `VerifySignaturePayloadFrom`, which takes
the payload the caller has already built. pr6173 had to spell the chain out a
second time inside the ante handler, because the payload there is built before
gas is charged and outside the simulate gate, and a helper that built its own
would have reordered both. Passing the payload in removes that reason, so a
third rendering is one arm in one place rather than a third nesting level in
two. The ante handler keeps building the payload exactly where it did.

A 20 KB deployment becomes 515 bytes:

```json
{"account_number":"42","chain_id":"dev","fee":{"amount":[{"amount":"1000000",
"denom":"ugnot"}],"gas":"200000"},"memo":"deploy foo","msgs":[{"@type":"/vm.m_addpkg",
"creator":"g1h34lmpywh4upnjdg90cjf4j70aee6z8qpqxv3c","max_deposit":"","package":
{"files":[{"body":"module = \"gno.land/r/albttx/foo\"\ngno = \"0.9\"\n",
"name":"gnomod.toml"},{"body":{"$sha256":"48ed9b0f…"},"name":"foo.gno"}],
"name":"foo","path":"gno.land/r/albttx/foo"},"send":""}],"sequence":"7"}
```

### Why strings, and not the whole document

The obvious reading of "hash signing" is to replace the whole of `msgs` with one
digest. It is less code and it would always fit. It also throws away everything
the device could still have shown.

Above, what survives is the chain, the account, the sequence, the fee, the
creator, the package path, the name of every file, and the whole of
`gnomod.toml`, which is short enough to pass through untouched. One value is
opaque. A signer still confirms **where** the code goes and **who** deploys it
on trusted hardware, and takes only the code text on faith.

512 is chosen so that no realistic transfer is ever digested: a bech32 address
is 40 characters and a coin string a handful more, so `bank.MsgSend` renders
identically in both modes and the money path keeps full on-device display. That
is a claim about realistic transfers, not an invariant of the type. `Coins`
render as one string, so enough denominations in one send cross any threshold.
`TestDigestNeverAppliesToMsgSend` pins where it holds.

### Why a third rendering is safe

The malleability question pr6173 answered for two renderings has to be answered
again for three: if a digest rendering of one transaction could equal a full
rendering of a **different** one, a signature authorising T1 would also
authorise T2.

It cannot. The digest rendering differs from the current one only where the
current one holds a string over 512 bytes and the digest holds an **object**.
For the digest rendering of A to equal a full rendering of B, B would need an
object where A has a string, keyed `$sha256`. Amino's object keys come from Go
struct tags; none begins with `$`, so no rendering of a real message can contain
that key, not even when an attacker writes the key into a field they control,
because a string stays a string. Equality forces A and B to be the same
transaction, barring a SHA-256 second preimage.

**The object shape is load-bearing.** Had the substitution produced a string,
say `"sha256:48ed…"`, then a transaction whose field legitimately held that
71-character literal would render byte-identically to the digest rendering of a
different transaction: one signature, two transactions. This was the first
design and it is wrong.

`TestDigestKeyIsUnreachableFromAmino` and `TestDigestRenderingIsDisjoint` pin
both halves, and they are the tests to consult before `MaxSignDocStringLen` or
`digestKey` is touched.

### Cost on the verification path

None of this is metered, as pr6173 set out: a transaction the ante handler
rejects pays no fee, so the cost of rejecting an invalid signature is borne by
the node. A third rendering therefore has to be accounted for honestly rather
than waved at, and the first draft of this change got it wrong: it leaned on the
fact that the third *curve operation* is skipped while paying for the third
*rendering* unconditionally, which measured at roughly twice the saving.

Three things bring it back down. The numbers are from
`tm2/pkg/std/doc_digest_bench_test.go`, which exists so that these claims can be
re-run rather than believed, on two ~1MB payloads: **digestable**, twenty 50KB
bodies all replaced, and **identity**, 7,650 hundred-byte bodies none of which
can be, the cheapest 1MB a node can be made to reject. Each comparison below is
between two measurements of the *same* fixture. For scale, a secp256k1
verification over 1MB is 2.44ms on the same machine.

**The rendering is derived, not rebuilt.** `VerifySignaturePayloadFrom` takes
the payload the caller already has and transforms it (`TransformInHand`, 9.20ms,
2.0MB). Rebuilding it from the transaction first, which the ante handler did
while the chain was written out twice, adds a `FullRender` on top: 8.76ms and
5.1MB per rejected signature, about 3.6 verifications of the same transaction
spent to avoid one.

**A payload that cannot hold an oversized string is never parsed.** The digest
rendering can differ from its input only if some JSON string literal exceeds the
threshold, and that is decidable in one allocation-free pass. On the identity
fixture `IdentityScan` is 1.34ms and zero allocations where `IdentityWalk`,
which is what the first draft paid, is 17.4ms, 11.2MB and 130,164 allocations:
thirteen times the time and all of the garbage, for an answer known in advance.
The scan measures *escaped* lengths, never shorter than decoded ones, so it
cannot miss a string that would be digested; false positives cost only the walk
that would otherwise have run anyway.

**The report is not built for a reader who does not exist.** The field paths and
the key ordering behind them exist for the client: `ReportedWalk` against
`IdentityWalk` is 20.8ms against 17.4ms and 198,933 allocations against 130,164,
so a fifth of the walk's time and over half its garbage would be produced and
discarded. The consensus path asks for no report and the walk skips both.

What remains: ordinary traffic pays one encoding and one curve operation,
exactly as before, because the scan answers no and the digest arm never runs. A
bad signature on an ordinary 1MB transaction pays the two renderings it already
paid plus 1.34ms. A bad signature on a transaction that really does carry
oversized fields pays one transform of a payload already in hand. The identity
case costs no byte comparison either: when nothing is substituted the transform
returns *its input*, so the skip is a pointer-level fact rather than a 1MB
`bytes.Equal`.

### The client half, which is not optional

Comparing the digest against the device screen is the entire security of the
mode. `gnokey sign` and `gnokey maketx` take `-sign-mode full|digest`,
defaulting to `full`, and in digest mode print every substitution before the
device prompts:

```
sign mode digest: 1 field(s) will reach the device as a digest only.
Compare each one against the device screen before approving.

  msgs[0].package.files[1].body (19812 bytes)
    48ed9b0f c9e890e6 2dafb1da 65d24046 4c0bcc3c 3de0c9b4 05dcad87 a4135e8d
```

Grouped in eights because a person comparing 64 hex characters against a
paginated device screen otherwise cannot. It also says so explicitly when it
digests nothing, rather than leaving a signer to wonder what the silence
covered.

**The mode is never selected automatically.** A silent fallback would change the
signer's display guarantee without telling them, which is the one thing this
change must not do. A misspelled mode is an error, not a fallback to `full`.

### What the signer actually gives up

Honestly stated, because the commit message cannot carry it: this is not
equivalent to full on-device display. The device authenticates the envelope:
chain, account, sequence, fee, memo, message type, creator, package path,
function, file names. It also commits to the bodies. What a human cannot check on
the device is the body text. A compromised host that swaps a file body is caught
only by someone comparing the hex against a second machine.

It is strictly better than blind signing, strictly worse than full display, and
only as good as the comparison the signer performs.

### Where the device limits are now named

`ledger.MaxPayloadSize` (16384) is checked before the device is asked anything,
since a payload it cannot hold does not become signable by validating the key
first. `explainSignError` classifies the three refusals that mean "too large to
display" and arrive saying nothing about size: 0x6988, which has no entry in
zondax/ledger-go's table; token exhaustion, a bare string from the app's JSON
parser; and a transport failure on a payload past `largePayloadSize` (8192),
which is the failure pr6173 measured. String matching is the only option either
library leaves, and the device's own words are always wrapped, because on a large
payload size is the likely cause and not a certainty.

**The device layer names no flag.** It reports what happened as
`ErrPayloadTooLarge`, `ErrTooManyJSONValues` and `ErrLargePayloadRefused`, and
each front end phrases the remedy it can actually offer: `gnokey` suggests
`-sign-mode digest`, and only when the signer is not already using it. An earlier
draft put that sentence in the device package, which reaches `gnoclient` through
`keys.Keybase` where no such flag exists, making it advice the caller could not
act on.

## Consequences

**This is a consensus change and needs a coordinated upgrade.** An upgraded node
accepts a superset of what an un-upgraded one does, so the failure mode is the
same shape as pr6173's: an upgraded proposer can include a digest-signed
transaction in a block that un-upgraded validators reject. Nodes before clients.
Nothing that works today stops working.

**`MaxSignDocStringLen` is a consensus constant.** Moving it changes the signed
bytes of every transaction holding a string near the boundary, and a node that
moved it would reject signatures every other node accepts.

**The mode bounds per-field size, not total size.** A hundred files of 500 bytes
each are all under the threshold, so the payload is still 50 KB and the device
still refuses. The token ceiling is the second wall: a digested file costs 7 JSON
values against a plain file's 5, which puts the limit around 105 files, or 81 on
the 600-token devices. Any realistic realm fits; a monorepo-sized deployment does
not, and gets an error naming the ceiling rather than a sliding threshold the
verifier would have to re-derive.

**JSON escaping inflates what the device sees.** The threshold counts decoded
bytes, but `\n` costs two and `\uXXXX` six, so a 512-byte body of newlines
reaches the device as 1 KB. Far from 16 KB, but the threshold is not a tight
bound on the payload.

**Multisig members may mix modes.** Verification is per-signature, so a Ledger
co-signer can sign the digest rendering while a software co-signer signs the full
one. This is a strict improvement on pr6173's note that members must agree on a
rendering: for this rendering they need not.

**Clients that build the payload themselves are unaffected** unless they want the
mode. `tm2-js-client`, Adena and gnonative keep working; a wallet that wants to
sign a large transaction on a Ledger implements `digestStrings` and nothing else.

**Nothing retires the mode, and nothing should.** Unlike pr6173's legacy
rendering, this is not a transition shim: it is the only way a device-held key
can authorise a large transaction, and it stays until gno ships an app whose
display is good enough to make it unnecessary, which, see below, it would not be.

### Tests

In `tm2/pkg/std`, against the rendering itself:

- the digest rendering is byte-identical to the current one for every document
  with no oversized string, and reports no substitutions, the property both
  verification sites' short-circuit relies on
- the boundary is exclusive and counts decoded bytes
- the digest is SHA-256 of the decoded string, checked with a body full of
  newlines so that hashing the escaped form fails
- the walk reaches a field at any depth, with the path reported as
  `msgs[0].package.files[1].body`, keys still sorted, short siblings untouched
- no full or legacy rendering of any document, including ones whose memo and
  chain ID spell out the digest object, contains an object keyed `$sha256`
- the digest rendering of one document equals no rendering of three documents
  built to impersonate it
- a digest signature verifies, reports itself as `PayloadRenderingDigest`, and
  fails against a document whose digested field was swapped for another of the
  same length
- a full-rendering signature over a document with an oversized field still
  reports as `PayloadRenderingCurrent`, so arm order relabels nothing
- a transfer is never digested, at one and at eight denominations
- the byte scan never misses an oversized string, over a table of strings whose
  escaped form is longer than their decoded one (newlines, quotes, control
  characters, a trailing backslash). A false negative is the one error that
  would make a node reject a signature its peers accept, so this is the test to
  consult before the scan is touched
- a false positive stays harmless: the payload is parsed, nothing is
  substituted, and the rendering comes back as the input itself

Benchmarks in `tm2/pkg/std/doc_digest_bench_test.go` back the cost claims above:
a full render against the in-hand transform, and the scan against the walk it
skips, with and without the client's report.

In `gno.land/pkg/sdk/vm`, against the message this exists for, since `tm2`
cannot import `gno.land`: the digest rendering of a `MsgAddPackage` is pinned
byte for byte, the full rendering is over `ledger.MaxPayloadSize` while the
digest rendering is under it, only the file body is digested, and a small package
renders identically in both modes.

Behavioural: the ante handler admits a digest-signed transaction and rejects one
whose digested content was swapped; `gnokey` refuses an unknown mode, defaults to
full, signs the digest rendering when asked, and prints the field, its size and
its grouped digest; `gnokey verify` reports a digest-rendering signature and
lists the fields nobody read. On the device layer, `explainSignError` classifies
the three size failures, leaves a rejected prompt or an unopened app alone, names
no client flag, and quotes no ceiling it cannot be under; the pre-check refuses an
oversized payload without asking the device anything. On the client, the remedy
is attached for each classification and withheld from a signer already in digest
mode. A txtar integration test deploys a package too large for a device against a
real node in digest mode, including through the simulation path, which signs the
transaction a second time with the gas raised.

## Alternatives considered

**Replace the whole of `msgs` with one digest.** What "hash signing" usually
means, and what a signer asking for it usually pictures: one hash on the screen,
one comparison. Less code, one substitution instead of a tree walk, and it
would never hit a size ceiling again. Rejected because the device then shows
nothing about the transaction except the fee and the sequence: not the package
path, not the creator, not an `m_call`'s send amount. Everything is still
committed to by the hash, so nothing can be swapped silently, but the only thing
a human verifies on trusted hardware is that 64 characters match a screen on the
machine they do not trust. The per-field rule already gives exactly that UX for
a single-file deployment, and keeps the envelope. Worth adding as a second,
explicitly named mode if a transaction ever cannot fit even digested; it is a
dozen lines from here, and it should be the signer's choice rather than a silent
fallback.

**An explicit mode byte on `std.Signature`.** `SessionAddr` shows the shape:
amino skips the zero value, so it costs nothing on the wire for a transaction
that does not use it. It would let the node verify exactly one rendering instead
of trying three. Rejected for now because it is a wire change every client must
adopt, and the byte-equality short-circuit already keeps the extra cost off every
transaction without an oversized field. It is the right answer if a fourth
rendering ever appears.

**A sliding threshold.** Lower `MaxSignDocStringLen` through a fixed ladder until
the payload fits, so a many-small-files package still signs. Deterministic from
the transaction, so the verifier could re-derive it, but it makes the signed bytes
depend on a search rather than on a constant, and it buys a case nobody has hit.
Rejected under the rule that a threshold a verifier has to re-derive is worse
than an error naming the ceiling.

**Ship a gno Ledger app.** Correct in the long run and the only route to a device
that displays gno's own message types meaningfully, as pr6173 already noted. It
does **not** solve this: the buffer and token ceilings are the platform's, not
the Cosmos app's, so a 20 KB deployment would need the same digest rendering
behind a gno app. The work here is not wasted either way.

**Use a session key instead.** `auth.MsgCreateSession` is already in the tree
with an expiry, a path allowlist and a spend limit. The Ledger signs one small,
fully displayable `create_session` transaction, and a hot key signs the
deployment; the device never sees the large payload and no rendering with a
reduced display guarantee enters the protocol. For the everyday case, a
maintainer deploying from a laptop, this is the better answer, and it needs no
consensus change at all. It is not a substitute when the master key itself must
authorise the transaction, which is what this ADR is for.

**Do nothing.** A Ledger-held key cannot deploy a realm. Viable only if that is
an acceptable permanent restriction, and the existence of
`-sign-mode digest` does not force anyone to use it.
