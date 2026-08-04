# GnoConnect: Wallet & Client Integration Standard

GnoConnect is a standard for enabling wallets, clients, and SDKs (such as Adena
Wallet, Gnoweb, and Gnobro) to interact seamlessly with Gno blockchains. It's a
minimalistic, URL-based alternative to the gno-js-client that allows users to
define actions in their apps without JS/TS components, making integration
straightforward for both users and developers.

## How GnoConnect Works

GnoConnect uses HTML/HTTP metadata to provide connection details for clients and
wallets.

By including the following metadata/headers in your app, clients and wallets will be able to recognize your app as Gno-compatible and get the data needed to generate transactions for users.

### HTML Metadata

```html
<meta name="gnoconnect:rpc" content="127.0.0.1:26657" />
<meta name="gnoconnect:chainid" content="dev" />
<meta name="gnoconnect:txdomains" content="auto,example.com" />
```

- `gnoconnect:rpc`: RPC URL.
- `gnoconnect:chainid`: Chain ID.
- `gnoconnect:txdomains`: Domains treated as transaction sources.
  The value `auto` includes the current domain in addition to any specified
  domains.

### HTTP Headers

Alternative to HTML Metadata, for a client that **fetches** the page rather than
runs inside it — a CLI resolving a TxLink URL, or an agent asking "is this
Gno-compatible, and on which chain" without parsing HTML.

```
Gnoconnect-RPC: 127.0.0.1:26657
Gnoconnect-ChainID: dev
Gnoconnect-TXDomains: auto,example.com
```

A client uses whichever source it can read. A client that reads both and finds
them in conflict prefers the header. This is a tiebreak, not a security boundary:
**a client that runs inside the page may have no access to response headers at
all**, so a client that reads only the metadata is conforming, and a producer
MUST NOT rely on a header to override a `<meta>` that contradicts it.

### Who `rpc` is for

`rpc` means different things to the two kinds of client, and conflating them is
the mistake this section exists to prevent.

- **A client with no networks of its own** — a CLI resolving a TxLink, an agent,
  an indexer — has no other endpoint. For it, `rpc` *is* the endpoint. That is
  what this metadata channel is for.
- **A wallet** holds the user's keys and the user's networks. For it, `rpc` is
  advisory and `chainid` is what selects. A wallet MUST NOT query or broadcast
  through a producer-supplied endpoint; see Network resolution.

The same distinction applies to `rpc` wherever a request carries it — a launch
link parameter or an in-page intent field. It is a declaration of what the
producer expects, never an instruction to the wallet.

## Network resolution

Every request that names a chain resolves it the same way, on either transport,
before anything is signed or disclosed.

A **configured network** is a chain id, one or more endpoints, and the endpoint
the user has selected for it. The **active network** is the configured network
currently in use. A wallet queries and broadcasts only through the selected
endpoint of a configured network.

**1. Determine the chain.** From the request's `chainid`; for an in-page request,
falling back to the page's `gnoconnect:chainid`.

Whether naming no chain is an error depends on what the request does:

- **A request that signs — `sendtx`, `signtx`, and their in-page equivalents —
  MUST name one**, or the wallet answers `invalid_request`. A signature is
  chain-bound, so a producer always knows which chain it wants, and "whichever
  the wallet happens to be on" is how a dapp built for one chain gets a signature
  valid on another.
- **A request that only discloses — `connect` — MAY omit it.** Nothing is signed,
  so there is no chain-bound artefact to get wrong, and "whichever chain you are
  on" is a complete and honest answer to "who are you": the response carries the
  `chainid` it was answered against, so the producer is never left guessing. A
  wallet MUST NOT answer `invalid_request` merely because a `connect` named no
  chain.

When a `connect` does name one, it resolves like any other request and MAY
therefore prompt the user to switch, or to add a chain the wallet does not have.

**2. Find it among the configured networks.** Its selected endpoint is the one
the wallet will use. The request's `rpc` plays no part: the user has already
chosen how they reach this chain.

**3. If the chain is not configured**, the wallet MAY offer to add it, prefilling
the endpoint from `rpc`. That offer MUST be an approval of its own — separate
from, and before, any signing approval — and MUST show the endpoint. A
**request-initiated** add MUST NOT create an endpoint for a chain that is already
configured: if the chain is known, the user has made this choice and the request
has nothing to propose. (A user adding a second endpoint themselves, for a node
their ISP blocks or one that is temporarily unreachable, is unconstrained — this
rule is about who initiates, not about what results.) A user who declines, or a
wallet that does not implement adding, answers `network_declined`.

**4. Switch if needed.** If the resolved network is not the active one, the
wallet asks the user; declining answers `network_declined`. Every query — `vm/qdoc`,
account number, sequence, gas — and the broadcast then use that network's
selected endpoint.

The review screen MUST show **the chain id and the endpoint in effect**, and
SHOULD show the network's name where the wallet has one. Those two identify the
network; a name is a label, and not every wallet has one to show. A wallet may
model a network as nothing more than a chain id and an endpoint — which is what
the resolution above actually needs — and a chain added from a request under step
3 arrives with no name at all, since a request has no field to propose one. A
MUST that some wallets structurally cannot meet, on the part that identifies
nothing, would only teach implementers to invent a placeholder.

What this guarantees, stated exactly, because it is narrower than "the producer's
value is never used":

- A request can never replace, shadow, or add to the endpoints of a chain the
  user has already configured.
- A user is never asked to approve an endpoint and a signature in the same
  interaction.
- For any chain the user had before the request arrived, `vm/qdoc`, sequence, gas
  and broadcast all go through the user's own choice.

A producer-supplied endpoint *can* become a configured one — that is what step 3
is — but only for a chain the wallet did not have, and only through an approval
that does not sign. The uncovered case is first contact with a genuinely new
chain, where the user has approved a node that then answers the `vm/qdoc` lookup
shaping the call and its labels. There, a wallet SHOULD mark the endpoint as
newly added on the review screen, and SHOULD prefer the positional argument form,
whose binding does not depend on the node.

### A declared `rpc` the wallet is not using

A chain id is **not globally unique**. `dev` names every local devnet, and a
reset or forked testnet reuses its id. So selecting on `chainid` alone can match
a *different network* than the producer meant, and nothing about the resulting
transaction looks wrong: it is signed for chain `dev`, and it is valid on chain
`dev`, just not the one the producer had in mind.

When a request declares an `rpc` that is not the endpoint the wallet uses for the
selected chain, the wallet SHOULD show both in the review. That divergence is the
only signal available that `chainid` may have selected the wrong network —
discarding it silently throws the signal away. The wallet MUST NOT adopt the
declared endpoint on that basis; this adds information to a screen the user is
already reading, and nothing else.

SHOULD rather than MUST because the divergence is usually benign: where a chain
id *is* unique, two endpoints are simply two nodes on one chain, the transaction
is chain-bound and lands either way, and a user running their own node for a
public chain would otherwise be warned on every transaction — a false positive
that would teach them to ignore the warning that matters.

> **Known limitation.** Naming a network with a string that is not unique is the
> underlying defect, and this is a mitigation, not a fix. Identifying a network by
> something derived from it — a genesis hash, a fingerprint — would make selection
> unambiguous and retire the problem. That is a new field and a separate design
> discussion; v1 does not attempt it.

## Transaction Links (TxLinks)

Transaction links define blockchain calls and can include optional arguments.

Without arguments:

```
$help&func=Foo
/r/path/to/realm$help&func=Foo
https://example.land/r/path/to/realm$help&func=Foo
```

With arguments:

```
$help&func=Foo&arg1=value1&arg2=value2
/r/path/to/realm$help&func=Foo&arg1=value1&arg2=value2
https://example.land/r/path/to/realm$help&func=Foo&arg1=value1&arg2=value2
```

Here `arg1`/`arg2` stand for the function's actual parameter names — a TxLink
names each argument directly. Launch links carry the same arguments namespaced
under `arg.<name>` (see Launch Links), because a launch link also has reserved
keys like `func` that a bare parameter name could otherwise collide with.

Links can be relative or absolute but must match one of the domains listed in
`gnoconnect:txdomains` (including the resolved `auto` domain if set). **When
`gnoconnect:txdomains` is absent, a receiver treats only the page's own origin as
a transaction source.** Same-origin is the safe default: it is what a page
without an explicit list can be assumed to have meant, and it never widens the
set silently.

