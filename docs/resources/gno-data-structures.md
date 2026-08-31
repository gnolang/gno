# Gno Data Structures

Gno supports the same basic data structures as Go. This guide covers how to use
them in your [realms](./realms.md) and [packages](./gno-packages.md).

## Quick Reference

| Type | Example | Best For |
|------|---------|----------|
| **[Array](#arrays)** | `[5]int` | Fixed-size collections |
| **[Tree-backed index](#tree-backed-indexes)** | `avl.Tree`, `bptree.BPTree` | Large/growing sorted datasets |
| **[Map](#maps)** | `map[string]int` | Small key-value stores |
| **[Slice](#slices)** | `[]string` | Dynamic lists |
| **[Struct](#structs)** | `type User struct{...}` | Grouped data |
| **[Pointer](#pointers)** | `*User` | Reference values |

## Arrays

Fixed-size collections where the size is set at compile time.

```go
var numbers [5]int
primes := [5]int{2, 3, 5, 7, 11}

// Access elements
numbers[0] = 1
value := numbers[0]
```

Arrays are copied when passed to functions. Use pointers to modify them: `func update(arr *[5]int)`.

## Tree-backed Indexes

For large or growing sorted datasets, prefer a tree-backed index over a
persistent map. Tree implementations store nodes or leaf pages separately, so
reading or writing one key does not require loading the whole collection.

Common choices include:

| Type | Good fit | Tradeoffs |
|------|----------|-----------|
| [`avl.Tree`](../../examples/gno.land/p/nt/avl/v0/README.md) | General sorted key/value indexes, range scans, offset pagination | `O(log n)` lookup, values are `any`, keys are strings |
| [`bptree.BPTree`](../../examples/gno.land/p/nt/bptree/v0/doc.gno) | Large sorted indexes where higher fanout and fewer pointer dereferences help | More tuning surface; choose fanout intentionally when the default is not enough |
| [`avl`](../../examples/gno.land/p/nt/avl/v0/README.md) or [`bptree/list`](../../examples/gno.land/p/nt/bptree/v0/list) | List-like APIs backed by tree storage | Still design keys and pagination around your product |

```go
import "gno.land/p/nt/avl/v0"

var users avl.Tree

// Set
users.Set("alice", User{Score: 100})

// Get (returns nil if key not found)
value := users.Get("alice")
if value != nil {
    user := value.(User) // Type cast required
}

// Has (check if a key exists without retrieving the value)
// Note: a key stored with a nil value is indistinguishable from an
// absent key when using Get; use Has to check for existence.
if users.Has("alice") {
    // key exists
}

// Iterate in sorted order
users.Iterate("", "", func(key string, value any) bool {
    user := value.(User)
    println(key, user.Score)
    return false // continue iterating
})
```

Use an explicit ID helper for append-like records. `seqid.ID` generates keys
that preserve numeric order when stored in a tree:

```go
import (
    "gno.land/p/nt/avl/v0"
    "gno.land/p/nt/seqid/v0"
)

var (
    nextPostID seqid.ID
    posts      avl.Tree // seqid string -> Post
)

func AddPost(post Post) string {
    id := nextPostID.Next().String()
    posts.Set(id, post)
    return id
}
```

Add secondary indexes for each lookup path users need:

```go
var (
    postsByID     avl.Tree // id -> Post
    postsByAuthor avl.Tree // author + "/" + id -> id
)
```

**Learn more:** [Effective Gno: Choose storage types by access pattern](./effective-gno.md#choose-storage-types-by-access-pattern)

For non-official storage helpers such as unique lists, sets, and queues, see
[Community Packages](./community-packages.md).

## Maps

Key-value stores with O(1) lookup.

```go
scores := make(map[string]int)

// Set
scores["alice"] = 100

// Get with existence check
score, exists := scores["alice"]

// Delete
delete(scores, "alice")

// Iterate
for username, score := range scores {
    println(username, score)
}
```

**Note**: In Gno, map iteration order follows insertion order, unlike Go which uses
randomized iteration order due to underlying C hashmap implementation.
While this makes Gno behavior deterministic, you should still not rely on
iteration order for correctness or public `Render` output. Use an explicit
ordered list or tree-backed index when users depend on stable ordering,
pagination, or range queries.

## Slices

Dynamic, growable lists.

```go
var users []string
users = append(users, "alice")
users = append(users, "bob")

// Iterate
for i, user := range users {
    println(i, user)
}

// Remove element at index
users = append(users[:index], users[index+1:]...)
```

Pre-allocate capacity for better performance: `make([]string, 0, 100)`.

Slices hold references to underlying backing arrays; so modifying
a value in a slice is like modifying it in a pointer to an array. This is
particularly relevant for cross-realm interactions: the elements of
a slice of another realm are *references*, not values.

To copy a slice, use append on a nil slice:

```go
users = append([]string(nil), otherSlice...)
```

## Structs

Group related data together.

```go
type User struct {
    Name  string
    Score int
}

// Create
user := User{Name: "alice", Score: 100}

// Methods with pointer receiver (can modify)
func (u *User) IncrementScore() {
    u.Score++
}

// Methods with value receiver (read-only)
func (u User) Display() string {
    return u.Name + ": " + strconv.Itoa(u.Score)
}
```

## Pointers

Reference values instead of copying them.

```go
x := 42
ptr := &x    // Get address
*ptr = 100   // Modify through pointer (x is now 100)

// With structs
func UpdateScore(u *User, score int) {
    u.Score = score // Modifies original
}

user := &User{Name: "alice"}
UpdateScore(user, 100)
```

Always check for nil: `if ptr == nil { return }`.

## Persistence in Realms

Global variables in [realms](./realms.md) are automatically saved between transactions.
This is a key feature of Gno's [automatic state management](./gno-memory-model.md).

```go
var (
    counter int              // Single value
    users   []string         // Entire slice
    scores  map[string]int   // Entire map
    tree    avl.Tree         // Only modified nodes
)
```

**Learn more:**
- [Gno Memory Model](./gno-memory-model.md) - How data is stored
- [Effective Gno](./effective-gno.md) - Best practices and detailed examples
- [Go-Gno Compatibility](./go-gno-compatibility.md) - Differences from Go
