# PRxxxx: Type Identity Checks

## Context

Gno uses `TypeID()` in many places as a deterministic structural type key. That
key intentionally ignores struct tags, which is correct for explicit
conversions: Go permits conversions between struct types that differ only in
tags.

The same key was also used for assignment, comparison, function signatures, and
interface method matching. Those contexts require Go type identity, where
struct tags, embedded field syntax, map element types, and variadic function
signatures matter. This let Gno accept programs that Go rejects, such as:

- assigning `struct{A int "b"}` to `struct{A int "a"}`;
- assigning `map[int]int` to `map[int]string`;
- assigning or converting `func([]int)` to `func(...int)`;
- treating `struct{T}` and `struct{T T}` as the same type.

## Decision

Add a separate type identity helper instead of changing `TypeID()`:

- `identicalTypes` implements exact identity for assignment, comparison,
  function/method signatures, and debug assertions.
- `identicalTypesIgnoreTags` implements the conversion rule that ignores struct
  tags but still respects the rest of type identity.

The assignment checker now uses exact identity for pointer elements, arrays,
slices, maps, structs, functions, compound assignment, and map element types.
Conversion preprocessing keeps the tag-ignoring rule so valid struct-tag
conversions continue to work.

## Alternatives Considered

1. **Change `TypeID()` to include tags and variadic signatures.** This would fix
   assignment but break valid conversions that intentionally ignore struct tags.

2. **Patch only struct assignment.** That would leave the same bug in pointers,
   maps, interfaces, function signatures, comparisons, and conversions.

3. **Keep using `TypeID()` and special-case tags.** This would not address other
   identity differences, especially embedded fields and variadic functions.

## Consequences

- Gno rejects more invalid programs at preprocessing/type-check time, matching
  Go identity rules more closely.
- Explicit struct conversions that differ only by tags remain valid.
- `TypeID()` remains a deterministic key for existing conversion/storage paths;
  stricter identity checks are opt-in at call sites that need Go identity.
- Filetests cover struct tags, embedded fields, map element types, variadic
  function assignment/conversion, interface method matching, and struct
  comparison.