TxLinks only prefill specified arguments. For non-specified arguments, clients
can call `vm/qdoc` to retrieve the remaining fields
(see [PR #3459](https://github.com/gnolang/gno/pull/3459)).

> **Note:** A future standard may define advanced rules for fields such as
> limits, format, and default values.

## Arguments: named or positional

A `MsgCall` takes its arguments **positionally**, in the realm's declaration
order, and every one of them is a string (`MsgCall.Args` is `[]string`) — so
order carries the entire meaning and no type check will catch a wrong one. There
are two ways to supply them, and the choice belongs to the whole call.

**Named** — `arg.<name>=value`, or `{ name, value }` in-page. The producer states
which parameter each value belongs to and says nothing about order.

- The wallet MUST resolve declaration order from the realm's signature via
  `vm/qdoc`, against the network resolved above.
- Inter-realm parameters (`cur realm`) are supplied by the VM, not the caller,
  and MUST NOT consume a positional slot.
- **A failed lookup is an error, not a fallback.** A wallet MUST NOT fall back to
  the order the arguments happened to arrive in. That order is incidental — it is
  whatever order the producer's code appended query parameters or iterated a map
  — so binding to it invents an assertion the producer never made. Nothing
  downstream catches the result: every argument is a string, so a permuted call
  is type-valid, the chain executes it, and a review screen with no `vm/qdoc`
  document has no parameter names to show the user either.

**Positional** — repeated `args=value`, or `{ value }` with no name in-page. The
producer asserts the order deliberately and takes responsibility for it.

- No lookup is required, so this is the form that works without network access.
- A wallet MAY still perform the `vm/qdoc` lookup to label the review screen, but
  MUST NOT reorder the arguments, and a lookup failure MUST NOT prevent signing.

**One form per call.** A request carrying both is answered `invalid_request`.
Mixing has no coherent meaning: a named argument among positional ones has a
position knowable only through `vm/qdoc`, at which point the call needs the
network anyway and the positional form has bought nothing.

**No arguments at all** — `args` absent, empty, or no `args=`/`arg.<name>=` on a
link — is **neither form**, and the rules above do not apply to it. The forms
exist to answer one question, "in what order do these values go", and with no
values there is no order to settle. So:

- The mixing rule is vacuous, and a wallet MUST NOT answer `invalid_request`
  merely because a request supplied no arguments.
- The wallet SHOULD still resolve the realm document, to label the review screen
  and to show the user any declared parameter left unsupplied — that is what
  makes a partly-filled call reviewable.
- **A failed lookup MUST NOT prevent signing**, exactly as for the positional
  form. Nothing about the binding depends on it: there is nothing to order and
  nothing to misplace. Requiring it would put the one call shape that needs no
  network information — a function that declares no parameters, `Increment()` and
  its kind — behind a network round trip, and so out of reach of offline signing,
  which is the opposite of what the forms are for.

The wallet cannot tell a complete zero-argument call from an unfilled one without
the lookup, and that is the honest position to present: with the document, it
shows the parameters and what is missing; without it, it shows a call supplying
no arguments and lets the user judge. Neither is a reason to refuse outright.

**A name that matches no declared parameter** MUST NOT be bound positionally and
MUST NOT be silently ignored. The wallet answers `invalid_request`, or surfaces
the argument to the user as unmapped for explicit confirmation. Dropping it
quietly means signing something other than what was asked, with nothing on screen
to say so.

**Signing offline.** A wallet MAY sign without network access whenever it can
obtain the chain id, account number, sequence and gas by other means — asked of
the user, or cached — exactly as `gnokey sign` takes `--chainid`,
`--account-number` and `--account-sequence` rather than querying. Those values
belong to the signer, not to the producer, so this standard does not carry them:
a producer-supplied sequence would be one more value the dapp chooses that shapes
what gets signed. The positional form is what makes offline signing reachable,
since named arguments require the `vm/qdoc` lookup.

## The `signer` pin

A request MAY pin the identity it expects to act as: `signer` in a launch link,
`signer` on an in-page intent. Both transports carry the same field and it means
the same thing in both, which is why it is defined here rather than twice.

**What it names.** The `address` a prior `connect` returned — an *identity*, not
a key. A producer MUST pin that address. It MUST NOT pin a delegated address
(the `session` a `connect` callback may also carry): a delegated key is
ephemeral, so a pin taken at connect time goes stale on rotation, and it names
key material rather than the party the producer means.

Pinning therefore presupposes a way to learn the address: `connect`, or in-page
`getAccount`. A wallet that implements neither cannot be pinned against, and a
producer holding no address simply omits `signer` — it is optional, and a request
without it is answered by whichever account the user approves.

**What the wallet owes.** If `signer` is present the wallet MUST produce a
transaction authorised by that identity, and MUST NOT substitute another. A
wallet that holds no account resolving to it MUST decline —
`signer_unavailable`, and for a launch link
`status=error&code=signer_unavailable` — rather than sign as whoever is to hand.

**What it does not constrain.** Which key signs. A wallet MAY sign with a
delegated or session key whose on-chain authority derives from the pinned
identity; that transaction is still authorised by the identity, which is what the
producer pinned. Key selection is the wallet's business, and a wallet that
rotates session keys could not honour a pin at all if it were not.

This is also what the chain records. A delegated key acts on the identity's
behalf: the message's caller and the state changes are credited to the identity,
not to the key that signed. So pinning the identity pins what the transaction
will be attributed to on chain, which is the thing a producer actually means —
pinning a key would pin something the chain does not attribute the action to.

So what a producer may rely on, stated exactly:

- The transaction is authorised by the identity it pinned, and the chain will
  attribute it there; or there is no transaction and the wallet said
  `signer_unavailable`.
- Nothing about *which key* produced the signature. A producer needing that must
  read it from the chain; the pin does not carry it.

For `sendtx` a producer that did not pin can still identify the transaction
afterwards, since the wallet returns its `hash`. For `signtx` it cannot: the
signed bytes are opaque and there is no hash until the producer broadcasts. See
the `signtx` obligations.

A wallet MAY additionally accept an address that on-chain resolves to the pinned
identity — its own delegated key, say — as naming that identity. This is
leniency about input, not a second meaning: the guarantee above is unchanged
either way, which is why the latitude is harmless where a wallet-specific
definition of the identity itself would not be.

**Where it is enforced.** At the point the signature is produced. A pin checked
only when the request arrives is not enforced: on both transports the user may
change the selected account between arrival and approval, and the guarantee is a
property of the signature, not of the request. Checking early as well is useful —
it refuses a hopeless request before spending the user's attention on it — but it
does not discharge the obligation.

`signer_unavailable` means the wallet *cannot* be the pinned identity, not that
it cannot right now: a held identity whose session has expired is a state the
user can fix, and a wallet SHOULD offer that rather than declining. Declining
tells the producer to re-`connect`, which is the wrong advice when the wallet
holds the identity all along.

## In-Page Wallets (browser extensions)

A wallet that runs code in the page announces itself; the page collects the
announcements and lets the user choose. This replaces the namespace race that
a single `window.<wallet>` global creates — with one global, the last
extension to load wins and the others become invisible, so a user with two
wallets installed cannot reach one of them.

The handshake is two events on `window`, matching EIP-6963 and the Wallet
Standard:

| Event | Direction | Payload |
|---|---|---|
| `gno:registerWallet` | wallet → page | `CustomEvent` whose `detail` is `{ info, provider }` |
| `gno:requestWallet` | page → wallets | none |

Neither side can assume it loaded first, which is the whole difficulty:

- A wallet MUST announce when it loads **and** on every `gno:requestWallet`.
  A wallet that announces only once is invisible to any page that started
  listening afterwards.
- A page MUST start listening before it asks, and MUST keep listening after —
  wallets load asynchronously, so the first answer is never known to be the
  last. A page that reads its list once, at load, will miss wallets.

A wallet MUST NOT cancel or consume the page's own events. Scraping a page's
markup and cancelling the submit or click that produced it makes the choice on
the user's behalf, invisibly to the page, and binds the wallet to one site's
DOM. Announcing is how a wallet becomes reachable; being called is how it acts.
This is the second thing the announce protocol replaces, alongside the
`window.<wallet>` global — with either one, a user who installed two wallets
reaches whichever got to the event first, and the page cannot offer the
choice.

Observing is not intercepting: a wallet may listen to events a page dispatches,
as long as it does not cancel them, stop their propagation, or act on them as
though it had been called.

```ts
// Wallet side
const announce = () =>
  window.dispatchEvent(
    new CustomEvent("gno:registerWallet", {
      detail: Object.freeze({ info, provider }),
    }),
  );
window.addEventListener("gno:requestWallet", announce);
announce();

// Page side
window.addEventListener("gno:registerWallet", (e) => wallets.add(e.detail));
window.dispatchEvent(new Event("gno:requestWallet"));
```

### `info` — how the wallet is presented

```ts
interface GnoWalletInfo {
  uuid: string; // announcement handle, unique per page load (UUIDv4)
  name: string; // human-readable, shown to the user
  icon: string; // data:image/ URI — no network fetch, works offline
  rdns: string; // durable identity, reverse-DNS (e.g. "land.gno.gnokey")
}
```

`uuid` deduplicates repeated announcements within a page; `rdns` is what
survives across page loads and versions, so it — not the display name — is
what a page should persist when remembering a choice.

### `provider` — what the page calls

The provider carries the wallet's methods. `sendTx` carries the same
**transaction intent** as the `sendtx` launch link — the same verb, cased for the
medium (a URL host is case-insensitive per RFC 3986, so it cannot be camelCase;
a JavaScript method conventionally is). What a launch link adds is the
**envelope**: out-of-band delivery (`callback`, `state`) that a direct call,
which simply returns a `Promise`, does not need. A direct call is also not
URL-bounded, so it carries large arguments a launch link cannot (see Payload
size).

```ts
type GnoArg =
  | { name: string; value: string }   // named — resolved via vm/qdoc
  | { value: string };                // positional — order is the producer's

// One message. Everything here varies per MsgCall.
interface GnoMsgIntent {
  path: string;    // full package path
  func: string;    // exported function name
  args: GnoArg[];  // one form per call; mixing is invalid_request
  send?: string;   // coins, gnokey syntax
}

// One transaction carrying one message. The fields below the message belong to
// the transaction, not to the call — see sendMsgs for why that distinction is
// in the types rather than in prose.
interface GnoTxIntent extends GnoMsgIntent {
  chainid?: string; // falls back to gnoconnect:chainid
  rpc?: string;    // advisory only — see Network resolution
  signer?: string; // bech32 identity — see The `signer` pin
}

type UserResponse<T> =
  | { status: "Approved"; args: T }
  | { status: "Rejected"; code?: ErrorCode };
```

`ErrorCode` is the launch links' enumerated `code` set (see `sendtx` callback
results). A rejection carries it rather than an untyped error, so a page handles
failures identically whether it called the wallet directly or handed off a launch
link:

```ts
type ErrorCode =
  | "invalid_request"
  | "network_declined"
  | "signer_unavailable"
  | "no_signer"
  | "not_connected"    // in-page only — see Connecting, and what it gates
  | "unsupported_host"
  | "tx_failed";
```

One set serves both transports; `not_connected` is the one code a launch-link
producer never sees, because a launch link carries its own consent and there is
no connection to be outside of.

`sendTx` is the core method: one call, signed and broadcast, returning the
`hash`.

```ts
sendTx(tx: GnoTxIntent): Promise<UserResponse<{ hash: string }>>;
```

A user declining is `Rejected`, not a thrown error: refusing to sign is an
answer. Only a genuine failure rejects the promise, and it rejects with the same
enumerated `code` a launch link would have returned (see `sendtx` callback results),
so a page has one error vocabulary whatever transport it used. User review before
signing is mandatory, as for `sendtx`.

`signer`, when present, pins the identity the producer expects to act as — see
The `signer` pin, which governs both transports. Without it a page that connected
as one account and rendered its address will sign as whatever account the user
has since switched to, and only find out from the chain.

#### Optional methods

A wallet MAY implement more of the surface. These are the defined shapes; a page
MUST feature-detect every method it calls rather than assume, and degrade — to
another wallet, a launch link, or the copy-paste command — when it is absent.

```ts
// Sign without broadcasting. The producer broadcasts; see signtx for the
// obligations that moves. `signedtx` is base64 amino-binary, opaque.
signTx(tx: GnoTxIntent): Promise<UserResponse<{ signedtx: string }>>;

// Ask the user which identity to act as. Discloses nothing until they agree.
// `chainid` is optional: given, it resolves like any other request and may
// prompt a switch; omitted, the wallet answers against the active network and
// says which in GnoAccount.chainid. See Network resolution, step 1.
connect(opts?: { chainid?: string }): Promise<UserResponse<GnoAccount>>;

// The connected identity, without re-asking. Answered against the active
// network; it names no chain, so it resolves none.
getAccount(): Promise<UserResponse<GnoAccount>>;

// The active network. Reports what is in effect; it names no chain, so it
// resolves none and never prompts.
getNetwork(): Promise<UserResponse<GnoNetwork>>;

// Ask the user to switch to a configured chain. A chain the user does not have
// is network_declined, not a silent add.
switchNetwork(chainid: string): Promise<UserResponse<{ chainid: string }>>;

// Several messages, ONE transaction: one signature, one broadcast, one hash.
// The launch-link analogue is the multi_msg feature, and the name matches it —
// what you supply is messages. The chain and the signer sit on the transaction,
// not on each message; see below.
sendMsgs(tx: GnoBatchIntent): Promise<UserResponse<{ hash: string }>>;

interface GnoBatchIntent {
  msgs: GnoMsgIntent[];  // at least one; empty is invalid_request
  chainid?: string;      // falls back to gnoconnect:chainid
  rpc?: string;          // advisory only — see Network resolution
  signer?: string;       // bech32 identity — see The `signer` pin
}

interface GnoAccount {
  address: string;         // bech32
  chainid: string;         // the chain this answer was given against
  pubkey: string | null;   // gpub, when the wallet exposes one
}

interface GnoNetwork {
  chainid: string;
  rpc: string;             // the endpoint in effect, not one a page declared
  name?: string;           // the wallet's label, when it has one — not an id
}
```

Announcing is not a claim to implement everything: the same additive
forward-compatibility contract as launch links applies (see Forward
compatibility) — capabilities are only ever added, never repurposed, and a page
degrades on any method it does not recognise.

Message signing is deliberately absent. Everything signable in Gno today is a
transaction — `gnokey sign` takes a tx document and nothing else — so a
`signMessage` would have no defined meaning to agree on. When one exists it
arrives as a new method, not as a re-reading of these.

### One transaction, several messages

`sendMsgs` sends **one transaction carrying several messages** — one signature,
one broadcast, one `hash`. It is named for what a page supplies, because what it
sends is a transaction, singular. Several *transactions* would be a different
method: several signatures, several hashes, and an answer to what happens when
the third fails after the first two have landed. Nothing here offers that.

It takes a **batch intent**, not an array of transaction intents, and the
difference is the whole point: `chainid`, `rpc` and `signer` describe the
transaction, `path`/`func`/`args`/`send` describe a message, and only the second
group can vary within one broadcast. One transaction lands on one chain under
one signature.

An array of `GnoTxIntent` let a page write `[{chainid: "dev"}, {chainid:
"test5"}]` — a request with no meaning at all. Every wallet then had to invent an
answer: reject it, take the first, take the last, take whichever element
happened to carry one. All four were conforming, because the standard said
nothing, and the two behaviours are indistinguishable to the page that sent it.
Hoisting the transaction-level fields is what makes that unwritable, which is
better than a rule against writing it.

This also matches the launch-link multi-message form, where the indexed fields
are `msg.<i>.path` / `msg.<i>.func` / `msg.<i>.arg.<name>` and `chainid`,
`signer`, `callback` and `state` stay top-level. The two transports now agree on
which fields belong to the message and which to the transaction.

`msgs` MUST carry at least one message; an empty batch is `invalid_request`,
since there is nothing to sign and nothing to show the user.

A wallet MAY decline a batch it cannot review honestly — see the review
obligations in Network resolution — rather than present a list the user cannot
follow. Signing several messages the user did not individually understand is
worse than refusing one request.

### Connecting, and what it gates

A page reaches an in-page wallet whenever it likes: the provider is simply there,
and calling a method costs nothing. So a wallet MAY require the user to approve
an **origin** before it answers that origin at all, and most do. Approving is
what `connect` performs, and the approval persists for the origin rather than for
the call — that is the whole difference between this transport and a launch link,
where each link carries its own consent and there is no state between them.

This is a wallet's choice, not a requirement. What the standard fixes is what a
page sees either way.

**A wallet MAY gate any method on an approved origin**, including `getAccount`
and `getNetwork`. Both disclose: one the user's identity, the other the endpoint
they chose, which is a durable fingerprint and sometimes a private or paid URL.
Answering either before the user has agreed to talk to the origin discloses what
they have not agreed to disclose.

**A gated method called from an unapproved origin answers `not_connected`.** Not
`no_signer` — the wallet may hold plenty of accounts, and telling a page its
`getNetwork` failed for want of a signer sends it looking for a problem that is
not there.

**A wallet that gates MUST implement `connect`**, otherwise `not_connected` names
no way forward and the page is simply stuck. `connect` is listed as optional
because a wallet that gates nothing does not need it — not because a page can be
left without a route to what it was refused.

**A signing request MAY carry its own approval.** `sendTx` is the core method and
`connect` is optional, so a page may reasonably implement `sendTx` alone; a
wallet that gates then has two conforming answers — perform the origin approval
as part of the request (the user sees the connect approval, then the transaction
approval), or refuse with `not_connected`. A page MUST handle both: on
`not_connected`, call `connect` and retry. Whichever the wallet does, it MUST NOT
sign before the origin is approved.

A page that wants the connection established deliberately — to show who is
connected before offering anything to sign — calls `connect` and does not rely on
either behaviour.

### Announcements are untrusted

Any script running in the page can dispatch `gno:registerWallet`, including
one injected by a compromised dependency. A page that builds a wallet chooser
from announcements is rendering attacker-controllable input, so it MUST:

- render `name` as text, never as markup, and clamp its length;
- accept `icon` only as a `data:image/` URI — a remote URL would both leak a
  page visit and let the entry render arbitrary fetched content;
- cap how many announcements it accepts, so a flood cannot push the real
  wallet out of the list.

None of this authenticates the wallet: the user picking a name from a list is
the trust decision, exactly as when they install an extension. What the page
owes them is that the list is legible and cannot be crowded out.

## Launch Links (external wallets)

Launch links hand an intent off to an external wallet — a mobile app or
standalone desktop signer registered under a custom URL scheme. They reach
what in-page discovery structurally cannot: a wallet that runs outside the
browser has no `window` to announce itself on. Gnoweb emits them from `$help`
Execute; any producer may author them.

The URL's host component selects the verb, and hosts are matched
**case-insensitively** — RFC 3986 makes the host component case-insensitive and
implementations normalise it, so a wallet lowercases before comparing.

| host | message | broadcasts? |
|---|---|---|
| `sendtx` | `MsgCall` | yes |
| `signtx` | `MsgCall` | no — returns the signed tx to the producer |
| `connect` | — | asks for the user's on-chain identity; with `pubkey`, also has a producer-held key authorised as a session of it |
| `disconnect` | — | ends what the wallet keeps for the producer; with `pubkey`, also revokes that producer-held session |

`send…` signs and broadcasts, `sign…` signs only. `MsgRun` follows the same
naming when it lands (`sendrun` / `signrun`); it is a separate host rather than a
mode of `sendtx` because its payload is a package of source files, sharing no
parameters with a call.

**Both axes are hosts, on purpose.** An unknown query parameter is silently
ignored — a wallet that didn't understand a `broadcast=false` flag would
broadcast anyway, exactly what a sign-only producer must never allow — whereas an
unknown **host** is declined with `unsupported_host`. The same argument rules out
a `type=run` parameter: it is one that must *not* be ignored, so it cannot be a
parameter. Anything whose absence would leave a dangerous default belongs in the
verb.

(A future multi-message bundle mixing calls and runs cannot select its message
type by host, since a host covers the whole request. There the type becomes a
**required** field per message — `msg.<i>.type` — which has no dangerous default
because its absence is `invalid_request` rather than a silent fallback.)

**Forward compatibility.** The standard evolves additively: a new capability is
always a new query parameter, host, or (in-page) method — existing ones are
never repurposed. Receivers therefore MUST tolerate what they don't recognise:

- A wallet MUST ignore query parameters it does not understand, and MUST NOT
  reject a request for containing them.
- A wallet that receives a host it does not implement SHOULD answer
  `status=error&code=unsupported_host` when the request carries a `callback`
  (the general answer-duty below); with no callback there is nothing to answer.
  Only `callback` and `state` may be read from such a request: every host shares
  those two, but a verb the wallet does not know may define its other parameters
  however it likes.
- For that answer to be possible at all, a wallet SHOULD register for the whole
  custom scheme rather than for an enumerated list of hosts. Where the OS routes
  links per host — an Android intent filter naming each `android:host`, say — a
  host the wallet never declared does not reach it, and it cannot answer.

A producer must therefore tolerate **both** shapes of refusal: an
`unsupported_host` callback, and a purely local launch failure when no installed
app claims the link at all. Neither is guaranteed — a wallet predating this rule
may accept the launch and simply do nothing — so **a launch is not a promise of
a response**, and a producer MUST NOT treat one as pending indefinitely.

There is no version field: additivity plus ignore-the-unknown is the whole
compatibility contract.

**Payload size.** A launch link is a URL, so it is bounded by the platform's
URL-length limits (no universal figure; keep well under ~2 KB). Launch links
suit ordinary calls, not large payloads — a bulk `MsgRun` body or very large
arguments belong on the in-page transport (no such limit) or another channel.
This bites `sendrun`/`signrun` hardest: a `MsgRun` carries whole source files, so
most run payloads will not fit in a URL at all.

**Wallet not installed.** A custom-scheme link requires the wallet already
installed; with none registered for the scheme the OS behaviour is
platform-specific and there is no in-protocol fallback. Producers should keep the
always-available copy-paste TxLink command as the wallet-agnostic fallback
(gnoweb renders it beside Execute); a graceful "not installed" path needs
Universal / App Links and is out of v1.

**Encoding.** Names and values are percent-encoded:

- Producers MUST percent-encode (`encodeURIComponent`). A literal plus is
  `%2B`.
- Wallets MUST accept form-encoded argument values as well: in `arg.<name>`
  and `args` values, `+` decodes to a space. Substitute **before**
  percent-decoding, so `%2B` still yields a literal `+`.
- Everywhere else — `path`, `func`, `send`, `rpc`, `chainid`, `callback`,
  `state`, `signer` — `+` is a literal plus and is not substituted.

The leniency is there because `URLSearchParams`, the obvious way to build a
link in a browser, emits `application/x-www-form-urlencoded`, where a space is
`+`. A wallet parsing strictly per RFC 3986 shows the user `testing+board` for
a board they named `testing board`, and signs that. The leniency stops at
argument values because elsewhere a `+` may be data: `state` is often base64,
and rewriting it would break the correlation check it exists for.

### `sendtx` — review, sign, broadcast

```
<scheme>://sendtx?path=<pkgPath>&func=<Foo>&arg.<name>=<value>&send=<coins>&rpc=<rpc>&chainid=<chainid>&callback=<url>&state=<token>&signer=<address>
```

- `<scheme>` is the wallet's registered custom scheme (e.g.
  `land.gno.gnokey`).
