# ADR: gnodev auto-imports the `dev` account at startup

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
user has to leave the gnodev terminal, run `gnokey add -recover dev`
(or whichever name they pick), paste the mnemonic at the prompt, and
come back. The mnemonic is public, the workflow is pure paperwork, and
new users routinely get stuck on it during onboarding.

## Decision

When gnodev starts, before importing the local keybase into its
in-memory `address.Book`, it ensures an entry named `dev` exists in
the user's gnokey keybase. Concretely, `ensureDevKey` in
`setup_address_book.go` does:

1. If `-no-dev-key` was passed, log `dev key skipped (-no-dev-key)`
   and return.
2. If `cfg.home == ""`, log a warning and return (no keybase to write
   to; this only happens when `-home ""` is set explicitly).
3. If `cfg.home` is set but the directory does not exist:
   - If it is the default home (compared with `filepath.Clean` on both
     sides so a path-equivalent form such as a trailing slash still
     counts), create it with mode `0o700` so the auto-import fires on a
     fresh install. This matches `gnokey add`'s behavior, which silently
     creates `~/.config/gno/` on first use.
   - Otherwise (an explicit `-home <path>` that doesn't exist, likely a
     typo), log a warning and return without writing. We refuse to
     silently materialize an arbitrary path on disk.
4. Open the keybase at `cfg.home` via `keys.NewKeyBaseFromDir`, which
   creates `cfg.home/data/` on disk if it does not exist (mode 0o700).
   That call panics rather than returning an error when it cannot create
   the dir (e.g. an unwritable home), so the open is wrapped to recover
   the panic and treat it as a normal failure.
5. If the deployer address is already in the keybase under any name, it
   is already signable: log `dev key already present in keybase,
   skipping` and stop. This is the key guard. The keybase enforces one
   name per address, so calling `CreateAccount("dev", ...)` for an
   address already stored under another name (commonly `test1`) would
   silently delete that other name. Skipping preserves the user's
   existing entry.
6. Otherwise, if the name `dev` belongs to a different address (the user
   has an unrelated key they named `dev`), log a one-line warning and
   leave it untouched.
7. Otherwise import via
   `kb.CreateAccount("dev", DefaultDeployerSeed, "", "", 0, 0)` and log
   `dev key imported`.

Every failure along the way (missing or unwritable home, locked or
corrupt keybase, failed import) degrades to a logged warning, never an
error. The import is a convenience, so a degraded keybase must never
stop gnodev from starting; the deployer address is still tracked
in-memory by `setupAddressBook`'s fallback when the import is skipped.

The mnemonic is read from the existing `DefaultDeployerSeed` constant;
no second copy is introduced. The startup no longer logs the mnemonic
at all: the previous banner is replaced by either `dev key imported`
(happy path) or `default address tracked in-memory only; gnokey cannot
sign with it` (opt-out / no keybase). Users who need the mnemonic can
read `integration.DefaultAccount_Seed` in the source; `gnokey export
dev` produces an armored, password-encrypted private key, not the seed
phrase.

A new boolean flag `-no-dev-key` (matching gnodev's `no-web`,
`no-watch`, `no-replay` naming convention) opts out of the import.

## Alternatives Considered

### 1. Default-off opt-in (`--dev-key`)

We chose default-on. Rationale:

- The mnemonic is already public, identical on every machine, and
  already documented in `gnodev`'s output. Importing it adds no secret
  to the user's machine that wasn't trivially derivable from
  `git grep DefaultAccount_Seed`.
