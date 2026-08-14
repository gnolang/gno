# ADR: GovDAO allowlist fail-open fix and executor disclosure

## Status

Proposed.

## Context

`r/gov/dao` keeps a package-level `allowedDAOs []string`. It is the *sole*
authorization for five privileged operations:

| Operation | Site |
|---|---|
| Replace the govDAO implementation | `proxy.gno` `UpdateImpl` |
| Run a proposal's executor with DAO authority | `r/gov/dao/types.gno:221` `SafeExecutor.Execute` |
| Obtain the mutable member store | `v3/memberstore/memberstore.gno:187` `Get` |
| Move treasury funds | `v3/treasury/treasury.gno:76` `Send` |
| Rewrite the treasury's GRC20 token registry keys | `v3/treasury/treasury.gno:64` `SetTokenKeys` |

All five gate on `dao.InAllowedDAOs(caller)`, and that helper **fails open**:

```go
// proxy.gno, InAllowedDAOs
func InAllowedDAOs(pkg string) bool {
	if len(allowedDAOs) == 0 {
		return true // corner case for initialization
	}
	...
}
```

Failing open is deliberate. It is the bootstrap window: at genesis the member
set has to be seeded before any DAO exists to authorize the seeding.
`v3/loader/loader.gno` documents this, and `misc/deployments/gnoland1/gen-genesis.sh`
is meant to close the window at its step 3 of 7, by running a lockdown `MsgRun`. Because
that `MsgRun` is a *genesis transaction*, no live chain is ever left in the
fail-open state with users transacting against it.

That lockdown step does not currently work, and this change does **not** repair
it. `govdao_prop1.gno` constructs four foreign-realm struct literals — `&
memberstore.Member{}` at lines 40 and 110, `dao.UpdateRequest{}` at 118, and
`dao.VoteRequest{}` at 127 — and a `MsgRun` executes in an ephemeral
`gno.land/e/<addr>/run` realm where the VM rejects allocating another realm's
struct type (`cannot allocate ... in realm gno.land/e/.../run`).

Two different seven-step sequences are in play here, which is easy to misread.
The shell script has its own steps, and so does the file it generates. The
script builds this transaction at its step 3 of 7. Inside the file, line 40
sits in section 1 and the lockdown is section 7 — so the abort happens in the
very first section and never reaches the lockdown in the last.

Three of the four have ready-made constructors (`memberstore.NewMember`,
`dao.NewVoteRequest`, `dao.NewUpdateRequest`), and the file additionally fails
to type-check against three `r/sys` constructors that gained a leading
`cur realm`: `validators/v2.NewPropRequest`,
`params.ProposeAddUnrestrictedAcctsRequest`, and
`params.ProposeLockTransferRequest`, the last of which is still called with no
arguments at all. Repairing it is mechanical but belongs in its own change with
genesis-level verification; an earlier revision of this patch fixed only the
`UpdateRequest` line, which achieved nothing, and that has been backed out.

Note also that `gen-genesis.sh:195` invokes `maketx run` **without**
`-broadcast`: it builds and signs the transaction into genesis rather than
executing it, so `set -eo pipefail` does not catch the abort at generation
time. The failure surfaces at `InitChain`.

The bug is that the window could be **re-entered after lockdown**.
`UpdateImpl` is the only writer of `allowedDAOs`:

```go
// proxy.gno, before
if r.AllowedDAOs != nil {
	allowedDAOs = r.AllowedDAOs
}
```

and the request constructor copies its input unconditionally:

```go
// r/gov/dao/types.gno:287
func NewUpdateRequest(d DAO, allowedDAOs []string) UpdateRequest {
	cp := make([]string, len(allowedDAOs))   // non-nil even for a nil argument
	copy(cp, allowedDAOs)
	...
}
```

So `NewUpdateRequest(newDAO, nil)` — documented as "swap the implementation,
leave permissions alone", and the exact form `v3/loader` uses — produces a
**non-nil, length-zero** slice. The `!= nil` test passes, the empty slice is
stored, and `InAllowedDAOs` reverts to returning `true` for every caller on the
chain. An explicit `AllowedDAOs: []string{}` struct literal does the same.

From that state any realm can call `UpdateImpl` to install its own DAO, take
the mutable member store from `memberstore.Get`, and call `treasury.Send`. The
reachability precondition is that the current DAO executes a proposal that calls
`UpdateImpl` with an empty list —
which is the natural way to write an implementation-only upgrade, and a
plausible drafting mistake rather than an attack in itself. The severity comes
from it being silent and self-sustaining. Nothing in the render output or the
event log says the allowlist just went open, and nothing closes the window
again on its own — a later `UpdateImpl` with a real list would close it, but
only if somebody notices. Meanwhile every realm on the chain can replace the
implementation, take the mutable member store, and move treasury funds, and
those consequences do not reverse when the window is closed.

Four smaller issues were found in the same audit and are fixed here because they
sit in the same blast radius — the govDAO proposal page and the code paths that
reach it. They are listed in full so the scope of this change is not
understated: a reachable nil dereference (on two paths), a disclosure gap, and
two markdown-injection slots.

