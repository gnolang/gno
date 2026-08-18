# `r/sys/users`

The system realm that owns the (name → address, address → user) registry for
gno.land. It is intentionally minimal: it stores `UserData` records, exposes
resolve/update/delete primitives, and gates writes through a controller
whitelist managed by GovDAO (`ProposeNewController` /
`ProposeControllerRemoval` / `ProposeControllerAdditionAndRemoval`).

This realm does **not** define what a "name" is, what registration costs, or
whether names can be transferred. Those policies live in *controller realms*
that the DAO whitelists. See `r/sys/namereg/v1` for one such controller, and
the `examples/gno.land/r/sys/names` realm for the related namespace verifier
that gates package deployment under `gno.land/r/<namespace>/...`.

## Trust boundary at genesis (height 0)

The whitelist check in `RegisterUser` (and the sibling
`AddControllerAtGenesis`) **short-circuits at chain height 0**:

```go
// store.gno
if runtime.ChainHeight() > 0 && !controllers.Has(runtime.PreviousRealm().Address()) {
    return NewErrNotWhitelisted()
}
```

This is **intentional**, not a bug. Genesis is the bootstrap window where:

1. The controller whitelist is empty (it can't be populated until *after* it
   exists).
2. System realms (`r/sys/users/init`, `r/sys/namereg/v1`, etc.) need to
   pre-seed users and add themselves as controllers.
3. Any realm whose `init()` runs at genesis can therefore call `RegisterUser`
   without authorization.

The protection model is **out-of-band trust**: chain operators control which
realms ship in genesis (via the contents of `examples/gno.land/r/...`), and
those realms are vouched for at chain-binary build time. The realm code does
not — and intentionally does not try to — enforce who is "allowed" to
pre-register at height 0.

### Audit reference

This bypass was flagged as audit finding #4 ("Genesis bypass — any caller can
register at height 0"). After review, it is treated as **WON'T FIX, working
as intended**:

- Removing the bypass breaks every legitimate genesis pre-registration use
  case (including this realm's own bootstrap and `r/sys/namereg/v1`'s
  preregister loop of system names).
- A hardcoded genesis-allowlist (a la "only `r/sys/*` realms may bypass")
  shifts the trust to a literal in source — a chain upgrade is required to
  add a new genesis-bootstrap realm. This trades flexibility for the same
  amount of trust.
- Path-prefix gating (e.g. "only `gno.land/r/sys/*`") couples this realm to
  the namespace verifier remaining locked-down, an implicit dependency that
  makes future refactors fragile.

If chain operators want post-deployment auditing of who pre-registered what
at genesis, the `RegisterUserEvent` is emitted on every successful
registration regardless of height, and the source of each registration can be
recovered by walking genesis-block events alongside the `examples/` tree.

### Sibling bypass: `AddControllerAtGenesis`

The same height-0 trust model applies to `AddControllerAtGenesis` in
`admin.gno`:

```go
func AddControllerAtGenesis(_ realm, addr address) {
    height := runtime.ChainHeight()
    if height > 0 {
        panic("AddControllerAtGenesis can only be called at genesis (height 0)")
    }
    if !addr.IsValid() {
        panic(ErrInvalidAddress)
    }
    controllers.Add(addr)
}
```

This was audit finding #7 ("AddControllerAtGenesis has no caller check"). It
is the **same intentional design** as #4 and is likewise treated as **WON'T
FIX**:

- Any realm whose `init()` runs at genesis can whitelist any address as a
  controller, without authorization.
- This is how the registry bootstraps itself: `r/sys/users/init.Bootstrap`
  adds its own package address, and `r/sys/namereg/v1/init.gno` likewise
  auto-whitelists `gno.land/r/sys/namereg/v1`. Removing the bypass would
  break the bootstrap pattern.
- After genesis (height > 0) the function hard-panics, so the privilege
  window is strictly one-time at chain birth.
- The trust model is identical: chain operators vouch for whatever realms
  ship in `examples/` at chain-binary build time.

If you need to add a new controller post-genesis, the supported path is a
GovDAO proposal via `ProposeNewController` — the same channel that rotates
every controller going forward.

### What the audit DID flag that's worth fixing

- #5: `ufmt.Sprint` used instead of `Sprintf` in controller-swap proposal
  description (governance-vote readability).
- #6: Add+Remove proposal silently no-ops if `add` fails on an already-listed
  controller — voted-on swap doesn't actually swap.
- #7: `AddControllerAtGenesis` shares the height-0 bypass; same trust model
  applies, same intentional design.

See `NAMEREG_AUDIT.md` for the full set and `NAMEREG_TODO.md` for tracked
"won't fix / accepted risk" items.
