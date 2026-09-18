# ADR: gnodev imports the `devtest` account when asked

## Context

`gnodev` is the local dev-chain runner shipped under `contribs/gnodev/`.
On boot it premines a well-known account whose mnemonic is hard-coded in
the test suite as `integration.DefaultAccount_Seed` (re-exported by
gnodev as `DefaultDeployerSeed`). The seed and its derived address
(`g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`) are public — every
developer running gnodev gets the same key — and the chain genesis funds
this address with 10⁹ GNOT so transactions work out of the box.

Today, before this change, gnodev does **not** write the seed to the
user's gnokey keybase (`~/.gnokey/` by default). It only reads existing
keys, then logs the mnemonic and address in a `Warn` line so the user
can copy them. To actually sign a transaction with this account the
user has to leave the gnodev terminal, run `gnokey add -recover devtest`
(or whichever name they pick), paste the mnemonic at the prompt, and
come back. The mnemonic is public, the workflow is pure paperwork, and
new users routinely get stuck on it during onboarding.

## Decision

gnodev writes to the user's keybase only when asked. A plain boot logs
that the premined address cannot be signed for and names the two ways to
change that: press `I` while gnodev runs, or start it with
`-import-dev-key`. Either one calls `importDevKey` in
`setup_address_book.go`, which does:

1. If `home == ""`, log a warning and return (no keybase to write
   to; this only happens when `-home ""` is set explicitly).
2. If `home` is set but the directory does not exist:
   - If it is the default home (compared with `filepath.Clean` on both
     sides so a path-equivalent form such as a trailing slash still
     counts), create it with mode `0o700` so a requested import lands on
     a fresh install. This matches `gnokey add`'s behavior, which silently
     creates `~/.config/gno/` on first use.
   - Otherwise (an explicit `-home <path>` that doesn't exist, likely a
     typo), log a warning and return without writing. We refuse to
     silently materialize an arbitrary path on disk.
3. Open the keybase at `home` via `keys.NewKeyBaseFromDir`, which
   creates `home/data/` on disk if it does not exist (mode 0o700).
   That call panics rather than returning an error when it cannot create
   the dir (e.g. an unwritable home), so the open is wrapped to recover
   the panic and treat it as a normal failure.
4. If the deployer address is already in the keybase under any name, it
   is already signable: log `dev key already present in keybase,
   skipping` and stop. This is the key guard. The keybase enforces one
   name per address, so calling `CreateAccount("devtest", ...)` for an
   address already stored under another name (commonly `test1`) would
   silently delete that other name. Skipping preserves the user's
   existing entry.
5. Otherwise, if the name `devtest` belongs to a different address (the
   user has an unrelated key they named `devtest`), log a one-line warning and
   leave it untouched.
6. Otherwise import via
   `kb.CreateAccount("devtest", DefaultDeployerSeed, "", "", 0, 0)` and log
   `dev key imported`.

Every failure along the way (missing or unwritable home, locked or
corrupt keybase, failed import) degrades to a logged warning, never an
error. The import is a convenience, so a degraded keybase must never
stop gnodev from starting; the deployer address is still tracked
in-memory by `setupAddressBook`'s fallback when the import is skipped.

The mnemonic is read from the existing `DefaultDeployerSeed` constant;
no second copy is introduced. The startup no longer logs the mnemonic
at all: it prints `default address tracked in-memory only; gnokey cannot
sign with it` followed by the line naming `I` and `-import-dev-key`, and
`dev key imported` once the import runs. Users who need the mnemonic can
read `integration.DefaultAccount_Seed` in the source; `gnokey export
devtest` produces an armored, password-encrypted private key, not the seed
phrase.

A successful import through `I` also replaces the placeholder name gnodev
invented for that address in its own address book, so the `A` panel lists
one row, under `devtest`.

## Alternatives Considered

### 1. Import at every startup, with `-no-dev-key` to opt out

The first shape of this change wrote the key on boot and let the user
opt out. Rejected: gnodev is a dev server, and a dev server writing to
the user's key store unasked is a surprise, whatever the key is worth.
The keystroke costs one character, once per machine, and it is offered
on the line that explains why signing fails.

What the opt-out shape had going for it, and what replaces it:

- The mnemonic is already public, identical on every machine, and
  documented in gnodev's own output, so the import adds no secret. True
  either way: consent here is about the user's keybase, not the secret.