**A reachable nil dereference.** `v3/impl/govdao.gno` `PreExecuteProposal`
did `g.pss.GetStatus(pid)` and used the result without a nil check.
`ExecuteProposal` already guards the same lookup, so a proposal unknown to the
current implementation — an unknown id, or a proposal created before an upgrade
replaced the govDAO — panicked in the *pre*-check with an opaque
`runtime error: nil pointer dereference` rather than a named error. We
reproduced the panic. The same nil dereference exists on the public render path
(`renderProposalPage`), where it is reachable by anyone visiting the proposal
page of a proposal orphaned by an upgrade.

**Two attacker-controlled strings were rendered raw.** `DeniedReason` is
`"execution failed: " + err.Error()` where the error comes from the proposal's
executor callback — third-party code — and it was written directly beneath the
`PROPOSAL HAS BEEN DENIED` line, where it could forge a heading and a vote
tally. The executor's creation realm is likewise third-party controlled
(`CreationRealm()` is dispatched through the public `Executor` interface, so
only `SimpleExecutor` supplies a VM-captured value), and un-gating its
disclosure meant it would render on every proposal.

**The realm receiving govDAO authority was not rendered.**
`render.gno` gated the `Executor created in: <realm>` disclosure behind
`if p.ExecutorString() != ""`, so any proposal built with an empty executor
description hid the realm whose code runs on passage. That is 16 call sites
across 7 production realms (`r/sys/users` ×6, `r/sys/namereg/v1` ×4,
`r/sys/params` ×2, and one each in `r/sys/names`, `r/sys/validators/v2`,
`r/sys/validators/v3`, `r/gnops/valopers/proposal`), plus filetests.
Separately, `NewUpgradeDaoImplRequest` — the proposal that rewrites
`allowedDAOs` itself — passed an empty description, so the realm being granted
that authority appeared nowhere on the proposal page.

## Decision

### 1. Ignore an empty `AllowedDAOs` in `UpdateImpl`

```go
if len(r.AllowedDAOs) != 0 {
	// ... every entry checked non-blank and unpadded ...
	allowedDAOs = r.AllowedDAOs
}
```

`UpdateImpl` is the only assignment site, so this makes the transition
`empty -> non-empty` **one-way**. The bootstrap window can be closed once and
never reopened through the proposal path.

No defensive copy is taken. That was decided twice, the second time after the
copy had been put back on a mistaken reading — the full account is under
Alternatives. The short version: no outside realm can retain a handle to the
stored slice. The VM refuses to allocate another realm's struct type, so the
request cannot be built as a literal, and it refuses a write through a request
returned by `NewUpdateRequest`, which copies anyway. Both were checked from a
separate realm rather than reasoned about.

`len() != 0` is necessary but not sufficient, so entries are validated too.
`InAllowedDAOs` compares by exact string and a user realm's `PkgPath()` is `""`,
so a single `""` entry admits any caller whose previous frame is a user realm —
the same fail-open outcome, spelled differently. `NewUpgradeDaoImplRequest`
passes its `realmPkg` argument straight into the list, so an empty one reaches
`UpdateImpl`. Blank entries are therefore rejected outright rather than dropped
silently, and `NewUpgradeDaoImplRequest` rejects a blank `realmPkg` at
construction so the mistake surfaces to the proposal author rather than at
execution. (This is latent rather than live today: none of the five gated
functions are `MsgCall`-encodable, so an `""` entry currently grants nothing.
It is one signature change from mattering.)

That last point was checked against the keeper rather than assumed, because it
is what makes the blank-entry case latent instead of live. Two rules block it,
and each of the five is blocked by one of them:

- `keeper.go:818` requires a callable function's first parameter to be
  `.uverse.realm`. `memberstore.Get(_ int, rlm realm)` takes an `int` first, so
  it is rejected outright. `SafeExecutor.Execute` is a method, and the lookup
  resolves package-level names only.
- `convert.go` turns string arguments into Gno values and handles primitives,
  arrays, and `[]byte` alone. Everything else panics with "unexpected type in
  contract arg". That stops `UpdateImpl` (a struct), `treasury.Send` (an
  interface), and `treasury.SetTokenKeys`, whose `[]string` fails the explicit
  `Elt == Uint8Type` test with "unexpected slice type in contract arg".

So a `""` entry grants nothing today. What would change that is any of the five
gaining a signature MsgCall can encode — not a change to this realm at all,
which is the reason the guard is worth having now.

The same comparison creates the opposite failure, and entries are checked for
it too. Because `InAllowedDAOs` matches whole strings and entries are stored
exactly as given, an entry with a space around it — `" gno.land/r/x "` — is
accepted by any non-blank test and then matches nobody. The list is still
non-empty, so the bootstrap window is shut as well. A proposal whose entries
were all padded would lock the DAO out of its own allowlist: no further
`UpdateImpl`, no treasury movement, no member-store change, and no way back
short of a genesis fix. Padded entries are rejected rather than trimmed, for
the same reason blank ones are — trimming would store a list the proposal did
not describe. `NewUpgradeDaoImplRequest` rejects a padded `realmPkg` at
construction as well.

