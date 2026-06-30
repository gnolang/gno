# ADR: Security Dashboard Realm (`r/sys/security`)

Related issue: https://github.com/gnolang/gno/issues/5084

## Context

Security-critical flags are scattered across multiple production realms
(`r/gnoland/boards2/v1`, `r/sys/validators/v2`, `r/sys/cla`, etc.) and
VM-level parameters (`sysnames_pkgpath`, `syscla_pkgpath`, etc.). There
is no unified view to monitor their state, making it easy to miss a
misconfiguration.

## Decision

Create a read-only realm at `gno.land/r/sys/security` that aggregates
these flags into a single rendered dashboard.

### Design choices

1. **Single unified check registry** — all checks (pre-seeded and
   user-added) live in one `avl.Tree` keyed by structured IDs formatted
   as `realm/path:flag_name` (e.g. `r/gnoland/boards2/v1:realm_locked`).
   The realm portion is extracted from the ID automatically for display.
   Pre-seeded checks are registered at init time with closures that call
   production realm APIs. There is no separate "hardcoded" vs "custom"
   distinction at the data layer.

2. **ID encodes the realm** — the check ID doubles as the realm
   identifier. The format `realm/path:flag_name` lets the dashboard
   derive the realm link from the ID without storing it separately. This
   eliminates the possibility of ID/realm mismatches and reduces the
   parameter count on proposal functions.

3. **Three GovDAO operations: Add, Update, Remove** — a single set of
   proposal functions works on any check in the registry. This means
   GovDAO can update a pre-seeded check if the target realm changes its
   API. If the realm itself changes path (e.g. v1 → v2), the old check
   is removed and a new one is added. Each proposal function takes a
   `description` parameter so the proposer can explain the change to
   voters.

4. **Closures evaluated at render time** — each check stores a
   `func() string` that is called when `Render` is invoked, so the
   dashboard always shows live values.

5. **`matchExpected` with threshold support** — the expected value
   `"> 0"` is handled specially so validator count checks work naturally.
   Plain string equality is used for all other checks.

6. **Bold on mismatch** — when the current value does not match the
   expected value, the current value is rendered in **bold** in the
   table. There is no separate status column; the visual emphasis on the
   value itself is sufficient and keeps the table compact.

7. **"Not Queryable" section (static)** — flags that live at the
   VM/param level cannot be read from Gno code (the `chain/params`
   standard library only exposes Set functions, not Get). The dashboard
   documents them with `gnokey query` instructions. This section is
   static for v0. Tracked flags include:
   - `vm:p:syscla_pkgpath` — CLA enforcement at VM level
   - `bank:p:restricted_denoms` — token transfer restrictions
   - `auth:p:unrestricted_addrs` — transfer lock bypass whitelist
   - `vm:p:default_deposit` — deployment deposit
   - `vm:p:storage_price` — per-byte storage cost
   - `auth:p:fee_collector` — transaction fee recipient
   - `vm:p:storage_fee_collector` — storage fee recipient

8. **Panic safety** — check closures are wrapped in `safeCall` which
   recovers from panics and renders an `ERROR: ...` string instead of
   crashing the entire `Render`.

9. **Help subpage** — `/r/sys/security:help` provides MsgRun templates
   for adding, updating, and removing checks via GovDAO. The main
   dashboard links to it in a footer section.

10. **Executor descriptions** — each `SimpleExecutor` includes a
    human-readable description of the action (e.g. "Add check
    [r/example:flag] to the security dashboard") so it is visible in
    the proposal's executor details.

11. **No `r/gnoland/users/v1` check** — the issue originally listed a
    `paused` flag on users/v1, but that realm was removed in PR #5194
    and the `paused` variable was unexported anyway (no public getter).

### Security model

The proposal functions (`NewAddCheckProposalRequest`, etc.) are
constructors — they return an inert `dao.ProposalRequest` value. The
callback closure is stored in an unexported `executor` field inside the
DAO package, with no public getter. The only code path that invokes the
executor is `r/gov/dao.ExecuteProposal`, which requires a supermajority
(66.66%) YES vote from GovDAO members. A malicious realm cannot extract
or execute the callback.

### Alternatives considered

- **Off-chain monitoring only** — rejected because on-chain visibility
  benefits governance participants who can check the dashboard from any
  Gno client without running custom scripts.

- **Auto-discovery of all realm variables** — not feasible in the GnoVM;
  there is no reflection or introspection mechanism to enumerate another
  realm's state.

- **Separate hardcoded + custom + override trees** — rejected in favour
  of a single registry. Having multiple trees added complexity without
  benefit; a single Add/Update/Remove API is simpler and more flexible.

- **Writable dashboard (direct admin mutations)** — rejected in favour
  of GovDAO proposals for all mutations, since this is a `sys/` realm
  and governance control is appropriate.

- **Separate ID and Realm fields** — an earlier design had a separate
  `Realm` field on the `Check` struct plus a short ID. This was merged
  so the ID encodes the realm (`realm/path:flag_name`), eliminating
  redundancy and reducing the API surface.

- **Governable non-queryable section** — considered adding a second
  `avl.Tree` for non-queryable flags with its own proposal functions.
  Deferred for v0 since these flags rarely change and a static section
  is simpler.

## Consequences

- Adds a cross-realm dependency from `r/sys/security` to `boards2/v1`,
  `validators/v2`, `cla`, and `gov/dao`. If any of these realms change
  their API, GovDAO can update the affected check via
  `NewUpdateCheckProposalRequest` without redeploying.

- Closures stored in realm state that reference removed or upgraded
  realms will panic at render time; the `safeCall` wrapper ensures
  graceful degradation.

- The dashboard is purely informational — it does not enforce any
  security policy. Operators should still set up off-chain alerting for
  critical flags.
