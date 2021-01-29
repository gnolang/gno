# Gno: a Language for Resistance.

 * Like interpreted Go, but more ambitious.
 * Completely deterministic, for complete accountability.
 * Transactional persistence across data realms.

## Ownership 

In Gno, all objects are automatically persisted to disk after every atomic
"transaction" (a function call that must return immediately.) when new objects
are associated with a "ownership tree" which is maintained overlaying the
possibly cyclic object graph.  The ownership tree is composed of objects
(arrays, structs, maps, and blocks) and derivatives (pointers, slices, and so
on) with struct-tag annotations to declare the ownership tree.

If an object hangs off of the ownership tree, it becomes included in the Merkle
root, and is said to be "real".  The Merkle-ized state of reality gets updated
with state transition transactions; during such a transaction, some new
temporary objects may "become real" by becoming associated in the ownership
tree (say, assigned to a struct field or appended to a slice that was part of
the ownership tree prior to the transaction), but those that don't get garbage
collected and forgotten.

We get a lack-of-owner problem when the ownership tree detaches an object
referred elsewhere (after running a statement or set of statements):

```
    A, A.B, A.C, and D are owned objects.
    D doesn't own C but refers to it.

	   (A)   (D)
	   / \   ,
	  /   \ ,
	(B)   (C)

    > A.C = nil

	   (A)       (D)
	   / \       ,
	  /   \     ,
	(B)    _   C <-- ?
```

Options:

 1. unaccounted object error (default)
   - can't detach object unless refcount is 0.
   - pushes problem to ownership tree manipulation logic.
   - various models are possible for maintaining the ownership tree, including
     reference-counting (e.g. by only deleting objects from a balanced search
tree when refcount reaches 1 using a destroy callback call); or more lazily
based on other conditions possibly including storage rent payment.
   - unaccounted object error detection is deferred until after the
     transaction, allowing objects to be temporarily unaccounted for.

 2. Invalid pointer
   - basically "weak reference pointers" -- OK if explicit and exceptional.
   - across realms it becomes a necessary construct.
   - will implement for inter-realm pointers.

 3. Auto-Inter-Realm-Ownership-Transfer (AIR-OT)
   - within a realm, refcounted garbage collection is sufficient.
   - automatic ownership transfers across realms may be desirable.
   - requires bidirectional reference tracking.
