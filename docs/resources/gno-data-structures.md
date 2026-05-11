# Gno Data Structures

Gno supports the same basic data structures as Go. This guide covers how to use
them in your [realms](./realms.md) and [packages](./gno-packages.md).

## Quick Reference

| Type | Example | Best For |
|------|---------|----------|
| **[Array](#arrays)** | `[5]int` | Fixed-size collections |
| **[AVL Tree](#avl-trees)** | `avl.Tree` | Large/growing datasets |
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

## AVL Trees

For large or growing datasets, prefer [`avl.Tree`](../../examples/gno.land/p/nt/avl/v0/README.md)
over maps as it is significantly more efficient in both [gas](./gas-fees.md)
cost and runtime performance.

```go
import "gno.land/p/nt/avl/v0"

var users avl.Tree

// Set
users.Set("alice", User{Score: 100})

// Get
value, exists := users.Get("alice")
if exists {
    user := value.(User) // Type cast required
}

// Iterate in sorted order
users.Iterate("", "", func(key string, value any) bool {
    user := value.(User)
    println(key, user.Score)
    return false // continue iterating
})
```

**Learn more:** [Effective Gno: Prefer avl.Tree over map](./effective-gno.md#prefer-avltree-over-map-for-scalable-storage)

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
iteration order for correctness to maintain compatibility with Go semantics.

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
