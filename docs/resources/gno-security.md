# Gno Security: Threat-Class Taxonomy

This document defines the numbered threat classes referenced throughout
the codebase (e.g. `// SECURITY (Class-4 captured callback)`) and the
companion `SECURITY_GUIDE.md`. It assumes you have read
[`gno-interrealm.md`](./gno-interrealm.md) — the language here uses that
document's vocabulary (realm-context, crossing function, captured realm
value, `IsCurrent()`).

A `cur realm` value is a **language-enforced capability token**: the
runtime mints one per crossing frame, refuses to persist it, and
`rlm.IsCurrent()` returns true only for the topmost live crossing
frame's cur (HIV pointer identity). All five threat classes below are
ways an interface or API design lets agency leak past those built-in
protections.

## The five classes

| # | Name | Mechanism |
|---|---|---|
| **1a** | cur-disclosure / impersonate-self | Hostile interface implementation captures `rlm.Address()` / `rlm.PkgPath()` from a `cur realm` parameter and later acts AS the realm that handed it the cur. |
| **1b** | cur-disclosure / impersonate-caller | Hostile implementation captures `rlm.Previous()` and acts AS that realm's caller. |
| **2** | designation-forgery | Public method takes `(caller address, ...)` or `(pkgPath string, ...)`; any attacker can call it with the victim's identity. The same shape applies to APIs that accept a `realm` value but skip `IsCurrent()` — a stored stale realm value's `.Address()` still resolves but no longer refers to the live caller. |
| **3** | impl-substitution | Public function accepts an open interface; attacker supplies an implementation that lies on read or always-denies (DoS) or always-approves (silent escalation). **Fires even when the interface has no realm-typed methods** — it is a read/behavior-integrity class, not a cur-leak class. |
| **4** | closed-over-authority | A canonical-typed value's constructor (or post-construction setter) takes attacker-controllable callback/data; the value passes an `IsCanonicalX` type check but carries hostile state. When Class 3 and Class 4 both apply, file as Class 4 — the allowlist passed; residual harm is captured-state. |

## Defenses, in one line each

- **Classes 1a/1b**: never declare an interface method that takes
  `cur realm`. Take `caller address` instead, and let the calling code
  derive the address from `cur.Previous().Address()` under an
  `IsCurrent()` guard at the call site.
- **Class 2**: never trust an `address` or `pkgPath` parameter as
  caller-identity; derive it inside the function from
  `rlm.Previous().Address()` under `rlm.IsCurrent()`. Never trust
  `rlm.Address()` without `IsCurrent()` either.
- **Class 3**: pass canonical implementations from the owning package
  (e.g. `NewMemberAuthority`, `NewContractAuthority`). For interfaces
  whose contract requires it, expose an `IsCanonicalX` predicate and
  reject foreign impls at boundary functions.
- **Class 4**: the constructor itself is the trust boundary. Document
  loudly that callback/data arguments under caller control mean the
  caller IS the authority for the lifetime of the constructed value.

See `SECURITY_GUIDE.md` for the long-form patterns and pitfalls,
including the three-rules summary and per-pattern vetting checklist.

## Sealing is not a security boundary

Unexported marker methods on an interface (`isCanonical()` etc.) are
bypassable via embedding in Gno; see
`examples/gno.land/p/test/seal/filetests/z_seal_*_filetest.gno` for the
four bypass tests. Sealing remains useful only as a documentation hint.
For real allowlists, use a concrete-type switch
(`switch v.(type) { case *MyImpl: ... }`) at the boundary function.
