# ADR: migrate boards permissions from commondao to `p/nt/groups/v0`

## Context

`p/gnoland/boards/exts/permissions` used a `*commondao.CommonDAO` purely as
role-bucket storage, via the `commondao/v0/exts/storage` extension — a
message-broker mirror that auto-added every role member to a root member set
so that `Has`/`Size`/`IterateByOffset` covered all users cheaply. This was
the only external consumer of commondao. `p/nt/groups/v0` (PR #6009) now
provides the same primitive directly. This is PR 3 of the GROUPS.md plan,
done before PR 2 (commondao's internal migration): with boards divorced,
PR 2 may freely delete `exts/storage` and the `Member*` types.

## Decision

1. **`Permissions` holds a `*groups.Group`** (unexported) instead of a
   `*commondao.CommonDAO` + storage ext. Both commondao imports are gone;
   the zero-caller `DAO()` accessor is removed.
2. **Base set = every user; role member sets are subsets.** The GROUPS.md
   sketch had base = guests-only with aggregated reads (`HasAny`,
   `TotalSize`, `IterateAll`), but the storage ext's observable behavior is
   a root set containing all users. Mirroring that as the base set keeps
   `HasUser`/`UsersCount`/`IterateUsers` on the base-only methods
   (`Has`/`Size`/`Iterate`): identical globally-address-sorted iteration
   order, O(log N) membership, O(1) count — versus O(N) scan-and-dedup per
   `UsersCount` call under the guests-only mapping. The subset invariant is
   maintained in one place (`SetUserRoles` always `group.Add(user)`;
   `RemoveUser` is `RemoveFromAll`).
3. **Mechanical mappings**: `Members().Grouping().Get/Add/Has` →
   `GetRole/AddRole/HasRole`; `storage.GetMemberGroups` →
   `RolesContaining` (same lexicographic order, same nil-when-none);
   role `SetMeta/GetMeta` → `Role.SetMeta/Meta` with the same
   `boards.PermissionSet` value type (compliant with the groups meta rule:
   `[]uint64`-backed, no mutator-bearing pointers).
4. **Intentional behavior deltas** (everything else is preserved exactly;
   all three are strict improvements, verified old-vs-new by review):
   - `IterateUsers` previously always returned `stopped == false` (bare
     `return` discarding the delegate's result); it now propagates the
     early-stop result. Iteration always halted internally in both — only
     the reported value was wrong. No in-repo caller branches on it.
   - `SetUserRoles` with an invalid role name now panics before mutating
     anything (role names are resolved upfront). The old code cleared the
     user's existing roles — and dropped root-set membership via the
     broker — before hitting the panic, leaving partial state observable
     under `recover()`.
   - Empty and whitespace-only role names are rejected in `AddRole`,
     regardless of super-role configuration. Old commondao
     trimmed-and-rejected whitespace inside `NewMemberGroup`, but
     `AddRole("")` silently no-op'd via the unset-superRole sentinel;
     the guard now lives in `AddRole` itself, above the sentinel.
     `WithSuperRole` rejects such names too — previously it silently set
     a super role with no backing role registration (a broken
     half-state), so that path is a fix, not a restoration. Panic
     messages differ from the old commondao text.

## Alternatives considered

- **Guests-only base per the plan sketch**: rejected — changes
  `IterateUsers` ordering (base first, then per-role order instead of
  globally sorted), makes `UsersCount` O(N) with an allocation, and turns
  `HasUser` into an O(R log N) scan. The plan's own text delegates the
  base-vs-aggregated choice to each call site.
- **Migrating commondao in the same PR**: rejected — that migration flips
  governance semantics (`Has` vs `HasAny` per call site) and breaks the
  `ProposalDefinition.Tally` interface; it deserves its own review (PR 2).

## Consequences

- boards no longer depends on commondao; commondao now has zero external
  consumers, unblocking PR 2's deletion of `member_*.gno` and
  `exts/storage` (~1,285 lines).
- Behavior is preserved exactly (the full 584-line permissions suite passed
  unmodified before the test additions, as do the boards and boards2/v1 +
  hub suites) except the three documented deltas above, each strictly less
  wrong than before.
- The validator trust boundary is now documented on `ValidatorFunc`:
  validators run with full mutation access to the live `Permissions` value
  under the owning realm's authority — owner-controlled functions only.
- One less indirection layer: no message broker, no storage factory — role
  membership writes go straight into the group.
- The cost profile flips from read-optimized to write-optimized. The old
  broker kept a per-user reverse index, so role reads were one index
  lookup; `RolesContaining` now scans all R roles (R = 4 in boards2),
  making `GetUserRoles`/`HasPermission`/`IterateUsers` roughly 2.2–2.4×
  the read gas at a 500-member board. In exchange, membership writes drop
  ~50% gas, per-board storage ~44%, per-membership storage ~61%, and the
  unchanged boards2 + hub suites net −0.6% to −4.9% gas overall (worst
  single filetest +1.06%). Reintroducing a reverse index would give those
  write/storage wins back; not warranted unless a consumer appears with
  many roles or very hot on-chain role reads at large membership.
