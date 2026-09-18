# ADR: gnoweb serves `/u/<name>` only for something that exists on the chain

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
3. `r/sys/users.ResolveName("<name>")` answers `true` through a new
   `ClientAdapter.Eval` method (`vm/qeval`). `ResolveName` returns
   `(nil, false)` for an unknown or deleted name and `(data, false)` for a
   name left behind by a rename; only the current name of a live user is
   `true`, which is the rule `r/sys/names` itself applies before authorizing
   a deploy.

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

## Consequences

- `/u/<unknown>` and `/u/<lookalike>` are 404. `@mention`s of a name that
  does not exist land on an error page instead of a fake profile.
- `/docs` (`DefaultAliases`, `/docs` → `/u/docs`) becomes a 404 on
  `gnoland-1` until #5760 serves the embedded docs there; it was an empty
  fabricated profile before. On gnodev with `examples/` loaded,
  `r/docs/...` exists and the page still renders.
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
- The `vm/qeval` payload is rendered text, one value per line, and the
  `UserData` line carries its own `(false bool)`, so only the last line is
  read. A last line that is neither `(true bool)` nor `(false bool)` is an
  error, not a "no": if `ResolveName` ever changes shape, that must surface
  instead of quietly 404ing every registered user.
