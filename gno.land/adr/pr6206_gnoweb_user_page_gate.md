# ADR: `/u/` is one identity, served only for something that exists on the chain

## Status

Proposed (PR #6206).

## Context

`GET /u/<name>` rendered a full profile page ("Gnome <name>", a contributions
list, a home tab) for any path segment, with HTTP 200. The handler never asked
whether `<name>` was a user: it listed the packages under `@<name>` and fetched
`/r/<name>/home`, and both come back empty for a name nobody owns. Two TODOs
left by the PR that introduced the page (#4024) asked for the check; the
reviewer had described the expected behaviour there: no page for a username
that does not exist in `r/sys/users`, always a page for a valid address.

On `gnoland-1` this is not cosmetic. Names are identity, every `@mention` in a
realm links to `/u/`, the omnibar suggests namespaces as users, and every page
is indexable. A one-character lookalike (`/u/mou1`) rendered exactly like the
real `/u/moul` with zero contributions, which is also what a freshly registered
user looks like.

The same handler knew only the half of the identity that was typed into the
URL. `/u/moul` never printed the address that name belongs to, and
`/u/g1manfred47kzduec920z88wfr64ylksmdcedlf5` listed packages under
`@g1manfred...`, found none, and rendered the same plausible empty page the
gate above exists to remove: 200, "Gnome g1ma...dlf5", zero contributions, for
a user with hundreds. The "user home" button linked `../r/<Username>/home`
after `Username` had been shortened for display, so on every address page it
pointed at `../r/g1ma...dlf5/home`, a path that cannot exist. The two TODOs
from #4024 asked for both halves in one line: "username + gno address to be
used".

## Decision

`GetUserView` serves a page only when one of three facts holds, checked in
this order:

1. `<name>` is a bech32 address (`crypto.AddressFromBech32`, HRP `g`, 20
   bytes). An address is a namespace by construction.
2. The namespace already holds at least one package (`ListPaths("@<name>")`,
   already issued for the contributions list). This is not a proxy for
   registration: on `gnoland-1`, `gnops`, `jeronimoalbi`, `leon`, `mason`,
   `sunspirit` and `tests` hold packages while `ResolveName` answers false,
   because `r/sys/names` gates deploys from block 1 but does not retro-register
   what genesis shipped. Their pages are legitimate and must keep working. What
   the rule proves is that the namespace is real and has content, which is what
   a user page shows. It is also the only proof available on gnodev, where
   nobody registers a name.
3. `r/sys/users.ResolveAny("<name>")` resolves to a live user whose current
   name is `<name>`, through a new `ClientAdapter.Eval` method (`vm/qeval`).
   `ResolveAny` delegates to `ResolveName` for a name and to `ResolveAddress`
   for an address, and returns `nil` for an unknown or deleted one. A name left
   behind by a rename resolves to the *current* name, not to itself, so it is
   not the current name of a live user, which is the rule `r/sys/names` itself
   applies before authorizing a deploy.
4. One of this gnoweb's own aliases points at the path. The operator published
   it deliberately, so the gate must not 404 a URL gnoweb advertises itself:
   `/docs` maps to `/u/docs` in `DefaultAliases`, and `docs` is neither
   registered nor a namespace holding a package, so rules 1 to 3 all refuse it.

Anything else is a 404. A chain that does not deploy `r/sys/users` answers
"no", which is the gnodev case; a chain that could not be asked (timeout, node
error) surfaces that error through the handler's usual mapping, because a 404
published on a blip deletes a real user's page for as long as a crawler
remembers it.

Before any query, `<name>` must match the registry's own name shape unless it
is an address: `gnolang.Re_name`, which `r/sys/users/store.gno` states its own
rule mirrors, plus that file's 64-byte cap. Taking the shared expression rather
than copying it means gnoweb cannot start refusing names the chain has begun to
accept. This closes `/u/foo/bar`,
which `vm/qpaths` treated as a sub-prefix (`p/moul/addrset` rendered as
"Gnome moul/addrset"), and guarantees that nothing reaching the `qeval`
expression can leave a Gno string literal.

`ResolveAny` is chosen over `ResolveName` because it answers both questions the
page has in a single `qeval`: whether the user exists, which gates the page,
and the other half of the pair, which the page prints. The handler then keys
every lookup on the **namespace** the packages live under rather than on the
URL segment: an address the registry resolves deploys under its name, so
`/u/<address>` switches to it. A name keeps its own segment even when it
resolves, because it is already the namespace, and following a rename here
would hide the packages the old name still holds. `UserData` carries
`Namespace` for links, which is never elided and repairs the home button, and
`Address` for display.

The gate runs before the `/r/<name>/home` fetch, which is dropped for a name
that has no page. Per request: an unknown name stays at two RPCs (`qpaths` and
`qeval` replace `qrender` and `qpaths`), a namespace with packages stays at
two, and a registered user with no packages pays a third. The `qeval` is the
heavier of the two shapes, because the `qrender` it replaces returned before
building a machine when the package was absent.

## Alternatives considered

- **`IsNameTaken`** is one boolean and trivially parsed, but it is
  `nameStore.Has`: true for deleted users and for old aliases after a
  rename. It answers "would `RegisterUser` fail", not "is this a user".
- **Resolving first, listing second** is the intuitive order, but it breaks
  gnodev, where names are not registered, and adds an RPC to every request
  for a namespace that has packages.
- **Rendering with a "this name is not registered" notice** keeps the 200
  and the crawlable space; the status has to be a 404 for links, crawlers and
  mention rendering to behave.
- **A dedicated `ResolveUser` method on the adapter** would hide the
  expression, but there is one caller. A generic `Eval` is the primitive every
  future read of a system realm needs, and the handler owns the expression.
- **`ResolveName` plus a second call for the address** keeps the boolean the
  gate wants and reads a plain string back, but it is two round trips, and
  `.Name()`/`.Addr()` on the returned pointer panic when it is nil, so the
  nil case has to be excluded first. `ResolveAny` answers both in one.
- **Redirecting `/u/<address>` to `/u/<name>`** gives one canonical URL and
  would avoid indexing the pair twice, but it needs a redirect path out of a
  handler that returns `(status, view)`, and it has nothing to return for an
  address with no name. Both forms render instead.

## Consequences

- `/u/<unknown>` and `/u/<lookalike>` are 404. `@mention`s of a name that
  does not exist land on an error page instead of a fake profile.
- `/docs` (`DefaultAliases`, `/docs` → `/u/docs`) keeps its 200 through rule 4,
  rather than 404ing on `gnoland-1` until #5760 serves the embedded docs there.
  A gate that breaks a URL the same binary advertises is a worse trade than the
  empty profile it removes, and the rule generalizes: any operator alias
  pointing into `/u/` is served.
- A registered user with no packages keeps their page, at one extra `qeval`.
- On gnodev a developer who has deployed nothing under their name gets a 404
  where they used to get an empty profile; deploying one package restores it
  through rule 2.
- The 404 body is the shared `StatusErrorComponent` text, "Something went
  wrong", which is wrong for a name that simply is not registered. Every other
  404 in gnoweb says the same thing, so the copy is a separate change.
- `$help`, `$source` and `$state` under `/u/` are not gated. They already 404
  on chain for a name that owns nothing, and they route through
  `GetPackageView` before this code, so closing them is a separate change.
- A name left behind by a rename renders while it keeps packages and 404s once
  it does not, which is what the verifier does.
- `ListPaths` returns `[""]` for an empty result, so rule 2 counts parsed
  contributions, not raw paths, or it would never fire.
- The `vm/qeval` payload is rendered text, one value per line. `ResolveAny`
  returns `(*UserData, bool)`, and only the first line carries the pair, so
  that is the line read. The realm exports no string-returning resolver and
  `.Name()` on the pointer panics when it is nil, so the pair is matched out of
  the value repr, keyed on each field's *type tag* rather than its position: a
  field added to `UserData` does not shift the result. `(nil ...)` is a
  legitimate "no user"; any other unrecognized shape is an error, not a "no",
  because quietly 404ing every registered user at once must surface.
- `/u/<name>` and `/u/<address>` now serve the same page and print both halves,
  so the pair is indexable twice. No canonical link tag is emitted; if that
  matters, it is a separate change.