The padded-entry brick is reachable, but not through the built-in path: that
callback always includes `gno.land/r/gov/dao/v3/impl` alongside `realmPkg`, so
a padded `realmPkg` alone leaves a working entry behind. Bricking needs a
bespoke executor that calls `UpdateImpl` with a list of its own. The check sits
in `UpdateImpl` because that is the single point every path goes through.

**This check does not make the allowlist typo-proof, and should not be read
that way.** Any wrong path bricks the DAO identically — `.../v3/impll` for
`.../v3/impl` passes every guard here and matches nobody, which was confirmed
by probe. No check at this layer can distinguish a typo from a realm that has
not been deployed yet, so none is attempted. Whitespace is singled out for one
reason: it is the only spelling a human reviewing the proposal cannot see. The
general case stays a review responsibility, and it is the reason the lost
escape hatch above matters.

Legitimate flows are unaffected, and the genesis flow in particular is
bit-for-bit equivalent:

- `v3/loader/loader.gno:37` passes `nil`, but it runs in `init()` when
  `allowedDAOs` is *already* empty. Previously that stored a non-nil empty
  slice; now it is ignored and the var stays nil. Both are length zero, so
  `InAllowedDAOs` behaves identically and the bootstrap window still opens —
  which is exactly what the loader's comment asks for ("AllowedDAOs is
  intentionally left empty so that the genesis MsgRun can manipulate the
  memberstore directly").
- The lockdown `MsgRun` — section 7 of `govdao_prop1.gno`, generated at the
  script's step 3 of 7 — passes a **non-empty** list, which the fix still
  stores.
- An implementation-only upgrade still swaps `r.DAO`; that assignment is
  untouched.
- `NewUpgradeDaoImplRequest` passes a two-element list naming both the incumbent
  and the new impl, so handing off to a fresh implementation still works.

### 2. Nil guard in `PreExecuteProposal`

Return `false, errors.New("proposal not found")` instead of dereferencing,
mirroring `VoteOnProposal` (`v3/impl/govdao.gno`), which already returns this
error for the same lookup. (`ExecuteProposal` *panics* with the same string; it
is not the precedent to cite.)

Scoped honestly: this does **not** make such a proposal resolvable.
`dao.executeProposal` panics on *any* error from `PreExecuteProposal`
(inside `executeProposal` in `proxy.gno`), ahead of its `execErrorRejects` branch. The gain is that the
caller sees `proposal not found` rather than
`runtime error: nil pointer dereference`, which reads like a VM fault. Making
the proposal rejectable means changing `executeProposal`, which is a proxy-level
behaviour change and out of scope here.

### 3. Disclose the executor's creation realm unconditionally — and escape it

Split the disclosure out of the description block in `render.gno` (and out of
the matching gate in the exported `StringifyProposal`) so
`Executor created in:` renders whenever it is non-empty.

The value is **not** inherently trustworthy, and an earlier draft of this change
said it was. `Executor` is a public interface and `CreationRealm()` is
dispatched through it, so only `SimpleExecutor` supplies a VM-captured
`rlm.PkgPath()` (guarded by `IsCurrent()`); a third-party executor returns any
string it likes. Un-gating the line therefore *widened* an existing injection —
it now renders on every proposal — so it is wrapped in `sanitize.InlineCode`.
That both neutralizes it (a code span carries no markdown, and the fence widens
if the value contains backticks) and suits a realm path, so the ordinary case
reads better than it did unescaped.

### 4. Give the upgrade proposal a real description

`NewUpgradeDaoImplRequest` now states the grant in words and names the realm.
The realm path is caller-supplied and executor strings are emitted raw at the
render site, so it is sanitized at composition time with
`sanitize.InlineCode`.

`md.EscapeText` (an alias for `InlineText`) was wrong for this slot: the value
sits inside a code span, and CommonMark §6.1 does not process backslash escapes
there, so its output rendered with visible backslashes on every upgrade
proposal — and a backtick in the path closed the span early and freed the rest
as live markdown. `InlineCode` widens the fence instead, and emits the fence
itself, so the surrounding backticks were removed.

The disclosure has a limit worth stating: the payload is `newDao`, an arbitrary
`dao.DAO` value, and nothing ties `realmPkg` to it. A proposer can name a benign
realm while installing a different implementation. The text therefore states the
grant, not the identity of the incoming DAO, and says so.

### 5. Clamp attacker-controlled strings before sanitizing

Sanitizing is expensive per byte — measured on this tree, `InlineCode` costs
~11,310 gas/byte and `InlineText` ~6,990, against ~31 for the raw concatenation
they replace. Rendering is reachable unauthenticated through `vm/qrender`, which
runs under `maxGasQuery = 3_000_000_000` (`gno.land/pkg/sdk/vm/keeper.go`).

How those two rates were obtained, because an earlier draft of this ADR got
them wrong. Run a throwaway filetest that calls the helper on an input of
100 bytes, then again at 10,100 bytes, and subtract the reported gas. VM gas is
deterministic, so this is exact and repeatable rather than a timing benchmark:

| helper | 100 B | 10,100 B | per byte |
|---|---|---|---|
| `InlineCode` | 1,593,411 | 114,710,657 | 11,312 |
| `InlineText` | 1,107,707 | 70,971,091 | 6,986 |

The raw-emission figure was measured the same way, but through a whole render
rather than a bare helper call — an executor whose `String()` returns 100 bytes
against one returning 250,100, both emitted unescaped: 12,510,964 against
20,316,954, giving **31 gas/byte**. A first pass at this differenced against a
different probe's baseline and produced 6, which is why both points come from
the same probe here.

`InlineText` was also measured on an all-punctuation input, which escapes and
so doubles in length: 1,190,264 and 71,083,311, giving 6,989 per byte. The rate
does not depend on content.

What *cannot* be used is the pair of whole-page figures below. Dividing their
difference by 250,000 gives ~12,066 and looks like a per-byte rate for the
sanitizer, but the larger render also pays to build, format and store a 250 KB
string, so the result over-attributes everything to escaping. An earlier pass
did exactly that and wrote ~12,100 into three files.

`ExecutorCreationRealm()` is dispatched through the public `Executor`
interface, so a hostile executor *computes* it on every call and stores almost
nothing on chain. Left unbounded, that is a denial of service that this change
would have introduced, while the executor object itself occupies a few hundred
bytes. Measured end to end on the render path, unclamped:

| executor returns | gas for one render | vs the 3,000,000,000 cap |
|---|---|---|
| 250 KB | 2,839,117,770 | under, page still renders |
| 265 KB | 3,008,818,394 | over, page cannot be rendered |
| 300 KB | 3,404,598,931 | well over |

An earlier draft of this ADR put 250 KB at 3,036,543,407 and called it past the
cap. That number was about 7% high — the same overshoot as the per-byte rate it
was derived alongside — and 250 KB actually lands just under. The correction
does not change the conclusion. The attacker picks the size, computes it fresh
on every call and stores almost nothing, so moving from 250 KB to 265 KB costs
them nothing and takes the page past the cap for everyone. It does change what
the number proves: the cap is crossed at roughly 265 KB, not at 250 KB.

The same shape applies to `DeniedReason`, less severely because it is stored, so
the attacker pays for the bytes once.

Both are therefore clamped (`clamp.gno`) *before* they reach the sanitizer, so
its cost is bounded by a constant instead of by attacker input. The 250 KB
payload then renders in **18,749,984 gas**. `clampField` cuts on a rune boundary
and marks the result, and the marker deliberately avoids markdown punctuation
because these values are escaped downstream.

The clamp has to come before **every** pass over the value, not just before the
escaping, and the first cut of this change got that wrong. It read
`clampField(strings.TrimSpace(value), max)` — trim first, then cut — and
`TrimSpace` walks every byte it is handed, so the clamp never got the chance to
bound anything. An attacker routes straight around it with whitespace, which
`TrimSpace` cannot skip:

| creation realm | gas for one render | vs the cap |
|---|---|---|
| 250 KB of spaces | 1,368,719,824 | 46%, and renders *nothing* |
| 560 KB of spaces | 3,048,904,520 | over — page unrenderable |
| 560 KB, after the fix | 16,657,792 | 0.6% |

That is the same denial of service the clamp exists to prevent, reached through
the one pass that ran ahead of it. Nothing caught it: the whitespace case in the
filetest returned five characters, and the large case returned `z`s, which
`TrimSpace` skips in constant time. The gap needed both properties at once.

Both call sites now read `sanitize.X(strings.TrimSpace(clampField(value, max)))`,
and the filetest covers a large whitespace value. One visible effect: such a
value now shows a truncation marker instead of nothing, because the marker
survives the trim — which is also what makes the ordering testable.

Not addressed here: `ExecutorString()` is equally attacker-computed and is still
emitted raw. It is ~31 gas/byte rather than ~11,310, so reaching the cap needs
orders of magnitude more data, and bounding it would truncate legitimate
executor descriptions across `examples/`. Worth its own change.

### 6. Clamp the proposal title, and the stored denial reason

Two more unbounded paths were found during the audit and are fixed here. Both
were reported first as known issues and then fixed on request; the measurements
below are before and after.

**The proposal title.** `render.gno` passes `p.Title()` through `md.EscapeText`,
which is `sanitize.InlineText` — ~6,990 gas a byte — on both the proposal page
and the list page, and nothing limits title length at creation. The list page is
the one that matters: it escapes one title per proposal shown, so a handful of
oversized titles priced the page every visitor loads out of the query cap.

| what | before | after |
|---|---|---|
| list page, five 90 KB titles | 3,172,507,361 — over the cap | 66,871,916 |
| proposal page, one 250 KB title | 1,758,132,515 | 22,477,986 |

`maxRenderedTitle` is 512 bytes. The longest title in `examples/` is about 40,
so this is roughly nine times anything real and long for a heading; no golden in
the tree moved when it was added. As everywhere else here, the clamp runs
*before* the escaping.