- Function arguments are named like TxLink arguments, but namespaced under
  `arg.` so realm parameter names cannot collide with the link's own reserved
  keys (`path`, `func`, `send`, `rpc`, `chainid`, `callback`, `state`,
  `signer`). The positional form is repeated `args=<value>`. One form per link
  (see Arguments), and a link may prefill only some named arguments.
- `send` (optional) is the coin amount to attach, in `gnokey` coin syntax
  (e.g. `1000000ugnot`).
- `chainid` is **required**: it selects the network (see Network resolution). A
  link without one is `invalid_request`.
- `rpc` (optional) is advisory. The wallet does not query or broadcast through
  it; its only use is prefilling an add-network proposal for a chain the wallet
  does not have. It may be scheme-less (`127.0.0.1:26657`), in which case
  `http://` is assumed.
- `callback` (optional) is the URL the wallet reopens with the result.
- `state` (optional, RECOMMENDED) is an opaque producer-generated token,
  echoed verbatim in every callback. A callback scheme is public — anything
  installed can open it — so without `state` a producer cannot tell its own
  result from one an attacker synthesised. Producers that consume callbacks
  should always send one and drop responses that match no outstanding request.
  The wallet treats `state` as opaque and SHOULD bound its length (e.g. ≤256
  characters).
- `signer` (optional) pins the **identity** the producer expects to act as — the
  `address` from a prior `connect`. See The `signer` pin, which governs both
  transports: the wallet MUST produce a transaction authorised by that identity
  or decline with `status=error&code=signer_unavailable`, and MAY sign it with a
  delegated key of that identity.

