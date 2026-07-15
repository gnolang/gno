# ADR: Separate "Open" and "Browse" in the gnoweb explorer listing

## Context

When visiting a user/namespace package listing in gnoweb (e.g.
`/r/g1.../`), the handler falls through to the explorer paths-list view
(`GetPathsListView` → `DirectoryView` with `ViewModeExplorer`, rendered by
`components/views/directory.html`).

In that view each entry had two navigation targets:

- The main entry link (the package name) pointed at `{{ .Link }}/` — the
  **directory listing** of the package (trailing slash → `GetDirectoryView`
  → file listing).
- A right-side inline button group offered `Open` (`{{ .Link }}`, the realm
  render), `Source` (`{{ .Link }}$source`), and `Action` (`{{ .Link }}$help`).

This is surprising: the most prominent target (clicking the package name)
led to the raw file/directory listing rather than the realm's rendered
output, which is what most users want first. The render was demoted to a
small secondary "Open" button.

## Decision

Reorder the two navigation intents in the explorer listing:

1. The main entry link now points at `{{ .Link }}` — the realm **render**
   (same target as the "Open" button), so clicking a package name opens
   what it renders.
2. A new **Browse** inline button (`{{ .Link }}/`) is added to the
   right-side group, immediately after "Open", so the directory listing is
   still one click away.

The right-side group in explorer mode is now: `Open`, `Browse`, `Source`,
`Action` (with `Open`/`Action` still gated out for pure `/p/` packages, as
before). `Browse` is always shown because browsing the directory is
meaningful for both realms and pure packages.

Only the explorer (paths-list) branch changes. The directory-mode reuse of
`renderDir` (source-file listing, `DirLinkTypeSource`) is unaffected: its
inline button group is gated behind `{{ if $.Mode.IsExplorer }}`, and its
main link never appended the trailing slash.

## Alternatives considered

- **Keep the main link as directory listing, only add "Render" button.**
  Rejected: the primary click should lead to the render, which matches user
  expectation and the behavior when opening a realm directly.
- **Add "Browse" only for realms (inside the `/p/` guard).** Rejected:
  browsing the file tree is equally useful for pure packages; there is no
  reason to hide it.

## Consequences

- Clicking a package name in a namespace/user listing now opens the realm
  render instead of the file listing — a behavior change for existing users,
  but the directory listing remains reachable via the new "Browse" button.
- No Go API or data-structure change; the edit is confined to the template
  plus a handler test asserting the new link targets
  (`TestHTTPHandler_ExplorerPathsListBrowse`).
