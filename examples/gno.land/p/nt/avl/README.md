# AVL Tree Package

The `avl` package provides a gas-efficient AVL tree implementation for storing key-value data in Gno realms.

## Basic Usage

```go
package myrealm

import "gno.land/p/nt/avl"

// This AVL tree will be persisted after transaction calls
var tree *avl.Tree

func Set(key string, value int) {
	// tree.Set takes in a string key, and a value that can be of any type
	tree.Set(key, value)
}

func Get(key string) int {
	// tree.Get returns the value at given key in its raw form,
	// and a bool to signify the existence of the key-value pair
	rawValue, exists := tree.Get(key)
	if !exists {
		panic("value at given key does not exist")
	}

	// rawValue needs to be converted into the proper type before returning it
	return rawValue.(int)
}
```

## Storage Architecture: AVL Tree vs Map

In Gno, the choice between `avl.Tree` and `map` is fundamentally about how data is persisted in storage.

**Maps** are stored as a single, monolithic object. When you access *any* value in a map, Gno must load the *entire* map into memory. For a map with 1,000 entries, accessing one value means loading all 1,000 entries.

**AVL trees** store each node as a separate object. When you access a value, Gno only loads the nodes along the search path (typically log2(n) nodes). For a tree with 1,000 entries, accessing one value loads ~10 nodes; but a tree with 1,000,000 entries only needs to load ~20 nodes.

## Storage Comparison Example

Consider a realm with 1,000 key-value pairs. Here's what happens when you access a single value:

**Map storage:**

```
Object :4 = map{
  ("0" string):("123" string),
  ("1" string):("123" string),
  ...
  ("999" string):("123" string)
}
```
- Accessing `map["100"]` loads object `:4` (contains **all 1,000 pairs**)
- Gas cost is proportional to total map size (1,000 entries)
- **1 object fetch, but massive data load**

**AVL tree storage:**

```
Object :6 = Node{key="4", height=10, size=1000, left=:7, right=...}
Object :9 = Node{key="2", height=9, size=334, left=:10, right=...}
Object :11 = Node{key="14", height=8, size=112, left=:12, right=...}
Object :13 = Node{key="12", height=6, size=46, left=:14, right=...}
Object :15 = Node{key="11", height=5, size=24, left=:16, right=...}
Object :17 = Node{key="102", height=4, size=13, left=:18, right=...}
Object :19 = Node{key="100", height=3, size=5, left=:30, right=...}
Object :31 = Node{key="101", height=1, size=2, left=:32, right=...}
Object :33 = Node{key="100", value="123", height=0, size=1}
```
- Accessing `tree.Get("100")` loads ~10 objects (the search path)
- Gas cost is proportional to log2(n) ≈ 10 nodes
- **10 object fetches, each containing only a single node**

## Further Reading

- [Why should you use an AVL tree instead of a map?](https://howl.moe/posts/2024-09-19-gno-avl-over-maps/) - Howl detailed analysis
- [Berty's AVL scalability report](https://github.com/gnolang/hackerspace/issues/67) - Real-world testing with up to 20M entries
- [Wikipedia - AVL tree](https://en.wikipedia.org/wiki/AVL_tree) - Algorithm details and balancing
- [Effective Gno](https://docs.gno.land/resources/effective-gno#prefer-avltree-over-map-for-scalable-storage) - High-level usage guidance