The `sendtx` host always signs **and broadcasts**; the callback returns `hash`. User
review before signing is mandatory. A producer that needs the signed transaction
*without* broadcasting uses the `signtx` host below.

**One message per link.** A `sendtx` link carries a single `MsgCall`. Multiple
messages are a planned additive extension — an indexed `msg.<i>.path` /
`msg.<i>.func` / `msg.<i>.arg.<name>` form, with today's flat fields the implicit
`msg.0`, advertised as the `multi_msg` feature (see `connect`). The `arg.`
namespace keeps that path collision-free, so it lands without migration; v1 has
no atomic multi-call bundle.

#### `sendtx` callback results

The wallet appends its response to `callback`:

```
<callback>?status=success&hash=<txhash>&state=<echoed>       # signed and broadcast
<callback>?status=cancelled&state=<echoed>                   # user declined
<callback>?status=error&code=<code>&state=<echoed>           # signing/broadcast failed
```

`status` is the outcome class — `success`, `cancelled`, or `error` — a closed
set. On `error`, `code` carries an enumerated, machine-readable reason (never
human text; producers MUST NOT parse it as prose):

- `invalid_request` — the request was malformed: no `chainid`, named and
  positional arguments mixed, or an argument naming no declared parameter.
- `network_declined` — the user rejected the network switch, or declined to add
  a chain the wallet does not have.
- `signer_unavailable` — the wallet holds no account resolving to the pinned
  `signer`, so it cannot produce a transaction authorised by that identity.
- `no_signer` — the wallet holds no account to sign with at all.
- `not_connected` — **in-page only.** The origin has not been approved and the
  wallet gates this method on that; see Connecting, and what it gates. A launch
  link carries its own consent, so this never appears in a callback.
- `unsupported_host` — the wallet does not implement the requested verb.
- `tx_failed` — the wallet could not sign or broadcast the transaction. The
  wallet has already shown the user the cause; the producer should still confirm
  on-chain, since a failure reported here does not guarantee nothing landed.

New reasons are added to `code`, never to `status`, so a producer's
`status=error` branch keeps working as the set grows.

`state` is echoed on **every** response, including failures, and is absent when
the request omitted it.

A wallet SHOULD answer every request it accepted — a producer waiting on a
callback cannot see an error surfaced on the user's device, and without a
`cancelled` or `error` response it waits indefinitely.

`hash` is a hint, not proof: the callback scheme is public, so a producer
should confirm the transaction on its own RPC before treating it as landed.

### `signtx` — review and sign, no broadcast

```
<scheme>://signtx?path=<pkgPath>&func=<Foo>&arg.<name>=<value>&send=<coins>&rpc=<rpc>&chainid=<chainid>&callback=<url>&state=<token>&signer=<address>
```

Identical to `sendtx` field for field, but the wallet **signs and returns the signed
transaction without broadcasting** — the producer broadcasts it on its own RPC.
This suits a dapp that owns its connection to the chain and only needs a
signature. User review before signing is mandatory, exactly as for `sendtx`. A wallet
that does not implement sign-only answers `unsupported_host` rather than falling
through to a broadcast, so a producer's "do not broadcast" is guaranteed by the
protocol shape, not by the wallet's goodwill. The single-message limit and the
`msg.<i>` multi-message extension apply exactly as for `sendtx`.

This is the host where signing offline is reachable (see Arguments): with
positional arguments the wallet needs no `vm/qdoc` lookup, and with the chain id,
account number, sequence and gas supplied by the user it needs no endpoint at
all. It still resolves the chain against a configured network, so the user is
told what they are signing for.

