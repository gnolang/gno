# Gno audit pattern harness

This directory contains the private-to-public audit pattern harness
for sanitized finding families.

The durable repo split is:

- `misc/audit-pattern-harness`: executable audit pattern harness, expected records, and
  fixtures.
- `docs/resources`: builder-facing guidance promoted from stable findings.
- `examples/gno.land`: compact public examples that demonstrate the safer
  contract pattern.

Run all current pattern slices:

```sh
make run AUDIT_PATTERN_FLAGS='-gno-bin /path/to/gno'
```

Development builds of `gno` may need `GNOROOT`:

```sh
GNOROOT=/path/to/gnolang/gno make run AUDIT_PATTERN_FLAGS='-gno-bin /path/to/gno'
```

Emit JSON:

```sh
go run ./cmd/auditpattern -format json ./expected/*.yaml
```

Run the deterministic agent contract test:

```sh
go test ./internal/auditpattern -run TestAgentPatternContract
```

This test treats the guide files and expected records as the spec. It verifies
that each pattern family is documented, has vulnerable and fixed fixtures, flags
the vulnerable fixture, and leaves the fixed fixture clean. The compile-check
variant runs automatically when `GNO_BIN` is set or `gno` is available on
`PATH`; otherwise it skips with instructions.

## Expected records

Current pattern slices:

- `current-guard`: `cur.Previous()` before `cur.IsCurrent()`.
- `render-markdown`: raw `Render(path)` markdown output.
- `payment-user-call`: `OriginSend()` without an `IsUserCall()` guard.
- `origin-caller-auth`: `OriginCaller()` used as authorization identity.
- `callback-param`: caller-supplied callbacks accepted by realm APIs.
- `interface-realm-param`: interfaces that expose `cur realm`.
- `exported-pointer-leak`: exported pointers or pointer getters for mutable
  state.
- `render-map-iteration`: public Render output that depends on map iteration
  order.

Each `expected/*.yaml` record describes one finding family and its fixtures:

```yaml
id: current-guard
title: cur.Previous without cur.IsCurrent
rule: current_guard
repair:
  from_fixture: vulnerable
  to_fixture: fixed
  goal: Check cur.IsCurrent before reading cur.Previous for authorization.
  allow_removed_exports: [] # optional, for intentionally removed unsafe APIs
fixtures:
  - name: vulnerable
    path: ../fixtures/current-guard/vulnerable
    want_gno_test: pass
    want_pattern_hits: 1
  - name: fixed
    path: ../fixtures/current-guard/fixed
    want_gno_test: pass
    want_pattern_hits: 0
```

Paths are relative to the YAML file. `want_gno_test` is `pass` or `fail`.
`want_pattern_hits` is the exact count expected from the rule.

The `repair` block is experimental. It describes the intended bad-to-good
fixture pair for an agent or tool to learn from. `TestRepairContracts` verifies
that the source fixture demonstrates the pattern, the target fixture removes it,
the `.gno` files actually changed, and exported top-level function names remain
stable across the repair except for names listed in `allow_removed_exports`.

## Adding a pattern slice

1. Add sanitized fixtures under `fixtures/<slice>/`.
2. Add an `expected/<slice>.yaml` record.
3. Add a `repair` block that points from the vulnerable fixture to the fixed
   fixture and states the intended remediation goal.
4. Teach `internal/auditpattern` the new rule.
5. Promote stable, sanitized lessons to `docs/resources` and
   `examples/gno.land` when they are useful for builders.

## Known limitations

Pattern detection is heuristic — the rules scan text, not an AST. Expect both
false positives and false negatives in real-world code.

### current_guard

Detects `.Previous()` before `.IsCurrent()` only within the **same function**.
If the `IsCurrent()` check lives in a helper function called from the same
function that calls `.Previous()`, the detector will not flag it. Check helper
call chains manually when auditing cross-realm code.

### payment_user_call

Flags any `OriginSend()` call not preceded by `.IsUserCall()` in the same
function — this catches both the no-guard case **and** the wrong-guard case
(`IsUser()` instead of `IsUserCall()`). The rule does not distinguish between
the two; both appear as pattern hits.

### render_markdown_escape

Flags `return` statements inside `Render` that contain the `path` variable
without the word `escape` on the same line. Patterns using intermediate
variable names (e.g. `sanitized := md.EscapeText(path); return sanitized`)
are not flagged even when unsafe.

### Spec corpus test

`TestAgentPatternContract` verifies that each pattern family's required terms
appear somewhere in the spec files. It confirms that the docs discuss the
topic; it does not verify that the security advice in those docs is correct.
Manual review is still required before promoting new guidance.