**The stored denial reason.** `DeniedReason` was clamped when rendered but not
when written, so the realm still stored the executor's full error. A 250 KB
error grew `r/gov/dao/v3/impl` by 250,133 bytes in one execution — about 25 GNOT
of deposit, charged to whoever executed the proposal rather than to whoever
wrote the executor, and never visible because the render clamps. It is now
bounded on the way in as well, to the same `maxRenderedReason`, so the realm
never holds a reason it cannot show: the same execution now grows storage by
1,165 bytes.

The render-site clamp stays. Reasons stored before this change are still
unbounded, and the page has to survive them.

The cost of that write did not land on the attacker, which is why it was worth
fixing rather than tolerating. `PreExecuteProposal` gates on `isValidCall` only
— there is no membership check — so **any** account may execute a proposal that
has passed, and `VMKeeper.Call` charges the storage deposit to `msg.Caller`, the
account that sent the transaction. The charge is capped by `msg.MaxDeposit`, or
by the chain's `DefaultDeposit` of 600,000,000 ugnot when the sender sets none —
about 6 MB at the default price, so a 250 KB error was paid silently rather than
refused. A sender who lowers `MaxDeposit` is not charged, but then the
transaction fails and the proposal cannot be finalized by anyone unwilling to
pay. Either way the attacker paid nothing.


Testing the write needed a package test rather than a filetest. Both bounds are
1024, so a page rendered from an unclamped store looks exactly like one rendered
from a clamped store — asserting through `Render` would pass either way.
`z_denied_reason_store_test.gno` reads the stored value directly.

**One escape site is deliberately left unclamped**, and the asymmetry is worth
stating because a reviewer scanning for consistency will find it.
`prop_requests.gno` wraps `realmPkg` in `sanitize.InlineCode` when it builds the
upgrade proposal's grant sentence, with no bound. Clamping it would be wrong,
not merely unnecessary: the same `realmPkg` goes into the allowlist verbatim
(`[]string{"gno.land/r/gov/dao/v3/impl", realmPkg}`), so a truncated disclosure
would name a shorter path than the one actually being granted authority. The
sentence would lie about what the proposal does, which is worse than the cost it
would save.

The cost is bounded anyway, by a different mechanism. `NewUpgradeDaoImplRequest`
is only reached when a proposal is built, which is a transaction: an oversized
`realmPkg` runs out of the transaction's gas and the proposal is never created,
rather than persisting and pricing a query. Its output reaches a render only
through `ExecutorString()`, which is emitted raw at ~31 gas/byte.

Checked while auditing this: the list page puts the title inside markdown link
text, `[title](path)`, so an unescaped `]` would let a title hijack the link.
`EscapeInline`'s set includes `[`, `]`, `(` and `)`, so `](evil)` renders as
literal text. That held before this change; the clamp does not weaken it.

## 6. Sanitize `DeniedReason`

`DeniedReason` is `"execution failed: " + err.Error()`, and that error comes
from the proposal's executor callback — third-party code. Rendered raw it wrote
attacker-controlled markdown directly beneath the `PROPOSAL HAS BEEN DENIED`
line, including a forged heading and vote tally.

We use `sanitize.InlineText`. Two earlier attempts were wrong and are recorded
here because the reasoning is the interesting part:

- `strings.TrimSpace(sanitize.Block(...))` **defeats the sanitizer.** `Block`'s
  own docstring states that its `"\n\n"` envelope is the only thing bounding
  CommonMark §4.6 HTML block types 6 and 7 (`<div>`, `<table>`), which it does
  not escape in any mode. `TrimSpace` deletes exactly that envelope, and the
  tally is appended with no blank line — so a reason whose second line opens a
  `<div>` swallows the vote tally.
- Plain `Block` also leaves `**bold**` live, so a denial reason could still
  render a bold `PROPOSAL HAS BEEN ACCEPTED`.

`InlineText` is the right helper because this is an inline slot on the
`REASON:` line: it folds newlines to spaces so the reason cannot leave its line
at all, and it escapes emphasis and raw HTML. The objection that it renders
`Boom!` as `Boom\!` was an artifact of reading the raw markdown — a
backslash-escaped punctuation mark displays as the bare character.

## Alternatives considered

**Make `NewUpdateRequest` return a nil slice for nil input.** Fixes the
`v3/loader` path but not the explicit `AllowedDAOs: []string{}` literal, and
leaves the fail-open state reachable from any direct struct construction. The
guard belongs at the single writer, not at one of its callers.