Sign-only moves real obligations to the producer, and they are easy to miss:

- **It must be able to broadcast what the wallet signed.** A wallet may sign with
  a scheme the producer's client does not know — a session key, a multisig — and
  a client that cannot represent that signature will re-encode the transaction
  into an invalid one rather than refuse it. The failure surfaces at the very
  last step and looks like the wallet's fault.
- **It owns the errors.** Out-of-gas, a rejected signature, a realm that refuses
  the call: all arrive at the producer, about a transaction the wallet composed,
  once the wallet's review screen is gone.
- **`status=success` means _signed_, not _landed_.** Nothing has been broadcast
  when the callback fires; a producer that treats it as completion will report
  success for a transaction that never reached the chain.
- **It does not learn who signed unless it pinned `signer`.** `signedtx` is
  opaque by the rule above, so a producer cannot read the identity out of it —
  and the reason it must not is exactly the reason it could not do so reliably:
  a session or multisig signature needs a client able to represent it. Without a
  pin, a `connect` earlier in the session proves nothing, because the user may
  have switched accounts since. A producer that will attribute the transaction to
  someone — gate on it, credit it, show it — MUST pin `signer`, and therefore
  MUST have called `connect` to obtain the identity to pin. Reading the caller
  back from the chain after broadcasting also works, but only once the
  transaction is already away.

Prefer `sendtx` when the producer has no specific reason to broadcast itself: the
wallet built the transaction, resolved the account sequence, and understands its
own signatures, so it is better placed to report what happened.

#### `signtx` callback results

```
<callback>?status=success&signedtx=<base64>&state=<echoed>   # signed, not broadcast
<callback>?status=cancelled&state=<echoed>                   # user declined
<callback>?status=error&code=<code>&state=<echoed>           # signing failed
```

`signedtx` is the signed transaction as **amino-binary, base64-encoded** — the
exact string `broadcast_tx_sync` / `broadcast_tx_commit` take as their parameter,
so a producer broadcasts it by passing it straight through.

