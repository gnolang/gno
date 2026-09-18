# ADR: gnoweb serves `/u/<name>` only for something that exists on the chain

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
   already issued for the contributions list). With `r/sys/names` enabled,
   which `gnoland-1` has had since genesis, a package under `r/<name>/` exists
   only if `<name>` was a registered name owned by the deployer. On gnodev,
   where `r/sys/names` is never enabled and `r/sys/users` is often not
   loaded, it is the only proof of existence there is.
3. `r/sys/users.ResolveName("<name>")` answers `true` through a new
   `ClientAdapter.Eval` method (`vm/qeval`). `ResolveName` returns
   `(nil, false)` for an unknown or deleted name and `(data, false)` for a
   name left behind by a rename; only the current name of a live user is
   `true`, which is the rule `r/sys/names` itself applies before authorizing
   a deploy.

Anything else is a 404. Every failure of rule 3, including a chain without the
registry and any RPC error, counts as "no": a lookup that failed must not
fabricate a profile.

Before any query, `<name>` must match the registry's own name shape
(`^[a-z][a-z0-9]*([_-][a-z0-9]+)*$`, at most 64 bytes, mirrored from
`r/sys/users/store.gno`) unless it is an address. This closes `/u/foo/bar`,
which `vm/qpaths` treated as a sub-prefix (`p/moul/addrset` rendered as
"Gnome moul/addrset"), and guarantees that nothing reaching the `qeval`
expression can leave a Gno string literal.

The gate runs before the `/r/<name>/home` fetch, so an unknown name costs two
RPCs instead of three.

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
- A name left behind by a rename renders while it keeps packages and 404s once
  it does not, which is what the verifier does.
- `ListPaths` returns `[""]` for an empty result, so rule 2 counts parsed
  contributions, not raw paths, or it would never fire.
- The `vm/qeval` payload is rendered text, one value per line, and the
  `UserData` line carries its own `(false bool)`, so only the last line is read.
