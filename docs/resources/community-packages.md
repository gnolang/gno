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

Start with the official markdown sanitizer for untrusted text when your target
Gno version provides it. See [Raw public text in `Render`](./gno-security-guide.md#59-raw-public-text-in-render)
for the recommended `gno.land/p/nt/markdown/sanitize/v0` shape:

```go
import "gno.land/p/nt/markdown/sanitize/v0"

func Render(path string) string {
    return "# Echo\n\n" + sanitize.InlineText(path)
}
```

Community markdown builders are still useful for composing output after you
know whether each helper sanitizes internally or expects sanitized input.

- [`gno.land/p/moul/md`](../../examples/gno.land/p/moul/md/md.gno): helpers for
  building markdown links, headings, lists, images, code blocks, and text
  escaping.

  ```go
  import "gno.land/p/moul/md"

  func Render(_ string) string {
      return md.H1("Tasks") + md.TodoList([]string{"review", "ship"}, []bool{true, false})
  }
  ```

- [`gno.land/p/moul/mdtable`](../../examples/gno.land/p/moul/mdtable/mdtable.gno):
  helpers for markdown tables. Use this when your `Render` output is tabular
  and you want pipe escaping handled consistently.

  ```go
  import "gno.land/p/moul/mdtable"

  func Render(_ string) string {
      table := mdtable.Table{Headers: []string{"ID", "Status"}}
      table.Append([]string{"1", "open"})
      return table.String()
  }
  ```

- [`gno.land/p/nt/mdalert/v0`](../../examples/gno.land/p/nt/mdalert/v0/README.md):
  helpers for Gno-Flavored Markdown alert blocks.

- [`gno.land/p/sunspirit/md`](../../examples/gno.land/p/sunspirit/md/md.gno):
  a builder-oriented markdown package. It is convenient when a view is assembled
  from optional fragments.

  ```go
  import "gno.land/p/sunspirit/md"

  func Render(_ string) string {
      return md.NewBuilder().Add(md.H1("Profile"), md.Bold("active")).Render("\n")
  }
  ```

## Storage Helpers

- [`gno.land/p/moul/ulist`](../../examples/gno.land/p/moul/ulist/ulist.gno):
  append-oriented list storage with range and offset iteration.
- [`gno.land/p/moul/addrset`](../../examples/gno.land/p/moul/addrset/addrset.gno):
  address set semantics.
- [`gno.land/p/moul/fifo`](../../examples/gno.land/p/moul/fifo/fifo.gno):
  queue-like storage.
- `gno.land/p/moul/collection`:
  indexed collection patterns built on tree storage and `seqid`.
- [`gno.land/p/nt/avl/v0`](../../examples/gno.land/p/nt/avl/v0/README.md):
  general sorted key/value indexes with range and offset iteration.
- [`gno.land/p/nt/bptree/v0`](../../examples/gno.land/p/nt/bptree/v0/doc.gno):
  B+ tree storage variants for large sorted datasets and scan-heavy indexes.
- [`gno.land/p/nt/seqid/v0`](../../examples/gno.land/p/nt/seqid/v0/README.md):
  sequential IDs encoded so they sort correctly as AVL keys.
- `gno.land/p/jeronimoalbi/bitset`:
  compact membership flags when IDs are dense numeric indexes.

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

For ordered tree keys, combine `avl.Tree` with `seqid` instead of stringified
integers:

```go
import (
    "gno.land/p/nt/avl/v0"
    "gno.land/p/nt/seqid/v0"
)

var ids seqid.ID
var posts avl.Tree

func AddPost(post *Post) {
    posts.Set(ids.Next().Binary(), post)
}
```

For compact allowlists over numeric IDs, a bitset can be cheaper than storing
one map or tree entry per flag:

```go
import "gno.land/p/jeronimoalbi/bitset"

var claimed = bitset.New(1024)

func MarkClaimed(id uint64) {
    claimed.Set(id)
}

func IsClaimed(id uint64) bool {
    return claimed.IsSet(id)
}
```

## Access-Control Helpers

- [`gno.land/p/nt/ownable/v0`](../../examples/gno.land/p/nt/ownable/v0/README.md):
  owner-gated administration with explicit `cur realm` checks.
- `gno.land/p/nt/pausable/v0`:
  pause switches layered on an `ownable.Ownable`.
- [`gno.land/p/moul/authz`](../../examples/gno.land/p/moul/authz/authz.gno):
  authorization helper patterns worth studying when a single owner is not
  enough.

Use these packages to reduce repeated authorization code, but keep the realm's
identity model explicit. A realm method should still pass its own captured `cur`
into the helper:

```go
import "gno.land/p/nt/ownable/v0"

var owner = ownable.NewWithAddress("g1...")

func SetName(cur realm, name string) {
    if !cur.IsCurrent() {
        panic("invalid realm")
    }
    owner.AssertOwnedBy(cur.Previous().Address())
    displayName = name
}
```

## Structured Data Helpers

- [`gno.land/p/onbloc/json`](../../examples/gno.land/p/onbloc/json/README.md):
  parse JSON into explicit nodes when a realm accepts structured text input.
- [`gno.land/p/onbloc/int256`](../../examples/gno.land/p/onbloc/int256/doc.gno)
  and [`gno.land/p/onbloc/uint256`](../../examples/gno.land/p/onbloc/uint256/README.md):
  fixed-width integer helpers for domains that need larger arithmetic than the
  built-in integer types.
- `gno.land/p/lou/query`:
  query-string parsing helpers for `Render(path)` and URL-like inputs.

Prefer explicit parsing over ad-hoc string splitting when the input format has
nesting, quoting, or escaping rules:

```go
import "gno.land/p/onbloc/json"

func DecodeTitle(input string) (string, bool) {
    root, err := json.Unmarshal([]byte(input))
    if err != nil {
        return "", false
    }
    title, err := root.GetKey("title")
    if err != nil || !title.IsString() {
        return "", false
    }
    value, err := title.GetString()
    return value, err == nil
}
```

## Application Patterns

- `gno.land/p/agherasie/forms`:
  form creation, typed answers, deadlines, and validation.
- `gno.land/p/lou/blog`:
  a larger package split across posts, comments, moderation, rendering, and
  query helpers.
- `gno.land/p/morgan/chess`:
  domain-heavy package structure with tests around game rules.

Use these as source-reading material for package boundaries and tests, not as
drop-in frameworks. A good reuse decision should identify the one piece you need
instead of copying an entire application shape.
