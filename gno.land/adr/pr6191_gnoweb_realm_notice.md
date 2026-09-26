# ADR: gnoweb community-realm notice on packages outside a trusted list

## Context

Third-party realms now deploy on mainnet, under `g1…` address namespaces or
`nym-…` registered names. gnoweb renders them with exactly the same chrome as
the team's own realms, so the site reads as an endorsement of whatever it
shows. With attention on the ecosystem growing, the team asked for a default
disclaimer on every package, lifted once a realm has been reviewed.

gnoweb already had a site-wide banner (`GNOWEB_BANNER_TEXT`), but no notion of
trust, per page or otherwise, and neither did the chain.

## Decision

gnoweb shows a notice, in the warning tone, on every `/r/`, `/p/` and `/u/`
page whose namespace is not under a trusted entry: render, `$source`, `$help`,
`?state`, and the user profile, which renders that user's home realm. The
actions page is where a user copies a transaction, so it is the page that
matters most; the notice sits at the top and scrolls out of view under the
sticky header on long pages, so a reminder next to the transaction form is a
possible follow-up. The notice is rendered under the site-wide banner, never
instead of it: an operator's emergency banner must stay visible on the very
pages it warns about.

The notice is on by default in the `gnoweb` binary (`-no-realm-notice` turns
it off), with the trusted list in `-trusted-paths` and the wording in
`GNOWEB_REALM_NOTICE_TEXT`. A text that renders to nothing refuses to start:
an on-by-default safeguard must not switch itself off on a typo, which is why
it is stricter than the opt-in banner. The library default
(`NewDefaultAppConfig`) leaves it off, so gnodev, where every package is the
developer's own, never shows it.

Trust is decided by namespace. That is sound on mainnet, as checked against
the live chain on 2026-09-17 (`r/sys/names.IsEnabled`, `r/sys/users`,
`r/sys/namereg/v0.ValidateNymFormat`):

- namespace enforcement is enabled: only the address that owns a name can
  deploy under it;
- the only controller of `r/sys/users` is the `r/sys/namereg/v0` realm, and
  its open registration only accepts names matching `nym-[a-z]{5,13}\d{3}`,
  so a plain name cannot be squatted. The default list holds three kinds of
  names: those preregistered at genesis to a party's key from this repo
  (`moul`, `aeddi`, `aib`, `howl`, `samcrew`, `onbloc`, `gnoswap`; see
  `misc/deployments/mainnet.gno.land/transactions/base/users-preregister/`),
  those the namereg seed registers to ownerless addresses (`gnoland`, `sys`,
  `gov`, `nt`, `demo`; see `r/sys/namereg/v0/preregister.gno`), and those
  not registered at all (`docs`, `tests`, `gnops`, `devrels`, `leon`,
  `jeronimoalbi`, `mason`). For the last two kinds only genesis or a GovDAO
  proposal can place code, and a GovDAO allocation of such a name to a new
  party has to be mirrored in the list.

An entry is a namespace or a package path, without the domain and without the
`/r/` or `/p/` prefix; one entry covers both trees because they share a deploy
key. An entry trusts its own path and everything under it, the most specific
entry wins, and the list is allow-only: a negative verdict on one package under
a trusted namespace means replacing the namespace entry with per-package
entries. A deny set is the later addition if that becomes frequent. This entry
format is the interchange format for any future producer of the list.

The wording says "reviewed", not "audited": the team reviews its own realms,
it does not audit them. The same text is used on pure packages, which hold no
coins; it addresses the end user, who reaches a package page from a realm.

The default list names namespaces whose code the team reviews in this repo or
whose deploy key belongs to a party it vouches for. A vouched key covers
everything it deploys, experiments included: on mainnet today that is the
`moul/x/daily/*` and `gnoswap/*` realms, which have no counterpart in this
repo. If the team wants a narrower endorsement, the lever is to replace the
namespace entry with the paths it stands behind. Changing the list is a
reviewed pull request; the deployment flag exists for removing an entry
without a release.

## Alternatives considered

- **Opt-in through an environment variable**, like the existing banner. Relies
  on operators remembering to set it and contradicts the "by default" ask.
- **A list fetched from GitHub over HTTPS.** gnoweb has no outbound call other
  than its RPC node and is meant to work without external dependencies or an
  internet connection (#2799); a served network registry was rejected for the
  same reason (#5584). Rejected outright.
- **A registry realm polled by gnoweb on a timer.** No precedent: every RPC in
  gnoweb is request-scoped and the one piece of chain-derived content, the
  home page, is synced by a sidecar script that writes a file (#4478). See
  Evolution.
- **A markdown alert inside the rendered article.** Needs renderer changes and
  would not reach `$source`, `$help` or `?state` pages.

## Evolution

Two lanes, both already practised in this repository.

**Fast lane, in place with this change.** A list compiled into the binary and
changed by pull request is how gnoweb has always carried an allowlist: the
image hosts in the CSP (#4058) work the same way, and the wallet registry
proposed in #5970 takes the same shape. A merge builds the image; whether the instance restarts on its
own is an infrastructure property, not gnoweb's.

**Slow lane, when a second reader or a governance need appears.** The source of
truth moves on chain under GovDAO, which is the trust root of every on-chain
list (`r/sys/users`, `r/sys/namereg`, `r/sys/params`): either the Auditor and
Audits Registry the Constitution asks GovDAO to run, or a curation realm with
a delegate slot that starts empty and is granted by GovDAO. The delegate may
add, may remove only what it added and may never empty the list, while GovDAO
keeps the whole-list reset; that is the `run_submitters` shape (#6088). No
timelock: none was ever adopted and the one cooldown was removed (#5767).

gnoweb consumes it the way it consumes the home page today: an external
syncer writes a file, one entry per line in the format above, and gnoweb reads
that file next to its compiled list, union of both, a local override winning,
exactly as `DefaultAliases` merges with `-aliases`. The failure of the sync
must be visible, which the home sync is not (#6121 §3.1), and the syncer must
be written against the deployed realm API, not against `examples/`.

**What the notice becomes.** The Constitution requires gnoweb to link
Qualified Audit Reports "but must not convey to the user any warrantees", and
reserves "audit" for vetted, accountable, non-automated auditors. So the
end state is disclosure, not endorsement: a package under a trusted entry
shows no notice; a package with a qualified report links to it; everything
else keeps the notice. A review pipeline, AI-assisted or not, can only propose
entries; a human lands them.

## Consequences

- Nothing needs reviewing for the protection to hold; only what the team
  chooses to endorse gets reviewed.
- `/` is aliased to `r/gnoland/home` and `/events` to `r/devrels/events`, so
  every default alias target must stay under a trusted entry.
- `sunspirit` is left out of the default list at the team's request.
- Reviewing this surfaced two pre-existing gaps that let third-party content
  render under a trusted path's chrome: `$source&file=` accepted path
  separators and is fixed alongside this change; `?state&oid=` still fetches
  any object regardless of the page's realm and is tracked separately.
