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

Read the default markdown report as follows:

- `status: PASS` means every fixture matched its expected compile/test result
  and pattern-hit count.
- `gno test: want pass|fail, ok true|false` compares the fixture's own Gno
  tests with the expected outcome.
- `pattern hits: got N, want N` compares the rule's text-scan matches with the
  expected finding count.
- listed `file:line` hits are the lines an auditor or agent should inspect
  before deciding whether the finding is real or a heuristic false positive.

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

## Adding a pattern slice

1. Add sanitized fixtures under `fixtures/<slice>/`.
2. Add an `expected/<slice>.yaml` record.
3. Teach `internal/auditpattern` the new rule.
4. Promote stable, sanitized lessons to `docs/resources` and
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

### origin_caller_auth

Flags `OriginCaller()` only when it appears in a direct equality or inequality
check. Benign logging reads are ignored, but an authorization check split across
multiple lines or hidden behind a helper can be missed.

### exported_pointer_leak

Flags exported package-level pointer variables and exported pointer-returning
functions. Constructors shaped like `NewX() *X { return &X{} }` are ignored as
fresh allocations; manually inspect constructors that may return aliases to
shared package state.

### render_map_iteration

Flags `for ... range <m>` inside `Render` where `<m>` is a package-level `map`
variable, matched at a word boundary so a map `scores` does not flag an
unrelated `range scoresList`. A map ranged behind a local alias, or built
inside `Render`, is not detected.

### Line reporting

Sources are gofmt-normalized before matching so irregular spacing cannot defeat
the rules, but every hit's `file:line` and text are mapped back to the original
on-disk source, so they stay accurate even on input that was not gofmt-clean.

### Spec corpus test

`TestAgentPatternContract` verifies that each pattern family's required terms
appear somewhere in the spec files. It confirms that the docs discuss the
topic; it does not verify that the security advice in those docs is correct.
Manual review is still required before promoting new guidance.
