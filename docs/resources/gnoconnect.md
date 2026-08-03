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

Every request that reaches a wallet — `connect`, `sendtx`, `signtx` — resolves
its network the same way, before anything is signed or disclosed.

A **configured network** is a chain id, one or more endpoints, and the endpoint
the user has selected for it. The **active network** is the configured network
currently in use. A wallet queries and broadcasts only through the selected
endpoint of a configured network.

**1. Determine the chain.** From the request's `chainid`; for an in-page request,
falling back to the page's `gnoconnect:chainid`. If neither names a chain the
wallet answers `invalid_request` — a signature is chain-bound, so a producer
always knows which chain it wants, and "whichever the wallet happens to be on" is
how a dapp built for one chain gets a signature valid on another.

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
selected endpoint. The review screen MUST show the network name, the chain id,
and the endpoint in effect.

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

interface GnoTxIntent {
  path: string;    // full package path
  func: string;    // exported function name
  args: GnoArg[];  // one form per call; mixing is invalid_request
  send?: string;   // coins, gnokey syntax
  chainid?: string; // falls back to gnoconnect:chainid
  rpc?: string;    // advisory only — see Network resolution
  signer?: string; // bech32; MUST sign as this identity or decline
}

type UserResponse<T> =
  | { status: "Approved"; args: T }
  | { status: "Rejected"; code?: ErrorCode };
```

`ErrorCode` is the launch links' enumerated `code` set, unchanged (see `sendtx`
callback results). A rejection carries it rather than an untyped error, so a page
handles failures identically whether it called the wallet directly or handed off
a launch link:

```ts
type ErrorCode =
  | "invalid_request"
  | "network_declined"
  | "signer_unavailable"
  | "no_signer"
  | "unsupported_host"
  | "tx_failed";
```

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

`signer`, when present, pins the identity the producer expects to act as — the
`address` from a prior `connect`. The wallet MUST sign as that identity and MUST
NOT sign as another; one that cannot answers `signer_unavailable`. Without it a
page that connected as one account and rendered its address will sign as whatever
account the user has since switched to, and only find out from the chain.

#### Optional methods

A wallet MAY implement more of the surface. These are the defined shapes; a page
MUST feature-detect every method it calls rather than assume, and degrade — to
another wallet, a launch link, or the copy-paste command — when it is absent.

```ts
// Sign without broadcasting. The producer broadcasts; see signtx for the
// obligations that moves. `signedtx` is base64 amino-binary, opaque.
signTx(tx: GnoTxIntent): Promise<UserResponse<{ signedtx: string }>>;

// Ask the user which identity to act as. Discloses nothing until they agree.
connect(): Promise<UserResponse<GnoAccount>>;

// The connected identity, without re-asking.
getAccount(): Promise<UserResponse<GnoAccount>>;

// The active network, after Network resolution.
getNetwork(): Promise<UserResponse<GnoNetwork>>;

// Ask the user to switch to a configured chain. A chain the user does not have
// is network_declined, not a silent add.
switchNetwork(chainid: string): Promise<UserResponse<{ chainid: string }>>;

// Several messages, one signature, one broadcast. The launch-link analogue is
// the multi_msg feature.
sendTxs(txs: GnoTxIntent[]): Promise<UserResponse<{ hash: string }>>;

interface GnoAccount {
  address: string;         // bech32
  chainid: string;
  pubkey: string | null;   // gpub, when the wallet exposes one
}

interface GnoNetwork {
  chainid: string;
  rpc: string;             // the endpoint in effect, not one a page declared
  name: string;
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
| `connect` | — | asks for the user's on-chain identity |

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
- `signer` (optional) pins the **identity** the producer expects to act as —
  the `address` from a prior `connect`. If present, the wallet MUST sign as that
  identity and MUST NOT sign as another; a wallet that cannot MUST decline
  (`status=error&code=signer_unavailable`) rather than substitute one. What
  counts as "that identity" — an exact address match, or an account the chain
  links to it (e.g. a delegated/session key) — is wallet-specific.
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
- `signer_unavailable` — the wallet cannot sign as the pinned `signer`.
- `no_signer` — the wallet holds no account to sign with.
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
`sendtx`. `chainid` is required and `rpc` advisory, resolved exactly as for
`sendtx` (see Network resolution) — so a `connect` may prompt the user to switch,
or to add a chain the wallet does not have, before it answers.

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

Error codes (`code`): `no_signer`, `network_declined`, `invalid_request`. As on
`tx`, `status` is the closed outcome class and `code` the enumerated reason.

`features` (optional) is a comma-separated list of the wallet's optional
capabilities, letting a producer tailor later requests. v1 tokens: the **hosts**
the wallet supports — `sendtx` (sign and broadcast) and `signtx` (sign only) —
plus `multi_msg` (accepts the indexed multi-message form). The two tx hosts are
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

### Callback URL rules

A wallet opens `callback`, so it MUST constrain it:

- Accept `https:` and custom app schemes, but **reject** schemes dangerous to
  open: `javascript:`, `data:`, `file:`, `content:`, `blob:`, `about:`, and
  (Android) `intent:`.
- Require an absolute URI with a scheme, no control characters, bounded length.
- The wallet appends its response keys, preserving any parameters already in
  `callback`. If a response-key name (`status`, `code`, `state`, `hash`,
  `signedtx`, `address`, `session`, `pubkey`, `chainid`, `features`) already
  appears, the wallet's appended value is authoritative — producers MUST read
  the **last** occurrence.
- For an `https:` callback the wallet SHOULD return the result in the URL
  **fragment** (`#status=…`) rather than the query, to keep it out of server
  logs and `Referer`. A custom-scheme callback travels no network hop, so query
  parameters are fine there.
- On violation for `connect`, drop the request — there is nowhere to answer.
  For `sendtx` the callback is optional, so the wallet MAY still let the user sign,
  but MUST make clear that the requesting producer will not be notified.

## Known Implementations

Informative, not normative — ecosystem status, carrying no requirement and
conferring no standing. Nothing in this standard is specific to any entry here.

- **Gnoweb** (producer)
- **Adena Wallet** (wallet)
- **Gnokey Mobile** (wallet)
- **Gnobro** (coming soon)
- _Add your clients here_

## Further Reading

- [Issue #2602](https://github.com/gnolang/gno/issues/2602)
- [Issue #3283](https://github.com/gnolang/gno/issues/3283)
- [PR #3609](https://github.com/gnolang/gno/pull/3609)
- [PR #3459](https://github.com/gnolang/gno/pull/3459)

