# Gno audit loop

This directory contains the private-to-public audit guide loop harness for
sanitized finding families.

The durable repo split is:

- `misc/audit-loop`: executable harness, expected records, and fixtures.
- `docs/resources`: builder-facing guidance promoted from stable findings.
- `examples/gno.land`: compact public examples that demonstrate the safer
  contract pattern.

Run all current slices:

```sh
make run AUDIT_LOOP_FLAGS='-gno-bin /path/to/gno'
```

Development builds of `gno` may need `GNOROOT`:

```sh
GNOROOT=/path/to/gnolang/gno make run AUDIT_LOOP_FLAGS='-gno-bin /path/to/gno'
```

Emit JSON:

```sh
go run ./cmd/auditloop -format json ./expected/*.yaml
```

## Expected records

Current slices:

- `current-guard`: `cur.Previous()` before `cur.IsCurrent()`.
- `render-markdown`: raw `Render(path)` markdown output.
- `payment-user-call`: `OriginSend()` without an `IsUserCall()` guard.
- `origin-caller-auth`: `OriginCaller()` used as authorization identity.
- `callback-param`: caller-supplied callbacks accepted by realm APIs.
- `interface-realm-param`: interfaces that expose `cur realm`.
- `exported-pointer-leak`: exported pointers or pointer getters for mutable
  state.

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

## Adding a slice

1. Add sanitized fixtures under `fixtures/<slice>/`.
2. Add an `expected/<slice>.yaml` record.
3. Teach `internal/auditloop` the new rule.
4. Promote stable, sanitized lessons to `docs/resources` and
   `examples/gno.land` when they are useful for builders.