**A producer MUST treat `signedtx` as opaque and broadcast it unmodified.** This
is what makes the obligation above ("it must be able to broadcast what the wallet
signed") satisfiable rather than merely stated. Decoding and re-encoding requires
a client that can represent whatever scheme the wallet signed with; a session key
or a multisig carries fields a generic client will drop, producing a
well-formed-looking but invalid transaction that fails at the last step and looks
like the wallet's fault. Passing the bytes through means the producer never needs
to understand the signature at all.

The `status` / `code` envelope and code set are the same as `sendtx`'s, except
`tx_failed` here always means signing failed (nothing is ever broadcast). `state`
echoing and the answer-every-request duty are identical.

### `connect` — request the user's identity

```
<scheme>://connect?callback=<url>&state=<token>&rpc=<rpc>&chainid=<chainid>
```

Asks the wallet which address the user wants to act as — the sign-in step
before any `sendtx`. `callback` is **required**: the verb exists only to deliver an
answer, so a request without a usable one is dropped. `state` behaves as for
`sendtx`.

`chainid` is **optional** here, unlike on `sendtx` and `signtx`. Given, it is
resolved exactly as theirs is (see Network resolution) and a `connect` may
therefore prompt the user to switch, or to add a chain the wallet does not have,
before it answers. Omitted, the wallet answers against its active network and
reports which in the callback's `chainid`. Nothing is signed either way, so
there is no chain-bound artefact for a missing `chainid` to spoil — and a
producer that only wants to know who the user is should not have to guess a
chain to ask. `rpc` is advisory as everywhere.

A `connect` may also carry **scope hints** (`allow`, `spend`, `period`,
`expires`) describing what the producer will send through the wallet, or a
**`pubkey`** naming a key the producer holds, to have it authorised as a session
of the identity. Both are optional and additive: a `connect` without them means
exactly what this section says. See Sessions below.

The wallet MUST ask the user before disclosing anything, and MUST show the
callback's host: a producer's claimed name is self-asserted and unverifiable,
so the destination is the only anti-phishing anchor the user has. The protocol
carries no producer-supplied display name — a producer's identity to the user
is its callback destination.

```
<callback>?status=success&address=<bech32>&session=<bech32>&pubkey=<gpub>&chainid=<id>&features=<tokens>&state=<echoed>
<callback>?status=cancelled&state=<echoed>
<callback>?status=error&code=<code>&state=<echoed>
```

Error codes (`code`): `no_signer`, `network_declined`, `invalid_request`, and for
a `connect` naming a key the session codes listed under Sessions. As on `tx`,
`status` is the closed outcome class and `code` the enumerated reason.

`address` is the identity, and it is the only one of these a later request may
pin as its `signer`. `session` (optional) is the delegated key the wallet
currently signs with, if it uses one: informational — it lets a producer read the
right account on chain — and explicitly **not** pinnable, because it rotates. A
wallet that signs directly with the identity omits it. The in-page `GnoAccount`
carries no equivalent field for the same reason: nothing a producer must do
depends on knowing the key.

`features` (optional) is a comma-separated list of the wallet's optional
capabilities, letting a producer tailor later requests. v1 tokens: the **hosts**
the wallet supports — `sendtx` (sign and broadcast) and `signtx` (sign only) —
plus `multi_msg` (accepts the indexed multi-message form) and `sessions` (grants
producer-held sessions: `pubkey` and `old` on `connect`, and the `disconnect`
host — see Sessions). The two tx hosts are
independent, so a pure signer may offer `signtx` without `sendtx`. `multi_msg`
extends the single-message baseline (every wallet handles at least one message).
A wallet that omits `features` is making no claim, and a producer should assume
nothing beyond the single-message baseline. Unknown tokens are ignored. A
producer may also simply attempt a host and treat `unsupported_host` as the
negative answer.

The returned identity is **display-level**. It carries no challenge and no
signature, so it proves nothing about control of the address: treat it as the
user stating who they are, not as authentication. Authority comes from the
on-chain `tx` the user reviews and signs. A proof-of-control extension
(challenge + signature) is left for producers with a backend able to verify one.

### Sessions: `connect` with a key, and `disconnect`

The standard treats sessions as the wallet's business: a producer pins an
identity (`signer`), never a key, and the wallet chooses which key signs. This
section keeps that model and gives producers exactly two verbs. A producer
**connects** and **disconnects**; which sessions exist, which one signs, and how
they are created, renewed and cleaned up is decided in the wallet, with the user.

A producer connects in one of two ways:

| the producer wants to | `connect` carries | the wallet | afterwards the producer |
|---|---|---|---|
| send through the wallet | nothing new, or optional **scope hints** (`allow`, `spend`, `period`, `expires`) | answers with the identity; it MAY first pick one of its own sessions that covers the hints, or offer to create one | sends `sendtx` / `signtx` pinned to `address` |
| sign by itself | a **`pubkey`** it holds, the grant it needs, and a proof of possession | authorises that key on chain as a session of the identity, inside the limits the user approved | signs every transaction with its key, without going back to the wallet |

The two are compatible. A producer that holds its own session key can still send
`sendtx` requests: the wallet signs them with its own key or one of its own
sessions. Both are sessions of the same identity, which is what the chain
credits and what `signer` pins.

**Renewing, changing the budget or the allowed paths** is a new `connect` with a
fresh key that names the current one in `old`. The wallet revokes the old
session and creates the new one in a single transaction. A producer may also
`disconnect` and then `connect` again, which costs two transactions and leaves a
gap between them.

**Why `disconnect` is a host and the grant is a parameter.** A wallet silently
ignores query parameters it does not understand (see Forward compatibility).
Ignoring `pubkey` is safe: a wallet that does not know it answers an
identity-only `connect` and grants nothing, and the producer detects it (see
Callback results). Ignoring a revocation would not be safe, since the producer
would believe a key was disabled when it still works. So revocation is the host
`disconnect`, and a wallet that does not implement it answers
`unsupported_host`.

Sessions are carried by launch links only in this version. The in-page
transport, and sessions a session could create itself (attenuated
sub-sessions, which would need consensus changes), are not covered.

#### Two kinds of session

A session the wallet signs `sendtx` and `signtx` requests with is
**wallet-held**: the producer never sees it, and its limits are the wallet's
business. Scope hints only help the wallet choose or prepare one; they give the
producer no handle on it.

A session requested with `pubkey` is **producer-held**. The key is generated
by, stored by and signs inside the producer. The wallet acts only as a
*broker*: it shows the user the grant, obtains the identity's signature on the
`auth` messages, and reports the result. Two consequences:

- **The private key never travels.** Only its public key appears in a request. A
  wallet MUST NOT generate a session key for a producer, and a producer MUST NOT
  send a private key in any parameter. It follows that a wallet cannot hand one
  of its existing sessions to a producer: with `pubkey`, the session is always
  the producer's own key.
- **The producer can only act on its own session.** Every request that names a
  key carries a proof of possession made with that key (see Proof of
  possession). The producer cannot list, read, replace or revoke any other
  session of the identity. The wallet stays responsible for housekeeping the
  identity's sessions (see Wallet housekeeping), and no request can ask for it.

#### Chain rules a producer must know

These are gno.land consensus rules today (`tm2/pkg/sdk/auth`,
`gno.land/pkg/gnoland`), not choices made by this standard:

- **A session cannot sign `auth/*` or `vm/add_package`.** Even a `*` grant
  excludes them. Only the identity's own key can create, replace or revoke a
  session, so a wallet that only holds a session key needs some other route to
  that key (see Obtaining the identity's signature).
- **Sessions cannot be updated.** Changing limits means revoking and creating
  again, which is what a `connect` with `old` does.
- **At most 16 sessions per identity, and expired ones count.** The chain never
  deletes an expired session; only `MsgRevokeSession` removes one.
- **At most 8 allow entries per session.** The grammar is `*`, or
  `<route>/<type>[:<path>]` with `<route>/<type>` one of `vm/exec`, `vm/run`,
  `bank/send` or `bank/multisend`. Only `vm/exec` takes a `:<path>`, matched as
  `path == entry` or `path` starting with `entry + "/"`.
- **The spend limit covers every outflow** charged to the identity for the
  session's transactions: gas fees, `MsgCall.Send`, bank sends and storage
  deposits. An empty limit means the session cannot spend anything, so it
  cannot even pay its own gas.
- **`period` is a resetting window.** Once `period` seconds have passed since
  the window opened, the next spend resets the counter and opens a new window
  at that spend's block time. `0` makes the limit a lifetime cap. Maximum:
  2,592,000 seconds (30 days).
- **The budget counts gross outflows.** Coins flowing back to the identity,
  such as a realm paying out winnings, do not lower the session's used amount.
- **A spend over the limit is refused before any fee is taken.** The session's
  total outflow is checked against the remaining budget first.
- **Maximum lifetime is 126,144,000 seconds (about 4 years).**
- **A session holds no coins.** Everything is debited from the identity's
  balance, so a session is only as funded as its identity. Anyone can fund the
  identity by sending to its address, and no session operation is involved.
- **The same key can be revoked and recreated in one transaction.** The new
  session gets a new account number and a sequence of 0, so transactions signed
  for the old one cannot be replayed.
- **Fees are charged on most rejections, and in full.** A message that fails in
  its handler (a bound exceeded, a duplicate key) still costs the transaction's
  fee, because only `ValidateBasic` (for example the 8-entry cap) and the
  session spend check reject a transaction before fees are taken. The fee
  charged is the whole `gas-fee` offered, whatever gas was used, so whoever
  composes a transaction derives it from the chain's minimum gas price
  (`auth/gasprice`) times the gas wanted, with some headroom, rather than
  offering a fixed large amount.
- **Every message in a transaction succeeds or none does.** A revocation
  followed by a failing creation leaves the revoked session in place.

#### Parameters

| parameter | `connect` (identity) | `connect` with `pubkey` | `disconnect` | `disconnect` with `pubkey` | meaning |
|---|---|---|---|---|---|
| `pubkey` | — | required | — | required | the producer's key, bech32 `gpub1…` |
| `old` | — | optional | — | — | the producer's current session key, `gpub1…`, to revoke in the same transaction |
| `allow` | optional hint, 1–8, repeatable | required, 1–8, repeatable | — | — | allow entries, chain grammar |
| `spend` | optional hint | optional | — | — | spend limit per period, `gnokey` coin syntax (`5000000ugnot`) |
| `period` | optional hint | optional | — | — | seconds, `0`–`2592000`; absent means `0` |
| `expires` | optional hint | required | — | — | lifetime in **seconds**, `1`–`126144000` |
| `chainid` | required | required | — | required | network, resolved as in Network resolution |
| `rpc` | optional | optional | — | optional | advisory, as everywhere |
| `signer` | — | optional; **required** with `old` | — | **required** | the identity (`address` from a previous `connect`) |
| `callback` | required | required | optional | optional | where the wallet answers |
| `state` | optional, RECOMMENDED | **required** | optional, RECOMMENDED | **required** | opaque correlation token, at most 256 characters |
| `iat` | — | **required** | — | **required** | issue time, unix seconds |
| `sig` | — | **required** | — | **required** | proof of possession by `pubkey` |
| `oldsig` | — | **required** with `old` | — | — | proof of possession by `old` |

Rules:

- **`pubkey` sets the mode.** A `connect` with `pubkey` asks for a grant to
  that key. Without it, `allow`, `spend`, `period` and `expires` are hints for
  the wallet's own sessions, and the request is otherwise a plain `connect`.
  Likewise a `disconnect` with `pubkey` revokes that key, and one without it
  only ends what the wallet keeps for the producer.
- **`old` comes with `pubkey`.** A wallet that implements `pubkey` MUST
  implement `old`. So the only wallets that ignore `old` are the ones that
  ignore `pubkey` too, and they grant nothing. `old` MUST differ from `pubkey`,
  and requires `signer` and `oldsig`; otherwise the request is
  `invalid_request`.
- **`expires` is a duration, not a timestamp.** The identity's signature may come
  minutes after the request (see Obtaining the identity's signature). For a
  grant, the lifetime counts from when the wallet composes the transaction (see
  Time below); for a hint, from when the request arrived. A producer cannot ask
  for a session with no expiry: an unlimited lifetime is a decision the user
  takes in their wallet, not something a dapp can request.
- **`callback` is required on `connect`:** the producer needs the answer, and
  with no `signer` it has no other way to learn the identity. On `disconnect` it
  is optional, because the producer can confirm the revocation on chain by
  itself.
- **`state` is required when a key is named,** not just recommended, because it
  is part of the signed payload and the replay defence.
- **`signer` is required with `old` and on a `disconnect` with `pubkey`.** They
  act on an existing session, and a session lives under one identity. Without `old` it is
  optional, and a missing pin means "whichever identity the user picks". A
  producer that already knows the identity SHOULD pin it.
- **Encoding** follows the launch-link rules. `+` is a literal plus in every
  parameter above (none of them is an argument value), and repeated `allow`
  parameters keep their order.
- **The usual validation applies.** A wallet answers `invalid_request` to a
  request with a missing required parameter, a malformed key, an allow entry the
  chain grammar rejects, more than 8 entries, or an out-of-range `period` or
  `expires`. It validates hints the same way, so the grammar is the same
  whichever mode a producer uses.

#### Proof of possession

Every request that names a key is signed by that key, so a producer can only
have a key authorised, replaced or revoked if it controls it.

**Payload.** Take all query parameters of the request except `sig` and `oldsig`,
percent-decode them, and sort them by key then value, bytewise in UTF-8 — which
is Unicode code-point order. (A UTF-16 comparison, the default for strings in
JavaScript, orders characters above U+FFFF differently: compare code points.)
The payload is these lines joined with `\n` (LF), with no trailing newline:

```
gnoconnect-session-v1
host=<host, lowercased>
<key>=<value>
<key>=<value>
…
```

- **No control characters.** A name or value containing one — C0, DEL or C1,
  as Go's `unicode.IsControl` defines them, line breaks included — is
  `invalid_request`.
- **No `=` in parameter names.** A decoded name containing `=` is
  `invalid_request`: `a%3Db=c` and `a=b%3Dc` would give the same line, so a
  relay could rename a signed parameter without breaking the signature.
- **Everything in the request is covered,** including `callback`, `chainid`,
  `iat`, `state`, `signer` and `old`. Changing any of them breaks the
  signature, so a request redirected to another callback or replayed on another
  chain is refused.
- **The host is covered too** (`connect` or `disconnect`), so a signed
  `connect` cannot be replayed as a `disconnect`, or the reverse.
- **The `gnoconnect-session-v1` prefix** keeps this payload from ever looking
  like a transaction sign document (which starts with `{`). That matters because
  the same key will sign transactions.
- **Parameters added by future extensions are signed too,** since every
  parameter except the signatures is covered. A wallet that does not know a
  parameter still verifies the signature over it, then ignores it.

**Signature.** It is made with the key's own scheme, as tm2's `PrivKey.Sign`
does, and checked as `PubKey.VerifyBytes` does:

- secp256k1: ECDSA over SHA-256(payload), 64 bytes `R‖S`, low-S. Wallets MUST
  support it.
- ed25519: RFC 8032 over the payload, 64 bytes. Wallets MAY support it, and
  answer `invalid_proof` if they do not.

`sig` and `oldsig` are **base64url without padding** (RFC 4648 §5). Both sign
the same payload.

**Wallet checks, before showing anything to the user:**

1. `sig` verifies against `pubkey`, and with `old` `oldsig` verifies against
   `old`. Otherwise answer `invalid_proof`.
2. `iat` is no more than 10 minutes old and no more than 1 minute in the future
   by the wallet's clock. Otherwise answer `stale_request`. (Ten minutes only
   bounds how long a link may wait before the wallet opens it; the identity's
   signature may come later.)
3. The wallet SHOULD remember the `(pubkey, state)` pairs it has accepted for
   that window and drop a repeat without answering: the first copy is being
   handled, and answering the copy could close the producer's request early.
   A request the wallet refused and receives again is not answered a second
   time either.

**What it proves.** That the requester holds the private key, and nothing about
*who* the requester is. As for every `connect`, the producer's identity to the
user is its callback destination, which the wallet MUST show. What the proof
stops is one producer replacing or revoking another producer's session, and a
producer getting a key it does not control authorised.

#### Time

A session's expiry is enforced against block time, so a wallet judges it the
same way: it takes "now" as the time of the latest block, never later than its
own clock, when it decides whether a session is expired and when it composes
`ExpiresAt = now + expires`. A node whose latest block carries no usable time
(no block yet, or still syncing) gives no such time, and the wallet does not
compose against it. A grant that would already be over when it lands (a chain
that has produced no block for longer than `expires`) is refused before the
user is asked.

#### `connect` without a key: scope hints

```
<scheme>://connect?allow=<entry>&allow=<entry>&spend=<coins>&period=<s>&expires=<s>&chainid=<id>&callback=<url>&state=<token>
```

The hints describe what the producer intends to send through the wallet. The
wallet MAY use them to choose a session it already holds for the identity that
covers them, or to offer the user to create one, before it answers, so the
first `sendtx` does not stop on a missing or too narrow session. A session
covers the hints when it allows every hinted entry, its budget is at least the
hinted one (and, when the hint names a period, resets at least that often), and
it lasts at least `expires` counted from the request's arrival (a wallet allows
a minute or so of slack, since block time trails the clock). A session with no spend limit cannot
pay its own gas, so it covers no hint. What the wallet does with the hints is
entirely its own decision:

- The user may also choose to sign with the identity itself, or to decide
  later. The answer is the same in every case: the `connect` callback above.
- Creating a wallet-held session is a grant like any other. Its review follows
  the Review section, and its costs are paid by the identity.
- A hint creates no obligation. The producer learns nothing about the session,
  and a later `sendtx` that the session cannot cover is the wallet's to handle
  in its own review.

#### `connect` with a key

```
<scheme>://connect?pubkey=<gpub>&allow=<entry>&allow=<entry>&spend=<coins>&period=<s>&expires=<s>&chainid=<id>&rpc=<rpc>&signer=<address>&callback=<url>&state=<token>&iat=<unix>&sig=<b64url>
```

The wallet:

1. Resolves the network, then the identity: the pinned `signer`, or one the user
   picks. If it cannot obtain a signature from the pinned identity, it answers
   `signer_unavailable`; if it has no identity at all, `no_signer`.
2. Looks up `pubkey` as a session of that identity
   (`auth/accounts/{address}/session/{session address}`):
   - **Present and not expired:** nothing to sign. After the user agrees to
     disclose the identity, as for any `connect`, the wallet answers with the
     grant read from the chain. This makes `connect` safe to repeat: a producer
     whose callback was lost asks again with the same key and gets its answer.
   - **Present and expired:** the wallet answers `session_expired`. The
     producer connects again with a fresh key and names the expired one in
     `old`, which frees its slot.
3. Checks on chain everything else the chain would reject, because a rejected
   transaction still costs its fee: the identity has a free session slot, or the
   wallet can free one (see Wallet housekeeping). If not, the wallet answers
   `session_limit`.
4. Shows the consent screen (see Review).
5. Composes one transaction: `MsgRevokeSession` for `old` if given and on chain,
   any housekeeping revocations, then `MsgCreateSession{Creator: identity,
   SessionKey: pubkey, ExpiresAt: now + expires, AllowPaths: allow, SpendLimit:
   spend, SpendPeriod: period}`, with "now" as in Time. It obtains the
   identity's signature and broadcasts.
6. Waits until the session is visible on chain, then answers.

**Narrowing.** The user MAY narrow the grant on the review screen: a shorter
lifetime, a lower spend limit, a longer period, fewer allow entries. The user MUST
NOT widen it: the producer did not ask for more, and more authority on a key a
third party holds helps no one. The callback reports what was actually granted,
and the producer MUST work with that.

**The network can change under a pending request.** A switch the user accepts
for another request (a `sendtx` for another chain, say) leaves a request parked
for review on a network the wallet is no longer on. Before planning or
composing it, the wallet resolves its network again: it offers to switch back,
or answers `network_declined`. Conversely, declining a switch answers
`network_declined` only to the requests that named that other chain; a parked
request for the current chain stays pending.

##### Renewal and rotation: `old`

```
<scheme>://connect?pubkey=<new gpub>&old=<gpub>&allow=…&spend=…&period=…&expires=…&chainid=<id>&signer=<address>&callback=<url>&state=<token>&iat=<unix>&sig=<b64url>&oldsig=<b64url>
```

This is how a producer renews before expiry, changes the budget or changes the
allowed paths. The rules of a `connect` with a key apply, plus:

- **The new grant is new consent.** The review MUST show the old and new limits
  side by side and highlight any widening: a longer lifetime, a higher limit, a
  shorter period, new allow entries.
- **Revocation and creation go in one transaction,** revocation first. The
  transaction is atomic: if the creation fails, the old session is still there.
  This is what makes `old` worth having over `disconnect` then `connect`: one
  signature by the identity, one fee, and no moment without a working session.
- **An `old` that has expired but is still stored is valid:** renewing it is
  exactly what `old` is for.
- **An `old` that is no longer on chain** (the user revoked it, or a previous
  renewal already did) does not fail the request. The proof still shows the
  producer held it; there is simply nothing to revoke. The review says so, the
  transaction creates the new session only, and the callback omits `revoked`.
- **A fresh key every time.** Since `old` and `pubkey` differ, a renewal is also
  a rotation: a copy of the old key that leaked stops working.

#### `disconnect`

```
<scheme>://disconnect?callback=<url>&state=<token>
<scheme>://disconnect?pubkey=<gpub>&chainid=<id>&signer=<address>&callback=<url>&state=<token>&iat=<unix>&sig=<b64url>
```

- **It always ends what the wallet keeps for the producer,** whatever that is
  (in-page, the origin's approval; a wallet that keeps nothing per producer has
  nothing to end). Without `pubkey` that is all it does, and the wallet answers
  `success` without `revoked`.
- **Without `pubkey` the request is not authenticated:** on a launch link the
  callback's host is self-declared. So it may only remove, never grant, and the
  wallet asks the user before any on-chain revocation it would derive from it
  (for example of a wallet-held session it created for this producer's hints).
  A producer that connected without a key may send it, or simply forget the
  address.
- **With `pubkey`, it also ends that producer-held session.** `pubkey` must be a
  session of `signer` on chain, or the wallet answers `session_not_found`.
- **The user still confirms.** Revoking only removes authority, but it costs gas
  to the identity and may need the identity's signature from another device. The
  wallet MAY use a lighter screen than for a grant.
- **The wallet MAY add housekeeping revocations** to the same transaction, as on
  `connect`, and shows them in the review.
- **A producer that has lost its key cannot ask for this.** The user then revokes
  from the wallet's own interface, which is wallet business.

#### Obtaining the identity's signature

This standard does not say how the wallet obtains the identity's signature on
the `auth` messages, only that the user approved the content first. Possible
routes:

- **The wallet holds the identity's key.** One approval, signed on the device.
- **A hardware signer.**
- **A desktop handoff,** for a wallet that only holds a session key and cannot
  sign `auth/*` itself. The wallet shows the user a command or QR code for the
  identity's signer on another device, then watches the chain. For a `connect`
  with a key, no `old` and no housekeeping:

  ```
  gnokey maketx session create -pubkey <gpub> -expires-at <expires>s \
    -allow-paths <entry> … -spend-limit <coins> -spend-period <s> \
    -gas-fee <fee> -gas-wanted <gas> -chainid <id> -remote <rpc> <identity>
  ```

  `-expires-at` takes a duration (`90000s`, `7d`) counted when the command
  runs, or an absolute unix time; a duration keeps the granted lifetime whole
  however long the user takes to reach the computer. For a `disconnect` with no
  housekeeping:

  ```
  gnokey maketx session revoke -pubkey <gpub> \
    -gas-fee <fee> -gas-wanted <gas> -chainid <id> -remote <rpc> <identity>
  ```

  `-broadcast` defaults to `true` in `gnokey`, so these sign and broadcast.
  Passing `-broadcast=false` instead prints the unsigned document.

  For a `connect` with `old`, and for any transaction carrying housekeeping
  revocations, `gnokey` has no single command. The wallet writes the unsigned
  multi-message document and shows:

  ```
  gnokey sign -tx-path <file> -chainid <id> -account-number <n> \
    -account-sequence <seq> <identity> \
  && gnokey broadcast -dry-run -remote <rpc> <file> \
  && gnokey broadcast -remote <rpc> <file>
  ```

  The `-dry-run` step runs the transaction without committing it, so anything
  the chain would reject (a session list filled since the wallet checked, a
  stale sequence) stops the command before a fee is charged.
  Without it, a doomed transaction is still included and costs its whole fee.

  `gnokey sign` does not look up the account: the wallet fills in the
  identity's account number and current sequence. The document goes stale as
  soon as the identity sends any other transaction (the broadcast then fails
  with `signature verification failed`), so the wallet regenerates it when the
  sequence moves. Its `ExpiresAt` is absolute, fixed when the wallet composed
  it (see Time).

So the answer may come minutes after the request. The wallet SHOULD keep the
request pending until the chain shows the outcome, MUST let the user cancel
(answering `cancelled`), and answers once the chain confirms. Some platforms
only let an app open a URL while it is in the foreground (Android blocks
activity starts from the background), so a wallet that watches the chain in the
background answers when the user returns to it. A producer MUST treat the
request as possibly unanswered, as the launch-link rules require, and SHOULD
watch the chain itself when it can (it always can once it knows the identity).

#### Review

Before obtaining any signature, the review screen MUST show:

- **the callback's host** (the producer's only identity);
- **the network**: chain id and the endpoint in effect, as in Network resolution;
- **the identity** that grants;
- **each allow entry,** with a warning on entries that let the key reach
  arbitrary code or move funds directly: `*`, a bare `vm/exec`, `vm/run`,
  `bank/send` and `bank/multisend`;
- **the budget in plain terms** ("5 GNOT per hour") **and its worst case** over
  the whole lifetime ("at most 3,600 GNOT over 30 days"), since the period
  resets for as long as the session lives;
- **the expiry,** as an absolute date (approximate when the identity signs on
  another device: the lifetime counts from when the transaction lands);
- **with `old`,** the old and new grants side by side, widenings highlighted, or
  a note that the old key is no longer on chain and nothing is replaced;
- **every housekeeping revocation** added to the transaction;
- **for every transaction,** what the identity pays in gas for it.

#### Callback results

```
<callback>?status=success&address=<bech32>&chainid=<id>&features=<tokens>&session=<bech32>&expires_at=<unix>&allow=<entry>…&spend=<coins>&period=<s>&revoked=<bech32>&hash=<txhash>&state=<echoed>   # connect with a key
<callback>?status=success&address=<bech32>&chainid=<id>&hash=<txhash>&state=<echoed>                                                                                                       # disconnect with a key
<callback>?status=success&state=<echoed>                                                                                                                                                    # disconnect without a key
<callback>?status=cancelled&state=<echoed>
<callback>?status=error&code=<code>&state=<echoed>
```

A `connect` without a key, hints or not, answers as described under `connect`.

- **`address`** is the identity, the one a later request pins as `signer`.
- **`session`** is the bech32 address of `pubkey`, which the producer computes
  itself. In a `connect` without a key, the same field is the wallet's own
  delegated key, informational only. **A producer MUST compare `session` with
  the address of its `pubkey`.** If they differ, or `session` is missing, the
  wallet did not grant anything: it does not implement these parameters and
  answered an identity-only `connect`. The producer still has a valid
  `address`, which is what a fallback (such as a desktop command it shows
  itself) needs.
- **`expires_at`, `allow`, `spend` and `period`** are what was actually granted,
  after any narrowing, or what is on chain for an existing session (step 2).
- **`revoked`** is the address of `old` when this transaction revoked it. It is
  absent when there was no `old`, or when `old` was no longer on chain.
- **`hash`** is optional. A wallet that watched a desktop signer broadcast may
  not know it, and an answer about an existing session has none. As everywhere,
  it is a hint and not proof.
- **`pubkey`** in a callback keeps the meaning it has for `connect`: the
  identity's public key, optional. It is never the producer's key.

**A producer MUST confirm on chain before relying on the result.** After a
`connect` with a key, it queries `auth/accounts/{address}/session/{session}` and
checks `ExpiresAt`, `SpendLimit`, `SpendPeriod` and `AllowPaths`; with
`revoked`, it also checks the old session is gone. After a `disconnect`, it
checks the session is gone. Anyone can open a callback scheme. An answer whose
`address` is not the identity the producer pinned, or whose `chainid` is not the
request's, is not about its request.

**Error codes.** All of `sendtx`'s apply where they make sense:
`invalid_request`, `network_declined`, `signer_unavailable`, `no_signer`,
`unsupported_host` (a wallet without `disconnect`), `tx_failed`. Requests that
name a key add:

- `invalid_proof`: `sig` or `oldsig` did not verify, or the key type is not
  supported.
- `stale_request`: `iat` is outside the window.
- `session_expired`: `connect` named a key that is an expired session of the
  identity.
- `session_not_found`: `disconnect` named a key that is not a session of
  `signer`.
- `session_limit`: the identity already has 16 sessions and the wallet could not
  make room (see Wallet housekeeping).

#### Wallet housekeeping

Expired sessions count towards the limit of 16 and are never removed by the
chain. When it composes a `connect` or `disconnect` transaction, a wallet MAY
add `MsgRevokeSession` messages for **expired** sessions of the identity, and it
SHOULD do so before answering `session_limit`. These messages MUST appear in the
review.

This is entirely the wallet's decision: there is no request parameter for it,
and the producer neither sees nor controls it. A wallet MUST NOT revoke a session
that has not expired, other than the `old` of a `connect` or the `pubkey` of a
`disconnect`, without the user explicitly choosing to.

#### Producer obligations

- **Generate the key on the device** and keep it in the platform's secure
  storage (Keychain or Android Keystore). Never transmit it.
- **Use a fresh key for every `connect` that grants,** first and renewal alike.
- **Keep the old key until the renewal is confirmed on chain,** then delete
  it. A transaction signed with the old key may fail if the swap happens while it
  is pending; re-sign it with the new key.
- **Watch the session and renew with `old` well before expiry.** With a wallet
  that hands off to a desktop, the user needs time to reach a computer, so a
  notification several days ahead works better than a failure mid-session.
- **Read the remaining budget from the chain**, not from a local tally:
  `remaining = (now ≥ SpendReset + SpendPeriod) ? SpendLimit : SpendLimit −
  SpendUsed`. What can actually be spent is also capped by the identity's
  balance (`bank/balances/{address}`). A low balance is fixed by funding the
  identity's address, not by a session operation. Winnings paid back to the
  identity raise its balance but leave the session's budget unchanged.
- **Sign transactions in the session format**: the signature carries the session
  address and the transaction names the identity as caller and sender. A client
  that cannot represent this produces invalid transactions; check that the one
  you use supports session signers.
- **Use `disconnect` when the user logs out,** then delete the key once the
  revocation is confirmed.

#### Example: a game that needs about an hour of play at a time

The producer asks for 5 GNOT per hour on its realm only, for 30 days.
Parameters before encoding:

```
land.gno.gnokey://connect
  ?pubkey=gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq…
  &allow=vm/exec:gno.land/r/demo/game
  &spend=5000000ugnot&period=3600&expires=2592000
  &chainid=gnoland-1&rpc=https://rpc.gno.land:443
  &callback=<its https page on game.example>
  &state=Qm9vcC0xNzI3&iat=1790841600&sig=…
```

The review shows the callback host `game.example`, `gnoland-1` and its
endpoint, the realm path, "5 GNOT per hour, at most 3,600 GNOT over 30 days"
and the expiry date. After the user's approval and the identity's signature, the
answer is appended to that callback, in its fragment:

```
<callback>#status=success&address=g1…&chainid=gnoland-1&features=sendtx,signtx,sessions&session=g1…&expires_at=1793433600&allow=vm%2Fexec%3Agno.land%2Fr%2Fdemo%2Fgame&spend=5000000ugnot&period=3600&state=Qm9vcC0xNzI3
```

Three days before `expires_at`, the producer notifies the user and sends a new
`connect` with a new key in `pubkey`, the current key in `old`, `signer` set to
the identity, signed by both keys. When the user logs out, it sends
`disconnect` for the current key.

A wallet that does not grant sessions answers the first request with
`status=success&address=g1…` and no `session` for that key. The producer then
knows it must obtain the grant another way, and already knows the identity.

### Callback URL rules

A wallet opens `callback`, so it MUST constrain it:

- Accept `https:` and custom app schemes, but **reject** schemes dangerous to
  open: `javascript:`, `data:`, `file:`, `content:`, `blob:`, `about:`, and
  (Android) `intent:`.
- Require an absolute URI with a scheme, no control characters, bounded length.
- The wallet appends its response keys, preserving any parameters already in
  `callback`. If a response-key name (`status`, `code`, `state`, `hash`,
  `signedtx`, `address`, `session`, `pubkey`, `chainid`, `features`,
  `expires_at`, `spend`, `period`, `revoked`) already appears, the wallet's
  appended value is authoritative — producers MUST read the **last**
  occurrence. `allow` is the exception: it repeats in a session answer, so a
  producer cannot tell its own occurrences from the wallet's, and its callback
  MUST NOT carry an `allow` parameter.
- Response names and values are percent-encoded like the request's; a wallet
  may encode characters it need not (`expires%5Fat`), so producers MUST
  percent-decode names as well as values.
- For an `https:` callback the wallet SHOULD return the result in the URL
  **fragment** (`#status=…`) rather than the query, to keep it out of server
  logs and `Referer`. A custom-scheme callback travels no network hop, so query
  parameters are fine there. A producer reads the answer where the wallet writes
  it — the fragment for an `http(s)` callback, the query otherwise — and
  ignores the other part.
- On violation for `connect`, drop the request — there is nowhere to answer.
  For `sendtx` the callback is optional, so the wallet MAY still let the user sign,
  but MUST make clear that the requesting producer will not be notified.

## Known Implementations

Informative, not normative — ecosystem status, carrying no requirement and
conferring no standing. Nothing in this standard is specific to any entry here.

- **Gnoweb** (producer)
- **Adena Wallet** (wallet)
- **Gnokey Mobile** (wallet; on iOS and Android also grants sessions: `connect`
  with a key, `disconnect`, scope hints)
- **Bubble Rumble mobile** (producer holding its own session key)
- **Gnobro** (coming soon)
- _Add your clients here_

## Further Reading

- [Issue #2602](https://github.com/gnolang/gno/issues/2602)
- [Issue #3283](https://github.com/gnolang/gno/issues/3283)
- [PR #3609](https://github.com/gnolang/gno/pull/3609)
- [PR #3459](https://github.com/gnolang/gno/pull/3459)

