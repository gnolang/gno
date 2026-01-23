# `avl` - AVL Tree Implementation

A self-balancing binary search tree implementation with AVL (Adelson-Velsky and Landis) properties, providing efficient operations with guaranteed O(log n) time complexity.

## Features

- **Self-balancing**: Maintains height balance automatically
- **Generic values**: Stores any type using `any` interface
- **String keys**: Uses string keys for ordering
- **Iterator support**: Forward and reverse iteration with callbacks
- **Index access**: Get elements by index position
- **Range operations**: Iterate over key ranges

## Usage

```go
// Create new AVL tree
tree := avl.NewTree()

// Insert key-value pairs
tree.Set("apple", "red fruit")
tree.Set("banana", "yellow fruit")
tree.Set("cherry", "red fruit")

// Check if key exists
exists := tree.Has("apple") // true

// Get value by key
value, exists := tree.Get("banana") // "yellow fruit", true

// Get by index (sorted order)
key, value := tree.GetByIndex(0) // "apple", "red fruit"

// Size
count := tree.Size() // 3

// Remove
removed := tree.Remove("banana") // "yellow fruit", true

// Iterate over all items
tree.Iterate("", "", func(key string, value any) bool {
    // Process each key-value pair
    return false // return true to stop iteration
})

// Iterate over range
tree.Iterate("a", "c", func(key string, value any) bool {
    // Only processes keys from "a" to "c"
    return false
})
```

## API

```go
type ITree interface {
    // Query operations
    Size() int
    Has(key string) bool
    Get(key string) (value any, exists bool)
    GetByIndex(index int) (key string, value any)
    
    // Iteration
    Iterate(start, end string, cb IterCbFn) bool
    ReverseIterate(start, end string, cb IterCbFn) bool
    IterateByOffset(offset int, count int, cb IterCbFn) bool
    ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool
    
    // Modification
    Set(key string, value any) (updated bool)
    Remove(key string) (value any, removed bool)
}

type IterCbFn func(key string, value any) bool
```

**Time Complexity:**
- Insert/Update: O(log n)
- Search: O(log n)
- Delete: O(log n)
- Iteration: O(k) where k is number of items iterated

**Space Complexity:** O(n)

## Sub-packages

- `avl/list` - List operations on AVL trees
- `avl/pager` - Pagination utilities for AVL trees
- `avl/rolist` - Read-only list interface
- `avl/rotree` - Read-only tree interface
