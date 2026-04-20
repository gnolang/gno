# PRxxxx: Interactive `gno init` with Template Scaffolding

## Context

`gno mod init` only creates a bare `gnomod.toml` file, and is buried as a subcommand
of `mod`. New developers must then manually create source and test files, look up
realm conventions (e.g. `Render`), and figure out the correct package declaration.
There is also no scaffolding for `gnokey maketx run` scripts. This friction slows
onboarding compared to tools like `cargo init` or `npm init` that scaffold working
starter code.

## Decision

### Make `gno init` a top-level command, interactive when run in a terminal

`init` is promoted from `gno mod init` to `gno init` for ergonomics — it's the
first command new users reach for and shouldn't be buried under `mod`.

### Three module kinds: realm, package, run script

When stdin is a TTY and `--bare` is not set, the user is prompted for the module
kind if it can't be auto-detected from the path:

```
Module kind: [r]ealm, [p]ackage, or [m]ain (run script)? [r/p/m] (default: p):
```

- **Realm** (`/r/`): creates `gnomod.toml`, `<pkg>.gno` with `Render`, `<pkg>_test.gno`.
- **Package** (`/p/`): creates `gnomod.toml`, `<pkg>.gno`, `<pkg>_test.gno`.
- **Run script** (main): creates only `main.gno` — no `gnomod.toml` needed since
  the keeper auto-generates one for `gnokey maketx run`.

If a module path is provided with `/r/` or `/p/`, the kind is auto-detected.
If the path is missing the letter segment (e.g. `gno.land/myname/foo`), the user
is prompted and the letter is inserted automatically.

### Template selection menu

Each module kind has a list of templates (currently one "basic" template each).
When only one template is available, it's auto-selected. When multiple exist,
a numbered menu is shown:

```
Select a template:
  [1] basic - minimal realm with a Render function
  [2] dao - DAO with proposals and voting
Choice [1-2] (default: 1):
```

Adding a new template requires only: (1) add `.tmpl` files under
`templates/<kind>/<name>/`, (2) add one entry to the registry in
`mod_init_templates.go`. No other Go code changes needed.

### Non-interactive fallback

When stdin is not a TTY (CI, piped input) or `--bare` is passed, behavior is
identical to the original `gno mod init`: only `gnomod.toml` is created, module
path must be provided as argument. This preserves backward compatibility.

### TTY detection

Uses `golang.org/x/term.IsTerminal()` on `os.Stdin.Fd()`. Already a dependency
via `tm2/pkg/commands/utils.go`.

### Template files are embedded, not hardcoded

Templates live as `.tmpl` files under `gnovm/cmd/gno/templates/` and are compiled
into the binary via `go:embed`. `text/template` renders them with `{{.PkgName}}`.

```
templates/
  realm/
    basic/
      source.gno.tmpl      # Render function
      test.gno.tmpl         # TestRender
  package/
    basic/
      source.gno.tmpl      # bare package declaration
      test.gno.tmpl         # placeholder test
  run/
    basic/
      source.gno.tmpl      # package main + func main()
```

### No README generation

Considered and rejected. Gno realms have `Render()` for user-facing description,
existing examples don't include READMEs, and a placeholder README adds noise.

## Alternatives Considered

1. **Always generate templates (no flag)** — Rejected because it would break
   non-interactive callers and existing test expectations.
2. **`--template` flag to opt-in** — Rejected because the common case (human in
   a terminal) should scaffold by default. Requiring a flag hurts discoverability.
3. **Separate `gno new` command** — Rejected. `gno init` in the current directory
   is sufficient and matches Go conventions. No need for directory creation.
4. **`--main` flag instead of kind prompt** — Rejected in favor of a unified
   interactive prompt for all three kinds, which is more consistent and extensible.
5. **Prompt for "include tests?"** — Rejected for v1. Tests should always be
   encouraged. Can be added later if users request it.

## Key Files

| File | Role |
|------|------|
| `gnovm/cmd/gno/main.go` | `gno init` registered as top-level command |
| `gnovm/cmd/gno/mod.go` | `newInitCmd`, `execModInit`, `execInitRun`, `promptModuleKind`, `selectTemplate`, `insertPathLetter` |
| `gnovm/cmd/gno/mod_init_templates.go` | `go:embed` declarations, `initTemplate` registry, `renderTemplate` |
| `gnovm/cmd/gno/templates/{realm,package,run}/basic/*.tmpl` | Template files |
| `gnovm/cmd/gno/mod_test.go` | Tests for helpers and init flows |

## Consequences

- New users get a working starter project with `gno init gno.land/r/myname/myrealm`.
- Run scripts can be scaffolded with `gno init` → select "main" → creates `main.gno`.
- Existing non-interactive usage is unchanged (all existing tests pass unmodified).
- The `--bare` flag provides an explicit escape hatch.
- Template content is minimal and opinionated — may need iteration based on feedback.
- Adding new templates (e.g. dao, grc20) only requires `.tmpl` files + one registry entry.