- The whole point of gnodev is "smallest possible loop from `make
  install` to signed transaction." A flag the user must remember
  defeats that.
- Users who do not want gnodev mutating `~/.gnokey/` (CI runners,
  shared dev boxes, security-conscious setups) can pass `-no-dev-key`.

### 2. Don't touch `~/.gnokey/`; let gnodev run its own keybase

gnodev would create `$XDG_STATE_HOME/gnodev/keys/` (or similar), import
`dev` there, and tell users to run
`gnokey -home $XDG_STATE_HOME/gnodev maketx call ... dev`.
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

Considered always overwriting any pre-existing `dev` to enforce a
canonical mapping. Rejected: silently replacing a user's named key,
even one they happened to name `dev` for unrelated reasons, is worse
than the inconvenience of a warning. Two collision cases are guarded
separately. Same name, different address (an unrelated key named `dev`)
is left untouched after a warning. Same address, different name (the
deployer seed already imported as, say, `test1`) is also left untouched:
the keybase enforces one name per address and would delete the existing
name if we imported `dev`, so gnodev detects the address up front and
skips the import entirely. Since `dev` is a more plausible name for a
real user key than something like `devtest`, the same-name guard matters
in practice.

### 4. Naming: `test1`, `devtest`, or `dev`

The existing in-process constant is `DefaultAccount_Name = "test1"`.
We considered three names for the user-facing keybase entry:

- **`test1`** — matches the internal constant. Reads as "the first of
  N test accounts" (it isn't), and an out-of-the-box keybase entry
  called `test1` looks like leaked test fixture rather than something
  the user is supposed to sign with.
- **`devtest`** — explicit "dev-chain test key". Self-documenting but
  verbose; reads slightly awkward as a CLI argument
  (`gnokey ... devtest`).
- **`dev`** (chosen) — short, idiomatic, and consistent with the
  command itself (`gnodev`). Trade-off: `dev` is a plausible name for
  a real user key, so the conflict-detection branch (case 3 above)
  matters more here than it would for a more obscure name.

The in-process constant is unchanged; only the keybase entry takes
the user-facing name. The address book still resolves any pre-existing
`test1`-keyed entries (from old `gnokey add -recover` habits) by
address.

## Consequences

- **`~/.gnokey/` is now mutated by gnodev** on first run, unless
  `-no-dev-key` is set. Within an existing home directory, gnodev
  creates the `data/` subdir with the same permissions `gnokey add`
  would (`os.EnsureDir(..., 0o700)`). This is the first time gnodev
  produces persistent state outside its own data dir.
- **Arbitrary `-home <path>` is never silently created.** If the user
  passes a `-home` that does not point at an existing directory,
  gnodev logs a warning and falls back to in-memory tracking instead
  of materializing the path. The default home (`gnoenv.HomeDir()`,
  typically `~/.config/gno/`) is created on demand if missing, so the
  auto-import flow works out of the box on a fresh install.
- Side effects are bounded: at most one new keybase entry, named
  `dev`, pointing at the well-known public address. Existing entries
  are never overwritten.
- Users who already imported the same seed under another name (commonly
  `test1`) keep that entry. gnodev sees the address is already present
  and skips the import, so no `dev` entry is added for them and they go
  on signing under their existing name. The keybase enforces one name
  per address, so a single address can never carry both names at once.
- Because `dev` is a plausible user-chosen key name, users who already
  have an *unrelated* key called `dev` will see the conflict warning
  and keep their existing entry untouched. They can either rename
  their key or run gnodev with `-no-dev-key`.
- A degraded keybase never blocks startup. A missing or unwritable home,
  a locked or corrupt keybase, or a failed import each logs a warning
  and falls back to in-memory tracking, matching the other
  `ensureDevKey` branches; gnodev still boots.
- The startup banner no longer logs the mnemonic. Tooling that scraped
  it from gnodev output will break; the same constant is available at
  `integration.DefaultAccount_Seed` in the source. `gnokey export dev`
  produces an armored private key, not the seed phrase.
- Test coverage in `contribs/gnodev/setup_address_book_test.go`
  exercises the keybase states (empty, address-present-under-`dev`,
  address-present-under-another-name, name-`dev`-with-conflicting-address),
  the opt-out, `home==""`, missing-`home`, unwritable-default-home, and
  broken or unwritable keybase branches of `ensureDevKey`, and two
  end-to-end `setupAddressBook` paths asserting the deployer address ends
  up in the address book under name `dev` (auto-import) or under the
  in-memory `_default#…` fallback (opt-out), with the fallback log not
  echoing the mnemonic.