**Also store a defensive copy of the caller's slice.** Implemented, then
removed. The concern was that a caller could retain a handle to the stored
allowlist and rewrite it without passing through `UpdateImpl`. The VM already
makes that impossible, verified in a filetest realm calling into `r/gov/dao`:
constructing another realm's struct fails with `cannot allocate
gno.land/r/gov/dao.UpdateRequest in realm ...`, and an element write through a
retained slice fails with `cannot directly modify readonly tainted object (use
a method or crossing function): r.AllowedDAOs[0]`. The only reachable
constructor, `NewUpdateRequest`, already copies — and `r/gov/dao/types.gno:280-286` documents
why. A second copy would have guarded an unreachable write, and its test could
only have exercised a same-package write no external caller can perform.

This was re-litigated once during the audit and got the wrong answer: the copy
was put back, with a test and a commit message calling the aliasing a live
TOCTOU hole. Re-running the filetest probe settled it — an outside realm hits
`cannot allocate gno.land/r/gov/dao.UpdateRequest in realm ...` on the literal,
and `cannot directly modify readonly tainted object` on a request obtained from
`NewUpdateRequest`. The copy and its test were removed again. The reason the
mistake was easy to make is worth recording: the test that "proved" the hole
lived in package `dao`, where building the literal is legal, so it passed
without ever exercising a path an attacker has.

**Make `InAllowedDAOs` fail closed.** The correct end state, but it breaks
genesis: the bootstrap `MsgRun` that seeds the member set would have nothing to
authorize it. That is a genesis-flow redesign, not a surgical fix.

**Panic on an empty `AllowedDAOs` instead of ignoring it.** Rejected: it would
abort implementation-only upgrades that pass `nil` deliberately — including the
`v3/loader` form — turning a silent bug into a loud regression on a path that is
otherwise correct.

**Escape `p.Description()` at the render site.** Rejected after checking the
call sites: descriptions are *deliberate* markdown (e.g. `NewAddMemberRequest`
builds a `####` portfolio heading), so escaping them globally would visibly
break roughly ten production realms. Only the two attacker-controlled slots —
`DeniedReason` and the upgrade proposal's realm path — are sanitized. Broad
description sanitization needs its own change with golden updates across
`examples/`.

## Consequences

### Positive

- The fail-open bootstrap state cannot be re-entered after lockdown.
- Proposals unknown to the current implementation report a named error instead
  of an opaque nil dereference. They still cannot be resolved — see Decision 2.
  Making them rejectable needs a proxy-level change that is out of scope here.
- Every proposal now discloses the realm whose code executes on passage, and
  upgrade proposals state the authority being granted.
- Two attacker-controlled strings can no longer forge page structure.

### Negative

- Governance can no longer deliberately return the DAO to the fail-open state
  through a proposal. This is the intent, but it does remove an escape hatch: a
  chain that somehow locks itself out now needs a genesis-level fix rather than
  a proposal. Given that the fail-open state grants treasury access to every
  realm on the chain, an unusable escape hatch is preferable to a reachable one.
- Golden output grows by the `Executor created in:` line for proposals that
  previously hid it (6 filetest files added or updated).

### Neutral

- `sanitize/v0` becomes a dependency of `r/gov/dao/v3/impl`.
- `storage_deposit_price_change.txtar:37` asserts a prefix of a balance that
  shifts whenever the packages it loads grow, and adding production code to
  `r/gov/dao` and `v3/impl` does grow them. It has been revised this way twice
  before by unrelated changes (`9999854` -> `9999853`, `9999817` -> `9999816`).
  This branch does not change it: the assertion currently reads `9999815`, and
  the code added here does not move the balance far enough to fall below that.
  Verified by running the txtar against this branch — it passes untouched. The
  file is worth knowing about because the next change to these realms may well
  have to revise it.
- `govdao_execute_reject_proposal.txtar` asserts the denial reason verbatim, so
  it is updated: `InlineText` backslash-escapes the `!` in `Boom!`. The rendered
  page is unchanged — a backslash-escaped punctuation mark displays as the bare
  character.

## Known-sound (audited, deliberately not changed)

- The genesis bootstrap window itself. Lockdown is a genesis transaction; the
  window is never open on a live chain.
- `isValidCall` in `v3/impl/govdao.gno` — correctly pairs `cur.IsCurrent()`
  with the pkgpath check, and then accepts a caller by one of two routes:
  `prev.IsUser()` for a direct `MsgCall` from an account, or, for `MsgRun`,
  `chain.PackageAddress(prev.PkgPath()) == unsafe.OriginCaller()`, which admits
  only the origin's own ephemeral run realm. Both branches matter — reading the
  first alone suggests `MsgRun` cannot reach these functions at all, and it can.
  What neither branch checks is membership, which is why any account can execute
  a proposal that has already passed.
- `UpdateImpl` and `SafeExecutor.Execute` read `cur.Previous()` without calling
  `cur.IsCurrent()` first, and that is correct here even though `AGENTS.md`
  states the check as mandatory. Both take `realm` as their **first** parameter,
  which makes them crossing functions, and the VM refuses any first argument
  other than `cur` or `cross(rlm)` — a stashed realm value fails with "only
  `cur` or `cross(rlm)` are allowed as the first argument to a crossing
  function", confirmed by probe. `NewSimpleExecutor` is the contrast: its realm
  sits in the second position, so it is not a crossing function, nothing is
  enforced for it, and it carries the `IsCurrent()` check it needs. Noted
  because applying the rule mechanically flags the first two as bugs.
