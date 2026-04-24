# PR5557: Interactive `gno init` with Template Scaffolding

## Context

`gno mod init` only creates a bare `gnomod.toml` file, and is buried as a subcommand
of `mod`. New developers must then manually create source and test files, look up
realm conventions (e.g. `Render`), and figure out the correct package declaration.
There is also no scaffolding for `gnokey maketx run` scripts. This friction slows
onboarding compared to tools like `cargo init` or `npm init` that scaffold working
starter code.

## Decision

### Scaffolding in the current directory

`gno init gno.land/p/demo/foo` creates `gnomod.toml` and template files
directly in CWD — no subdirectory is created. This is the simplest,
least-surprising behavior: the user picks the directory; `gno init` fills
it. Users who want a new directory run `mkdir foo && cd foo && gno init …`
themselves.

All non-run branches (`--bare`, non-interactive with arg, interactive with
arg, full wizard) share the same sequence:

1. Bail out early if `gnomod.toml` already exists in CWD.
2. `validateModulePath(modPath)` — reject invalid paths (e.g. uppercase)
   before any filesystem side effects.
3. Pre-render templates and check for file conflicts (e.g. a pre-existing
   `foo.gno`); bail out if any file would be overwritten.
4. Write `gnomod.toml`, then template files.

Since we only ever write individual files (never directories), a failed
init never leaves orphan directories on disk. No rollback machinery is
needed.

The `.gno` run-script branch is the only case where `gno init` creates
directories — and only for the parent path of the given script (e.g.
`gno init run/hello.gno` creates `./run/` if missing). This matches the
user's explicit request.

**Stderr output** for the success path ends with a next-steps hint:

```
Initialized realm gno.land/r/demo/foo (basic template)
  gnomod.toml
  foo.gno
  foo_test.gno
Next: gno test .
```

For `--bare` (no tests), the hint is a generic "add your code and run
`gno test .`". `gno test .` is used rather than `gno test ./...` because
the latter requires a `gnowork.toml`, which a freshly-scaffolded single
module does not have.

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
- **Run script** (main): creates `run/<name>.gno` — the user chooses the script name.
  No `gnomod.toml` is needed since the keeper auto-generates one for `gnokey maketx run`.

If a module path is provided with `/r/` or `/p/`, the kind is auto-detected.
If the path is missing the letter segment (e.g. `gno.land/myname/foo`), the user
is prompted and the letter is inserted automatically.

### `.gno` argument shorthand for run scripts

If the argument ends in `.gno`, a run script is created at that path without
`gnomod.toml`. For example, `gno init run/create_proposal.gno` creates
`run/create_proposal.gno`. The `--template` flag is respected; `--bare` is rejected
as mutually exclusive.

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

### Directory-based templates

Templates are defined by directory, not by listing individual files. `renderTemplateDir`
walks the template directory, processes every `.tmpl` file, and produces a map of
output filename → rendered content. Filenames are also Go templates: `{{.PkgName}}.gno.tmpl`
produces `<pkgName>.gno`. This supports complex multi-file templates (e.g. a DAO
template with state, helpers, and tests) without special-casing source/test files.

```
templates/
  realm/
    basic/
      {{.PkgName}}.gno.tmpl       → <pkg>.gno (Render function)
      {{.PkgName}}_test.gno.tmpl  → <pkg>_test.gno (TestRender)
  package/
    basic/
      {{.PkgName}}.gno.tmpl       → <pkg>.gno (bare package declaration)
      {{.PkgName}}_test.gno.tmpl  → <pkg>_test.gno (placeholder test)
  run/
    basic/
      {{.ScriptName}}.gno.tmpl    → <name>.gno (package main + gnokey comment)
```

`templateData` provides `{{.PkgName}}` (from module path) and `{{.ScriptName}}`
(for run script filenames). The run template includes a block comment showing
both `gno run` and `gnokey maketx run` commands with the script path.

### Non-interactive fallback

When stdin is not a TTY (CI, piped input), `gno init <path>` scaffolds template
files (kind auto-detected from the path, first available template used) in
addition to `gnomod.toml`. The `--bare` flag is the explicit escape hatch for
callers that want `gnomod.toml` only — preserving the legacy `gno mod init`
behavior under a single well-defined flag.

### Backward compatibility: `gno mod init`

`gno mod init` is kept as a `gno mod` subcommand that preserves its original
bare behavior: it always writes a minimal `gnomod.toml` in the current
directory (never triggers the interactive wizard and never scaffolds template
files). It prints a one-line hint on stderr pointing at `gno init` for users
who want the richer flow. Existing scripts, Makefiles, and CI pipelines keep
working unchanged.

### Shared prompt primitives in `tm2/pkg/commands/`

