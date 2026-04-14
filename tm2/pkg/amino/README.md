# Amino

https://github.com/gnolang/gno/tree/master/tm2/pkg/amino

NOTE: This project used to be gnolang/gno/pkgs/amino, derived from
tendermint/go-amino.

Amino is an encoding/decoding library for structures.

In Amino, all structures are declared in a restricted set of Go.

From those declarations, Amino generates Protobuf3 compatible binary bytes.

For faster encoding and decoding, Amino also generates Protobuf3 schemas, and
also generates translation logic between the original Go declarations and
Protobuf3/protoc generated Go.  In this way, Amino supports a restricted set
of Protobuf3.

Though Amino supports a subset of Protobuf3 and uses it to optimize encoding
and decoding, it is NOT intended to be a Protobuf3 library -- complete support
of Protobuf3 is explicitly not its design goal.

You can see the performance characteristics of the improvements in [XXX -- this
exists, but i forget the filename.].

The gist is that in recent versions of Go, the reflection-based binary
encoding/decoding system is about 3x slower than the protoc generated Go ones,
and that the translation system works pretty well, accounting for only ~25% of
the total time, so the performance hit isn't fatal.

While creating and finalizing this library, which I believe it is, roughly, the
final form of a well structured piece of software according to my tastes, it
occurred to me that Proto3 is a complexification that is not needed in well
written software, and that it mostly serves systems that are part of the mass
public surveillance database amassed in datacenters across the world.  Somewhere,
there is a proto3 field number and value that describes some new aspect about me,
in several instances of Google's massive database.

What I want instead is a language, and for that, the original implementation
of Amino is better suited.

# Amino JSON

This is experimental and subject to change.

## Amino in the Wild

* Amino:binary spec in [Tendermint](https://github.com/tendermint/tendermint/blob/main/spec/core/encoding.md)


# Amino Spec

#### Registering types and packages

Previous versions of Amino used to require a local codec where types must be
registered.  With the change to support Any and type URL strings,
we no longer need to keep track of local codecs, unless we want to override
default behavior from global registrations.

Each package should declare in a package local file (by convention called amino.go)
which should look like the following:

```go
// see github.com/gnolang/gno/tm2/pkg/amino/protogen/example/main.go
package main

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule"
)

var Package = amino.RegisterPackage(
	amino.NewPackage(
		"main", // The Go package path
		"main", // The (shorter) Proto3 package path (no slashes).
		amino.GetCallersDirname(),
	).WithDependencies(
		submodule.Package, // Dependencies must be declared (for now).
	).WithTypes(
		StructA{}, // Declaration of all structs to be managed by Amino.
		StructB{}, // For example, protogen to generate proto3 schema files.
		&StructC{}, // If pointer receivers are preferred when decoding to interfaces.
	),
)
```

You can still override global registrations with local `*amino.Codec` state.
This is used by `genproto.P3Context`, which may help development while writing
migration scripts.  Feedback welcome in the issues section.

## Unsupported types

### Floating points
Floating point number types are discouraged as [they are generally
non-deterministic](https://gafferongames.com/post/floating_point_determinism/).
If you need to use them, use the field tag `amino:"unsafe"`.

### Enums
Enum types are not supported in all languages, and they're simple enough to
model as integers anyways.

### Maps
Maps are not currently supported.  There is an unstable experimental support
for maps for the Amino:JSON codec, but it shouldn't be relied on.  Ideally,
each Amino library should decode maps as a List of key-value structs (in the
case of languages without generics, the library should maybe provide a custom
Map implementation).  TODO specify the standard for key-value items.

## Amino and Proto3

Amino objects are a subset of Proto3.
* Enums are not supported.
* Nested message declarations are not supported.

Amino extends Proto3's Any system with a particular concrete type
identification format (disfix bytes).

## Amino and Go

Amino objects are a subset of Go.
* Multi-dimensional slices/arrays are not (yet) supported.
* Floats are nondeterministic, so aren't supported by default.
* Complex types are not (yet) supported.
* Chans, funcs, and maps are not supported.
* Pointers are automatically supported in go-amino but it is an extension of
  the theoretical Amino spec.

Amino, unlike Gob, is beyond the Go language, though the initial implementation
and thus the specification happens to be in Go (for now).

## Limitations

* Pointer types in arrays and slices lose pointer information.
* Nested pointers are not allowed.
* Recursive ReprType not allowed.