- Proposal descriptions as markdown — deliberate, see Alternatives.
- The per-proposal allowlist snapshot. `CreateProposal` stores `allowedDAOs[:]`
  in each proposal, which shares a backing array with the
  package variable. Nothing authorizes against it: the only reader is
  `Proposal.AllowedDAOs()` (`r/gov/dao/types.gno:124`), it copies on the way out, and no
  caller exists anywhere in the tree. `UpdateImpl` replaces the slice wholesale
  rather than writing into the existing backing array, so a snapshot cannot
  change after the proposal was created.

## Known gap in this patch's own coverage

Re-gating the disclosure inside `StringifyProposal` (as opposed to
`render.gno`) survives the suite: the only caller is a filetest whose executor
*has* a description, which is the case that was never broken. `StringifyProposal`
has no production caller, so this is a coverage gap in dead code rather than an
untested production path — recorded so the mutation table above is not read as
stronger than it is.

## Found but deliberately not fixed here

- **`Render` echoes the attacker-supplied path into the page.**
  All three of `renderProposalPage`, `renderProposalListItem` and
  `renderVotesForProposal` in `render.gno` emit `strconv.ParseInt`'s error raw, so
  `/r/gov/dao:v3/impl` with a crafted path renders attacker inline markdown —
  including live links — inside a govDAO-branded error page. Go quotes the
  string, so no block structure can be forged and raw HTML is dropped by
  gnoweb; the residue is inline emphasis and `<a href>`. Pre-existing, on a path
  this patch does not otherwise touch, but `AGENTS.md`'s `Render(path)` rule
  covers it and it should be fixed.

- **A retired implementation keeps its authority.** `NewUpgradeDaoImplRequest`
  always retains `gno.land/r/gov/dao/v3/impl` in the list, so after a handoff to
  v4 the replaced realm can still mutate the member store and move treasury
  funds — and `impl.AddMember` is a public entrypoint. This is existing
  behaviour and arguably deliberate (it is the recovery path if the new
  implementation is broken), but it means an "upgrade" is an addition, not a
  handoff.
- **`treasury.SetTokenKeys` reports the wrong operation.** Its guard
  (`treasury.gno:64-65`) panics with `"this Realm is not allowed to send
  payment: ..."`, copied from `Send` (`treasury.gno:76-77`). Operator-visible text, not a security property; changing
  it is a separate, assertion-breaking change.
- **`StringifyProposal` renders `p.Title()` unescaped** while `render.gno`
  escapes it. The function has no production caller today.

## Out of scope

- Global description sanitization across `examples/`.
- `misc/deployments/gnoland1/govdao_prop1.gno` (does not currently build).
- The dead `memberstore.NewChangeTiersRequest` path and the unused
  `Proposal.allowedDAOs` field.

## Testing

- `examples/gno.land/r/gov/dao/allowlist_test.gno` (new) drives the real
  `UpdateImpl` through `testing.NewCodeRealm`, covering both reopen paths
  (`NewUpdateRequest(d, nil)` and an explicit empty literal), a legitimate
  extension, and that the genesis bootstrap window still opens. Verified to
  fail against unfixed code.
- `examples/gno.land/r/gov/dao/v3/impl/denied_reason_test.gno` (new) asserts an
  injected heading, list item and rule cannot survive into the stats block, and
  that plain text renders verbatim. Verified to fail against unfixed code
  (every injection assertion fires).
- `examples/gno.land/r/gov/dao/v3/impl/filetests/executor_disclosure_filetest.gno`
  (new) covers the disclosure changes in an isolated realm — the `impl` package's
  unit tests share proposal ids across files and swap the DAO implementation
  partway through, so they are the wrong home for these. 16 assertions, all goldened `true`.
- `examples/gno.land/r/gov/dao/v3/treasury/test/treasury_test.gno` gains
  `TestTreasuryLockdownCannotBeReopened` — the only test that proves the fix
  protects *funds* rather than a boolean. With the fix reverted, the
  non-allowlisted realm clears authorization and reaches `banker not found`.
- `proxy_test.gno` cleanup updated: its old reset relied on `UpdateImpl`
  accepting an empty list, which is now a no-op by design.

Every production change was mutation-tested against a verified-clean baseline
(an unmutated hard-linked copy must report zero failures first — an earlier
sweep produced false kills from a stale golden, and two more from mutations that
merely failed to compile).

A later sweep produced one more, and the way it slipped through is worth
recording. Screening on the harness's `FAIL: 0 build errors, N test errors`
line does **not** separate a real failure from a compile error: a Gno type error
inside the package under test is counted as a *test* error, so a mutation that
does not compile still reports zero build errors and reads as a clean kill. The
"drop `PreExecuteProposal` nil guard" row was screened that way and was wrong —
the pattern `if status == nil {` matched `VoteOnProposal`'s guard first, so the
sweep deleted a different function's check and got `declared and not used: tie`.
Re-run against the right block, it is a genuine kill, and the abort value is
exactly the `runtime error: nil pointer dereference` this ADR quotes.

The whole table was then re-run under two rules: screen on `gnoTypeCheckError`
rather than the build-error count, and refuse to apply a mutation whose anchor
text does not appear exactly once, so it cannot land on a same-shaped line in
another function. That turned up a second bad row. "Emit the creation realm raw"
removes the only `sanitize.` call in `render.gno`, which leaves the import
unused and stops the file compiling; the mutation has to drop the import too.
With that fixed it is a genuine kill.