Interactive prompts (string input, single-key choice, numbered select)
are extracted into `tm2/pkg/commands/prompt.go` as reusable primitives. This
allows other CLI tools (e.g. `gnokey maketx`) to build their own interactive
wizards without duplicating prompt logic or adding external dependencies.

Key functions: `IsInteractive()`, `PromptString()`, `PromptChoice()`,
`PromptSelect()`.

The wizard is a linear flow — no back-navigation. Backspace cannot be detected
in line-buffered terminal mode, and alternative go-back keys (like `<` or `b`)
conflict with valid user input in `PromptString`. The user can Ctrl+C and
restart the wizard instead.

### File conflict detection

Before writing any template file, `writeModule` and `execInitRun(Script)` check
whether each output file already exists. If a conflict is found, the init
aborts with an error like `file already exists: myrealm.gno`. This prevents
silent overwrites on accidental re-init.

### TTY detection

Uses `commands.IsInteractive()`, which wraps `golang.org/x/term.IsTerminal()`
on `os.Stdin.Fd()`. Already a dependency via `tm2/pkg/commands/utils.go`.

### Template files are embedded, not hardcoded

Templates live as `.tmpl` files under `gnovm/cmd/gno/templates/` and are compiled
into the binary via `go:embed`. `text/template` renders them with `{{.PkgName}}`
and `{{.ScriptName}}`.

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
6. **External TUI library (Bubble Tea, huh, survey)** — Rejected. Every mature
   Go prompt library pulls in the Charm stack (~15 transitive deps). The
   `commands` package already has `GetString`/`GetPassword`/`GetConfirmation`
   and `golang.org/x/term`; building on those keeps the dependency footprint
   at zero new imports. The shared primitives in `prompt.go` are ~190 lines
   of code, fully testable via `commands.IO`, and sufficient for sequential
   wizard flows.
7. **Go-back navigation (`ErrGoBack`)** — Rejected. Backspace/delete key cannot
   be detected in line-buffered terminal mode (the terminal handles it before
   the application sees input). Alternative go-back keys (`<`, `b`) conflict
   with valid user input in `PromptString` (e.g. namespace "b"). A linear
   wizard without back-navigation is simpler; the user can Ctrl+C and restart.

## Key Files

| File | Role |
|------|------|
| `tm2/pkg/commands/prompt.go` | Shared prompt primitives (`PromptString`, `PromptChoice`, `PromptSelect`, `IsInteractive`) |
| `tm2/pkg/commands/prompt_test.go` | Tests for prompt primitives |
| `gnovm/cmd/gno/main.go` | `gno init` registered as top-level command |
| `gnovm/cmd/gno/mod.go` | `gno mod` subcommands: `newModCmd`, `newModInitCmd` (bare gnomod.toml), `newModDownloadCmd`, `newModGraphCmd`, `newModTidy`, `newModWhy` |
| `gnovm/cmd/gno/init.go` | `newInitCmd`, `execModInit`, `validateModulePath`, `scaffoldModule`, `renderModuleFiles`, `writeFiles`, `execInitRun`, `writeRunScript`, `promptModuleKind`/`promptModulePath`, `selectTemplate`, `insertPathLetter`, `validateGnoPath`, `validateNamespace` (accepts a name or a bech32 address) |
| `gnovm/cmd/gno/mod_init_templates.go` | `go:embed` declarations, `initTemplate` registry, `renderTemplateDir` |
| `gnovm/cmd/gno/templates/{realm,package,run}/basic/*.tmpl` | Template files with `{{.PkgName}}` and `{{.ScriptName}}` in filenames |
| `gnovm/cmd/gno/mod_test.go` | Tests for `gno mod` subcommands and `gno init` integration cases (via `testMainCaseRun`) |
| `gnovm/cmd/gno/init_test.go` | Unit tests for init helpers, wizard prompts, and scaffolding flows |
| `gnovm/cmd/gno/testdata/init/*.txtar` | End-to-end testscript scenarios covering CWD scaffolding, `--bare`, conflicts, invalid paths, run-script shorthand, legacy `gno mod init` alias, and flag validation |

## Consequences

- New users get a working starter project with `gno init gno.land/r/myname/myrealm`.
- Run scripts can be scaffolded with `gno init run/create_proposal.gno` (shorthand)
  or `gno init` → select "main" → choose script name.
- Existing non-interactive usage is unchanged (all existing tests pass unmodified).
- The `--bare` flag provides an explicit escape hatch.
- Template content is minimal and opinionated — may need iteration based on feedback.
- Adding new templates (e.g. dao, grc20) only requires `.tmpl` files + one registry entry.
- Complex multi-file templates are supported — just add more `.tmpl` files in the
  template directory.
- The shared prompt primitives in `tm2/pkg/commands/prompt.go` can be reused by
  other CLI wizards (e.g. an interactive `gnokey maketx` flow) without code
  duplication or new dependencies.
