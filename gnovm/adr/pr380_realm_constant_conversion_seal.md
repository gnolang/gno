# Enforce interface seals on constant conversions

## Status

Implemented in PR #380. Hard-fork-class because package acceptance changes.

## Context

The uverse `realm` interface is intentionally sealed. Its VM definition has an
undeclarable `.seal` method, so only the native `.grealm` runtime type can
satisfy it. Code that authorizes an operation from a `realm` value relies on
this invariant: its liveness (`IsCurrent`) and identity (`Address`, `PkgPath`)
must come from a genuine runtime realm rather than attacker-controlled virtual
methods.

The constant-expression branch of preprocessing returned before the normal
interface-conversion check. A named primitive type could therefore implement
all public `realm` methods and be converted with an expression such as
`realm(fake(1))`. Preprocessing relabelled the expression with the static
`realm` type without checking the hidden seal, while the stored value retained
the attacker-defined concrete type. Calls through the interface then
dispatched to attacker-controlled methods.

This is exploitable through `chain/banker`: a forged value can report both an
arbitrary address and `IsCurrent() == true`, obtain a `RealmSend` banker, and
move coins from that address. The Go type-checking shim also described `realm`
using only its public methods, so it did not reject the conversion earlier.

## Decision

Before evaluating and relabelling a constant conversion whose target is an
interface, preprocessing now calls the same `mustAssignableTo` check used by
non-constant interface conversions. This is the consensus-enforced fix and
applies to every interface, not only `realm`.

The Go type-checking shim also gives `realm` an unexported `seal` method. User
packages cannot directly implement that package-scoped method, so the
preliminary type check rejects structural forgeries that reproduce only the
public method set, including this exploit. A type embedding `realm` can still
inherit the shim method through promotion; the VM's undeclarable `.seal`
remains authoritative and rejects promoted seals.

## Alternatives considered

- Add a concrete-realm check only to `chain/banker`. This would protect the
  known sink but leave the broken sealed-interface invariant available to
  other and future authorizers.
- Change only the Go type-checking shim. That improves diagnostics but does not
  enforce the rule in the VM's own preprocessing layer or in callers that do
  not run the Go-compatible check first.
- Change only VM preprocessing. This closes the exploit, but leaves the two
  type systems inconsistent and delays rejection until the second check.
- Add the seal to the transpiler's separate builtin shim in the same change.
  That path is not used for consensus execution, so it is left for a scoped
  compatibility follow-up rather than expanding this urgent patch.

## Consequences

- Forged `realm` constant conversions are rejected before execution, so they
  cannot construct an authorized banker.
- Constant-to-interface conversions now perform the interface-satisfaction
  walk they previously skipped. Valid conversions to empty or implemented
  interfaces continue to pass, including `nil` conversions.
- Packages that depended on an invalid constant-to-interface conversion will
  stop being accepted. This is a consensus behavior change and requires a
  coordinated network upgrade.
- Preprocessing charges the existing interface-satisfaction work on this path;
  no new gas schedule or native operation is introduced.

## Testing

- `gnovm/tests/files/zrealm_seal_constconv.gno` records both the preliminary
  type-check error and the VM seal error for the vulnerable expression, after
  confirming a valid conversion to a normal non-empty interface still passes.
- `gnovm/tests/files/zrealm_seal_constconv_iface.gno` converts a constant to a
  named interface embedding `realm`, so a check narrowed to `realm` itself
  would let that conversion forge a realm.
- `gno.land/pkg/integration/testdata/banker_realm_constconv.txtar` verifies that
  a package containing the banker exploit cannot be deployed and that the
  targeted balance remains unchanged.
- Existing VM filetests cover successful constant conversions to `any`,
  including `any(nil)`, guarding the valid empty-interface cases.
