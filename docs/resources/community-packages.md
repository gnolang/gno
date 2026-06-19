# Community Packages

Gno packages under `examples/gno.land/p/...` are useful references and may be
deployed on public networks, but many are community-maintained rather than
official APIs. Treat them like dependencies:

- read the package docs and tests;
- check whether the package is maintained and used by other realms;
- verify the exact escaping, ordering, storage, and error semantics you rely on;
- prefer small, focused packages that are easy to replace;
- write and publish better alternatives when the current package does not fit.

This page is not an endorsement list. It is a starting point for discovering
patterns that may be useful when designing your own realm or package.

## Markdown Helpers

- [`gno.land/p/moul/md`](../../examples/gno.land/p/moul/md/md.gno): helpers for
  building markdown links, headings, lists, images, code blocks, and text
  escaping.
- [`gno.land/p/moul/mdtable`](../../examples/gno.land/p/moul/mdtable/mdtable.gno):
  helpers for markdown tables.
- [`gno.land/p/nt/mdalert/v0`](../../examples/gno.land/p/nt/mdalert/v0/README.md):
  helpers for Gno-Flavored Markdown alert blocks.

Prefer the official Gno markdown sanitizer package when it is available in your
target Gno version. Community markdown builders can still be useful, but check
whether each helper sanitizes internally or expects the caller to sanitize first.

## Storage Helpers

- [`gno.land/p/moul/ulist`](../../examples/gno.land/p/moul/ulist/ulist.gno):
  append-oriented list storage with range and offset iteration.
- [`gno.land/p/moul/addrset`](../../examples/gno.land/p/moul/addrset/addrset.gno):
  address set semantics.
- [`gno.land/p/moul/fifo`](../../examples/gno.land/p/moul/fifo/fifo.gno):
  queue-like storage.
- [`gno.land/p/moul/collection`](../../examples/gno.land/p/moul/collection/collection.gno):
  indexed collection patterns built on tree storage and `seqid`.

Use these as references for access-pattern-specific storage, not as a reason to
skip your own review. If a package almost fits but has unclear semantics,
consider improving it or publishing a clearer alternative.

For append-only feeds, a list helper can replace a hand-rolled `nextID` plus
tree pattern:

```go
import "gno.land/p/moul/ulist"

var posts = ulist.New()

func AddPost(post Post) {
    posts.Append(post)
}

func RecentPosts(offset, count int) []ulist.Entry {
    return posts.GetByOffset(offset, count)
}
```
