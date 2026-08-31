# hub

Simplified, read-only view types over `gno.land/p/gnoland/boards` data:
`Board`, `Thread`, `Comment`, `Flag` and `Member`.

Realms use these to expose board contents through a query API without
handing callers access to their own persistent state.

## Invariant

**A safe type is a snapshot, never a live reference.**

Each `NewSafe*` constructor reads the fields it needs off the `boards`
value and copies them. It keeps no pointer to the source, so a value
returned to a caller cannot be used to reach — let alone mutate — the
realm data it was built from. Counts (`ThreadCount`, `FlagCount`, …) are
resolved at construction time and do not track later changes.

Keep it that way when adding fields:

- Copy scalars. Never store the `*boards.Board` / `*boards.Post` the
  constructor was handed, and never expose a `boards.PostStorage`,
  `boards.FlagStorage` or `boards.Permissions` — those are handles onto
  live realm state.
- Deep-copy slices and maps, as `NewSafeBoard` does for `Aliases` and
  `NewSafeMember` does for `Roles`. Copy on the way out too: a getter
  that returns the stored slice lets a caller mutate the snapshot and
  change what the same value reports on the next call, so `Aliases()`
  and `Roles()` each return a fresh slice.
- Convert `boards` types to plain ones where practical, the way `Member`
  flattens `[]boards.Role` to `[]string`.

An earlier version of these types carried a `ref` field plus `Iterate*`
methods that walked realm storage through it. That is the thing this
package exists to not do.

## Usage

```go
import (
    "gno.land/p/gnoland/boards"
    hubexts "gno.land/p/gnoland/boards/exts/hub"
)

func GetBoard(id uint64) (hubexts.Board, bool) {
    b, found := gBoards.Get(boards.ID(id))
    if !found {
        return hubexts.Board{}, false
    }
    return hubexts.NewSafeBoard(b), true
}
```

The constructors panic on a nil reference, and on a post whose kind does
not match (`NewSafeThread` on a comment, or `NewSafeComment` on a
thread). Resolve and check existence before calling them.
