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

gnoweb shows a notice on every `/r/`, `/p/` and `/u/` page whose namespace is
not under a trusted entry: render, `$source`, `$help`, `$state`, and the user
profile, which renders that user's home realm. The actions page is where a
user copies a transaction, so it is the page that matters most: when a
transaction link asks for coins there, its warning also says "This is a
community realm, deployed by its author." Markdown responses of these pages
(`Accept: text/markdown`) keep their body verbatim and carry the response
header `X-Gnoweb-Realm-Notice: community` instead.

The notice is the second row of the sticky header, under the path bar and the
tabs and above the header's bottom border. It uses the header's own surface,
grid and hairline, a circled info icon in the site's green, and secondary
text with the lead-in "Community realm" in semibold primary text, not green,
so it does not read as a link; the row has `role="note"` and the
accessible name "Community realm notice", and the main landmark points at its
text with `aria-describedby`, so a reader who skips the header still gets it.
Being part of the sticky header, it stays in view while the user scrolls a
long page or fills a transaction form. Official pages render exactly as
before, with no empty row. The site-wide banner and any network banner stay
above the header, unchanged: an operator's emergency banner must stay visible
on the very pages it warns about, and the notice never replaces it.

The row's height is fixed, so nothing below the header can drift. The server
sets `data-realm-notice-lines` on `<html>`: 1 for the default text, 2 for an
operator's. The row's block size derives from that count, the font size and
line height tokens, and sticky elements below the header (`--s-header-offset`)
and anchor jumps (`scroll-padding` on the root) move down by exactly that
size. The default text fits one line at every width from 320px: "**Community
realm.** Read the code first." up to `--lg`, and "**Community realm**,
deployed by its author. Read the code before you interact or send coins." from
`--lg` up; both are rendered and CSS shows one, without JavaScript. An
operator's `GNOWEB_REALM_NOTICE_TEXT` has no short variant: it shows at every
width, clamped to two lines, and long words such as URLs break rather than
widen the page. On viewports shorter than 30em the header does not stick, so
a taller header never takes over the screen (WCAG 1.4.10).

The notice is on by default in the `gnoweb` binary (`-no-realm-notice` turns
it off), with the trusted list in `-trusted-paths` and the wording in
`GNOWEB_REALM_NOTICE_TEXT`. The text is inline markdown, rendered like the
banner's but without images, so the sticky header never loads a remote image.
A text without a visible character refuses to start: empty markup, links
without a label, zero-width and other format characters, spaces and blank
fillers do not count. An on-by-default safeguard must not switch itself off
on a typo or be blanked on purpose, which is why it is stricter than the
opt-in banner. Entries of `-trusted-paths` that cannot match the package they
name (a domain, an `r/`, `p/` or `u/` prefix, uppercase letters) are logged at
startup and kept, trusting no intended package. The library default
(`NewDefaultAppConfig`) leaves the notice off, so gnodev, where every package
is the developer's own, never shows it.

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

The default list therefore assumes namespace enforcement as on mainnet
(`r/sys/names` enabled); a chain without it lets anyone deploy under any
name, and its gnoweb must set `-trusted-paths` itself.

An entry is a namespace or a package path, without the domain and without the
`/r/` or `/p/` prefix; one entry covers both trees because they share a deploy
key. An entry trusts its own path and everything under it, the most specific
entry wins, and the list is allow-only: a negative verdict on one package under
a trusted namespace means replacing the namespace entry with per-package
entries. A deny set is the later addition if that becomes frequent. This entry
format is the interchange format for any future producer of the list.

The default wording says the package was "deployed by its author" and claims
no review or audit. The same text is used on pure packages, which hold no
coins; it addresses the end user, who reaches a package page from a realm.

The default list names namespaces whose code the team reviews in this repo or
whose deploy key belongs to a party it vouches for. A vouched key covers
everything it deploys, experiments included: on mainnet today that is the
`moul/x/daily/*` and `gnoswap/*` realms, which have no counterpart in this
repo. If the team wants a narrower endorsement, the lever is to replace the
namespace entry with the paths it stands behind. Changing the list is a
reviewed pull request; the deployment flag exists for removing an entry
without a release.

"Trusted" vouches for code, not for what users write through it. A trusted
realm still renders user-written content, such as validator descriptions,
board posts or token names, without the notice.

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
  would not reach `$source`, `$help` or `$state` pages.

Presentation options weighed for the notice itself:

- **A full-width strip above the header**, in the warning or the info tone.
  Reads as a second site-wide banner, competes with the operator's banner,
  and scrolls away on long pages.
- **A callout at the top of the content column.** Scrolls away, and its
  position depends on each view's layout.
- **A "Community" tag in the path bar, disclosing the sentence on demand.**
  Hides the message behind an interaction and crowds the path bar on phones,
  next to the network controls.
- **A tinted header row.** Same structure as the decision, but a second
  surface colour in the header added weight without adding meaning.
- **A notice height left to the text.** Sticky elements below the header
  either overlapped a wrapped row or left a gap above a short one; the
  reserved line count keeps both exact.
- **A kind or tone field on the notice.** Only one kind exists. If another
  notice kind is ever needed, a kind on `RealmNotice`, mapped to a modifier
  class, is the extension point.

## Evolution

Two lanes, both already practised in this repository.

**Fast lane, in place with this change.** A list compiled into the binary and
changed by pull request is how gnoweb has always carried an allowlist: the
image hosts in the CSP (#4058) work the same way, and the wallet registry
proposed in #5970 takes the same shape. A merge builds the image; whether the
instance restarts on its own is an infrastructure property, not gnoweb's.

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
  separators and is fixed in #6261; `$state&oid=` still fetches
  any object regardless of the page's realm and is tracked separately.