Both bad rows failed the same way — the mutation never ran, and a compile error
was read as a caught bug. Every row in the table has now been re-run under those
two rules and every one is killed.

The two rows that name more than one package were also checked one package at a
time, since "killed by A, B" claims more than running A and B together shows.
Both hold: for the `!= nil` row, `r/gov/dao` and `r/gov/dao/v3/treasury/test`
each fail on their own, and for the re-gating row, `r/sys/namereg/v1` and
`r/gov/dao/v3/impl` each fail on their own.

Two of these rows were added after the mutation survived. Swapping the clamp and
the escaping — `clampField(InlineCode(x))` instead of `InlineCode(clampField(x))`
— passed the whole suite on both paths. The tests that were supposed to cover it
asserted only that the output was short and carried the truncation marker, and
both of those stay true when the order is wrong: the escaper still processes the
full attacker-supplied string, and the cut lands inside the escaped text. The
creation-realm case now checks that the marker sits inside the closing fence,
and the `DeniedReason` case now uses a punctuation payload so escaping changes
the length. Worth knowing because a test named for a property is not evidence
that it checks it.

| Mutation | Killed by |
|---|---|
| `len(r.AllowedDAOs) != 0` → `!= nil` | `r/gov/dao`, `r/gov/dao/v3/treasury/test` |
| accept blank `AllowedDAOs` entries | `r/gov/dao` |
| accept padded `AllowedDAOs` entries | `r/gov/dao` |
| accept a padded `realmPkg` (`prop_requests.gno`) | `r/gov/dao/v3/impl` |
| guard the raw creation realm instead of the sanitized one (`render.gno`) | `r/gov/dao/v3/impl` |
| drop `PreExecuteProposal` nil guard | `r/gov/dao/v3/impl` |
| drop `renderProposalPage` nil guard | `r/gov/dao/v3/impl` |
| re-gate the creation-realm disclosure | `r/sys/namereg/v1`, `r/gov/dao/v3/impl` |
| emit the creation realm raw | `r/gov/dao/v3/impl` |
| `DeniedReason` raw | `r/gov/dao/v3/impl` |
| upgrade description back to `""` | `r/gov/dao/v3/impl` |
| `realmPkg` via `md.EscapeText` in backticks | `r/gov/dao/v3/impl` |
| remove the clamp on the creation realm | `r/gov/dao/v3/impl` |
| clamp outside the escaping — creation realm | `r/gov/dao/v3/impl` |
| clamp outside the escaping — `DeniedReason` | `r/gov/dao/v3/impl` |
| drop `TrimSpace` on the creation realm (`render.gno`) | `r/gov/dao/v3/impl` |
| remove the clamp on `DeniedReason` | `r/gov/dao/v3/impl` |
| clamp ignoring the rune boundary | `r/gov/dao/v3/impl` |
| accept a blank `realmPkg` | `r/gov/dao/v3/impl` |

Every row above was re-run against the final tree, after the work was split into
its three commits, and every one was killed. Each reported `0 build errors`, so
none is a compile failure being mistaken for a passing check — the trap the
preamble warns about. Re-running the whole table after the split was worth the
time on its own: it confirms the rebuild did not drop a guard while moving code
between commits.

The re-run also removed a row. The table listed the `TrimSpace` mutation twice,
once naming the creation realm and once naming "the disclosure guard". There are
two sites with that code — `render.gno` and the `StringifyProposal` twin in
`types.gno` — so the second row looked like separate coverage when it was not.
One row now, naming the file it covers.

Finding that duplicate exposed a real gap: at the time, none of the three
mutations on the `StringifyProposal` copy were caught, because its filetest only
ever passed it a benign realm path. An untested copy of a security-relevant
expression is free to drift from the tested one, so
`stringify_proposal_00_filetest.gno` now feeds it a hostile creation realm —
padded with spaces *and* carrying a run of two backticks, which is what lets one
payload cover two properties at once:

| mutation on the `StringifyProposal` copy | caught now |
|---|---|
| emit the creation realm raw | yes |
| drop `TrimSpace` | yes |
| remove the clamp | no |

The clamp is not covered because the payload is well under 256 bytes. That one
is left alone deliberately: the clamp is a gas bound, and `StringifyProposal`
has no production caller, so there is no live gas path behind it to protect. The
escaping and the trim are different — they decide what the function *returns*,
and any realm that starts calling it inherits that.

## A note on how things are cited here

Citations into files this change edits name the function rather than a line
number. Line numbers in those files went stale four separate times while this
was being audited: any edit moves everything below it, so a number checked in
one pass can be wrong by the next. Two of them ended up pointing at a blank line
and at an unrelated function, which is worse than no citation at all. Files this
change does not touch still carry line numbers, and those stay right until
master moves them.

## References

- `docs/resources/gno-interrealm.md`
- `docs/resources/gno-ai-contract-review.md`
- `docs/resources/effective-gno.md` § Verifying inbound Coin payments