- The whole point of gnodev is "smallest possible loop from `make
  install` to signed transaction". `I` keeps that loop inside the
  terminal already running, with no second command and no prompt to
  answer before the node boots.
- CI runners and scripts, where no terminal is attached and interactive
  mode is off, pass `-import-dev-key` when they want the key.

### 2. Don't touch `~/.gnokey/`; let gnodev run its own keybase

gnodev would create `$XDG_STATE_HOME/gnodev/keys/` (or similar), import
`devtest` there, and tell users to run
`gnokey -home $XDG_STATE_HOME/gnodev maketx call ... devtest`.
Pros: zero side effects on the user's main keybase. Cons:

- The user-facing acceptance test for this work is *literally*
  "`gnokey maketx call ... dev` works without `-home`". Requiring
  a flag every invocation makes copy-pasted snippets from docs, tests,
  and PRs break unless they carry the `-home` everywhere — which they
  do not today.
- Tools that don't take `-home` (third-party wallets reading
  `~/.gnokey/`, scripts) still wouldn't see the key.
- Hidden separate keybase splits the user's mental model: "did I add
  the key to gnodev's keybase or my own?" — exactly the kind of
  paperwork this change is removing.

We picked "mutate `~/.gnokey/`" because the cost is one extra
local key file (containing a public seed) and the payoff is a
zero-flag workflow.

### 3. Conflict policy: overwrite on name collision

Considered always overwriting any pre-existing `devtest` to enforce a
canonical mapping. Rejected: silently replacing a user's named key, even
one they happened to name `devtest` for unrelated reasons, is worse than
the inconvenience of a warning. Two collision cases are guarded
separately. Same name, different address (an unrelated key named
`devtest`) is left untouched after a warning. Same address, different
name (the deployer seed already imported as, say, `test1`) is also left
untouched: the keybase enforces one name per address and would delete the
existing name if we imported `devtest`, so gnodev detects the address up
front and skips the import entirely.

### 4. Naming: `test1`, `dev`, or `devtest`

The existing in-process constant is `DefaultAccount_Name = "test1"`.
We considered three names for the user-facing keybase entry:

- **`test1`** — matches the internal constant. Reads as "the first of
  N test accounts" (it isn't), and an out-of-the-box keybase entry
  called `test1` looks like leaked test fixture rather than something
  the user is supposed to sign with.
- **`dev`** — short and consistent with the command itself (`gnodev`).
  It is also a plausible name for a real user key, so the
  conflict-detection branch (case 3 above) fires more often than it
  would for a more obscure name.
- **`devtest`** (chosen) — explicit "dev-chain test key", and the name
  the documentation already gives this account: `docs/resources/gnodev.md`
  names it in every sample it prints. Verbose as a CLI argument, and
  unlikely to collide with a key the user made.

The in-process constant is unchanged; only the keybase entry takes
the user-facing name. The address book still resolves any pre-existing
`test1`-keyed entries (from old `gnokey add -recover` habits) by
address.

## Consequences

- **`~/.gnokey/` is mutated only on request**, by `I` or
  `-import-dev-key`. Within an existing home directory, gnodev then
  creates the `data/` subdir with the same permissions `gnokey add`
  would (`os.EnsureDir(..., 0o700)`). This is the first time gnodev
  produces persistent state outside its own data dir, and it never
  happens unasked.
- **Arbitrary `-home <path>` is never silently created.** If the user
  passes a `-home` that does not point at an existing directory,
  gnodev logs a warning and falls back to in-memory tracking instead
  of materializing the path. The default home (`gnoenv.HomeDir()`,
  typically `~/.config/gno/`) is created on demand if missing, so a
  requested import lands on a fresh install.
- Side effects are bounded: at most one new keybase entry, named
  `devtest`, pointing at the well-known public address. Existing entries
  are never overwritten.
- Users who already imported the same seed under another name (commonly
  `test1`) keep that entry. gnodev sees the address is already present
  and skips the import, so no `devtest` entry is added for them and they go
  on signing under their existing name. The keybase enforces one name
  per address, so a single address can never carry both names at once.
- A user who already holds an *unrelated* key called `devtest` sees the
  conflict warning and keeps that entry untouched. They can rename their
  key, or leave the import unasked for.
- A degraded keybase never blocks startup. A missing or unwritable home,
  a locked or corrupt keybase, or a failed import each logs a warning
  and falls back to in-memory tracking, matching the other
  `importDevKey` branches; gnodev still boots.
- The startup banner no longer logs the mnemonic. Tooling that scraped
  it from gnodev output will break; the same constant is available at
  `integration.DefaultAccount_Seed` in the source. `gnokey export devtest`
  produces an armored private key, not the seed phrase.
- Test coverage in `contribs/gnodev/setup_address_book_test.go`
  exercises the keybase states (empty, address-present-under-`devtest`,
  address-present-under-another-name,
  name-`devtest`-with-conflicting-address),
  `home==""`, missing-`home`, unwritable-default-home, and broken or
  unwritable keybase branches of `importDevKey`, the plain boot that writes
  nothing, the `-import-dev-key` boot that puts `devtest` in the address
  book, and the `I` path, which imports and then replaces the placeholder
  name. The fallback log is asserted not to echo the mnemonic.
