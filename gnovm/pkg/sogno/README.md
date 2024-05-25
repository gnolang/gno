# Sogno

Sogno is a runtime for executing Gno code as compiled Go code. The goal is to
provide the features of the GnoVM, without a virtual machine.

The binaries are independent programs, which interact with the outside world
primarily through the use of their interface through stdin and stdout.

---

## Notes

- Changing the value of an exported package should be forbidden.
	- This is sane, allows to have things like error values as constants, and
		marks that they are "pure" packages.
- All `int` and `uint` values should be converted to 64-bit.

## This is in MVP-mode

So here are some things which are not ideal, and should be fixed for real
production use.

- This package should try to add a very minimal overhead in terms of binary size
	to the resulting binary. Consequently, we should try to avoid some standard
	libraries which add a lot of overhead, like the `unicode` package (and ones
	that depend on it, like `strings` and `reflect`) and the `time` package (and
	ones like `os`).
- List evolving...

## Value representation

- `var myGlobal T` are directly loaded
	- The runtime gets the stored variables at startup, and those can be
		re-fetched later.
- type `*T` => `sogno.Pointer[T]`
	- `type Pointer[T any] struct { ObjectID string; V *T }`.
	- If ObjectID is set, the object is not loaded; if it's "", it is.
		- map[string]uintptr to map object ID and loaded pointers.
	- Capable of loading cross-realm, even of not directly imported realms
	- Values obtained from other realms are read-only, but their methods may be
		called.
	- The same goes for slices and maps, which because they have an "implicit
		pointer", should be lazily loaded.
- Closures
	- Any closure is simplified to a combination of a struct type with the
		closure's local variables + a `Call` method. Closures without local
		variables are simple functions.
- `sogno.Borrow` for pointers, maps and slices whose underlying values are
	borrowed from other realms. (Makes sure they are read-only, and if anything
	are called using methods.)
- Interface values => `sogno.TypedValue`
	- `type TypedValue struct { TypeID string; V interface{} }`
	- This is tricky. An interface type may be an "owned" type, or it may be an
		non-owned type.

Imports to other realms / transitive packages are processed to get other types.

### Then

- `float`: enforce software float implementation.
- `map[T]T1`: binary tree.
- Optimize realms that don't have storage.
- To improve compile-time: make it zero-dependency, possibly with own GOROOT and
	runtime, so that the source code is as minimal as possible, and we can
	better take advantage of packages like `reflectlite`.
