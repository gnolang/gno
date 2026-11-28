# Charge CPU gas for every element a value copy duplicates or a zero value fills

## Context

Assigning an array or a struct by value copies it through
`TypedValue.Copy`, which calls `ArrayValue.Copy` or `StructValue.Copy`. Each
walks every element or field and copies it, recursing into nested arrays and
structs. None of the three charged CPU gas: the copy paid only the allocation
gas of the new value and the flat cost of the op that asked for it.

`c := *src` for a `[1000]int` therefore cost about 7.5 gas per element, all of
it allocation gas, while the walk itself ran free. The `gas/copy_array_*`
filetests measure it: declaring a `[1000]int` and copying it ten times costs
112,922 gas before this change, and the same with a `[2000]int` costs 194,822.

`OpCPUSlopeCopyElement` (40 per element) and `OpCPUSlopeCopyPrimitive` (4 per
byte) were calibrated on `unrefCopy` and `copyDataToList`, and their comments
named the assignment path, but only the `append` and `copy` builtins charged
them.

Building a zero value walks the same way. `defaultArrayValue` writes each
element of an array one at a time, `defaultStructFields` each field of a
struct, and `make([]T, n)` each element of the new slice, for every `T` that
is not an interface or a byte. They are reached from `new`, `var`, `make`,
named results, a missing map key, a failed type assertion and the elements an
array literal or a keyed struct literal leaves out. None charged CPU gas per
element either: 100 calls of `new([1000000]int)` cost 591,438,269 gas,
about 5.9 per element, all of it allocation gas.

`ArrayValue.Copy` and `StructValue.Copy` also left the list or field buffer
they were filling out of the allocator's count. A collection triggered while
a nested element was copied dropped the elements already copied, so
`gas/compute_map_key_big_struct.gno` held two 40 MB copies under its 50 MB
cap. `defaultArrayValue` and `defaultStructFields` already anchor their buffer
for this reason.

## Decision

`ArrayValue.Copy` charges `OpCPUSlopeCopyElement` per element of a list-backed
array and `OpCPUSlopeCopyPrimitive` per byte of a byte array.
`StructValue.Copy` charges `OpCPUSlopeCopyElement` per field. A nested array or
struct is charged again by its own `Copy`, so the total follows the number of
values duplicated. Since `pr6504_copy_stored_value_children.md`, `Copy` loads a
stored child held as a `RefValue` and copies it through the same methods, so
its elements pay too.

The charge sits in `Copy` itself, so every caller pays it: assignment and
definition, call arguments, results copied into named result slots, `range`
over an array, conversions, composite literal elements, value receivers and
map keys. `Copy` holds an `*Allocator` and no `*Machine`, so the charge goes
through the allocator's gas meter, which `NewMachineWithOptions` sets to the
machine's own, and it lands before the new value is allocated. A nil
allocator, or one with no meter, charges nothing, as its allocation gas
already does. An assignment or declaration to `_` copies nothing: `Assign2`
returns before the copy, since nothing reads the blank slot. A range over a
map neither loads nor copies a key or value bound to `_`, as a range over a
list already skipped a blank value.

`range` over an array by value already charged `OpCPUSlopeRangeIterArray`,
15 per element, whose fit times that same copy. The slope is removed: the copy
is the only part of a range's setup whose cost grows with the length, and
`Copy` now charges it. Ranging through a pointer or over a slice copies
nothing and paid the 15 for no work: 1,000 ranges that break at once take the
same time over 1,000 elements as over 100,000.

Building a zero value charges `OpCPUSlopeCopyElement` for each element or
field written one at a time. `fillDefaultList` and `fillDefaultStructFields`
count the slots still undefined and charge for them after the buffer is
allocated and before any is written, so `Allocate` refuses an oversized
buffer itself and a fill out of gas writes nothing. They serve
`defaultArrayValue`, `defaultStructFields`, `make([]T, n)` per element of its
length, or of its capacity when one is given, and the literals. A byte
array's `Data` and an interface slot are left to allocation gas: Go's `make`
hands them back zeroed, and no loop writes them, as with `make([]byte, n)`
before this change. An array or struct literal copies the elements or fields
it names into a buffer left undefined, then fills only the rest, so it pays
`OpCPUSlopeArrayLit` or `OpCPUSlopeStructLit` for each one named and the
zero-fill for each one missing. A complete `[3]int{1, 2, 3}` costs what it
cost before, and a named element of array or struct type no longer pays for
the nested default the old fill built only to overwrite. For the same reason
`m[k] = v` on a new key no longer gives the entry a zero value first when the
map holds arrays or structs: only an assignment writes such an entry, `m[k]
= v` or a range clause assigning to `m[k]`, and its `Assign2` overwrites the
slot at once. Any other value type still gets
its zero value, which `m[k] += 1` reads.

`var a [N]int` keeps both charges, the fill and the copy `Assign2` makes of
it. The copy is not only work: `StructValue.Copy` stamps the struct with its
declared type's realm, where the zero value carries the stamp of its unnamed
base type, so storing the zero value itself would change which realm may write
the variable.

`ArrayValue.Copy` and `StructValue.Copy` anchor the buffer they fill, with
`PushAnchor` and `PopAnchor`, as `defaultArrayValue` does. `StructValue.Copy`
pops after `NewStruct`, which allocates, per the anchor's rule that nothing
allocates between the pop and the moment the owner is reachable.
`defaultStructValue` now holds its anchor across `NewStruct` the same way,
and `new(T)` anchors its zero value across the heap item and pointer type it
allocates before the result is pushed. `make([]T, n)` filled its list with no
anchor at all, so every collection during the fill dropped the elements
already built: under a 1 MB cap `make([][1000]int, 100)` completed where `var
a [100][1000]int` was refused, and gas was the only bound left, about 2.5 GB
for a 3e9-gas transaction against a 500 MB cap. It anchors the list through
the fill and `NewSlice` (`alloc_26_make_len.gno`, `alloc_27_make_cap.gno`).
`append` into a new list copied its elements into an unanchored list the same
way, so `append([][1000]int(nil), src...)` held a second copy of `src` under a
cap that refuses `make` followed by `copy` (`alloc_28_append_copy.gno`); its
list branches now anchor the list through the copies and `NewSlice`, the
byte branches of `make` and `append` anchor their data array across
`NewSlice`, and `append(b, s...)` anchors the byte copy it makes of `s` until
it returns.

These anchors cover a value while it is built.

A three-clause `for` loop gives each iteration its own copy of a variable its
init declares, as Go 1.22 does, once that variable lives on the heap. The heap
item holding it is replaced before the post statement, and the new one gets a
`Copy` of the value, where it used to share the old one's array or struct, so
a pointer or closure kept from one iteration read the last iteration's value.
`&a`, a closure and a pointer method value move a variable to the heap; so
does slicing an array-valued init variable, `a[:]` or `s.buf[:]`, which
aliases it as `&a` does. Only an init variable moves for a slice, so `var buf
[32]byte; f(buf[:])` elsewhere keeps its block slot, and only when the array
sits inside it, reached through array elements and struct fields: `s.sl[0][:]`
slices a slice's backing array and leaves `s` in its slot
(`gas/loopvar_init_slice_through_slice.gno`). Whether the slice outlives the
iteration does not matter, as for `&a`: the interpreter cannot see whether a
callee keeps it, so `n += len(a[:])` in the body costs the copy too. `_` in
the init takes no block slot and is not counted, so the copy covers exactly
the names declared: counting it ran the copy one slot past the init
variables, which panicked with an index out of range on an empty body and
otherwise copied a variable the body declares (`loopvar_init_slice_copy.gno`,
`loopvar_init_blank.gno`).

`len` and `cap` of an array, or of a pointer to one, fold to a constant at
preprocess when the operand holds no call or function literal, the case in
which Go's spec makes them constant and leaves the operand unevaluated. A call
passed `len` its argument by value, so `len(a)` copied the whole array each
time and, with the copy charged, a loop `for i := 0; i < len(a); i++` over a
`[1000]int` cost 50,745,267 gas. `pr6504_copy_stored_value_children.md` now
passes `len` and `cap` their argument without a copy, and the call alone
still costs that loop 4,019,590; folded it costs 1,131,329, measured the same
way and pinned by `gas/len_array_loop.gno`, the same as hoisting `len(a)`
into a constant. An
operand that would panic if evaluated, `a[i]` with `i` out of range or `*p`
with `p` nil, now yields the length, as in Go. The folded operand still counts
as a reference for package initialization order, as in Go. A `ConstExpr`
hides its source from the transcriber, so `codaInitOrderDeps` walks the
source of every folded constant except a bare const name, which reaches the
operand also inside a larger fold such as `len(a)-1`
(`len_array_initorder.gno`); const decls are left out, as before. `len` and
`cap` of a nil pointer to a named array type no longer panic at run time
either: `GetLength` and `GetCapacity` take the element's base type. A
conversion is not a call, so `len(T(a))` and `len(T(*p))` fold too, while a
call inside the conversion's operand still keeps the call. A deferred `len` or
`cap` fails preprocessing, as Go refuses it: folded, it left the `defer` no
call and the transcriber panicked.

The op benchmarks hand the allocator its own meter and report it as
`alloc-gas/op`, which `gen_analysis.py` subtracts from the measured time.
`Copy` and the zero-fills charge CPU gas through that meter, so a refit read
`OpCPUSlopeValueDecl` as -51, which would make every `var` panic with `gas
must not be negative`. `benchMachine` now wraps the meter to drop
`CPUCycles`.

`OpCPUSlopeArrayLit` and `OpCPUSlopeStructLit` are refitted, since a literal
no longer zero-fills the elements it names. No reference Xeon was at hand, so
each family was timed on one machine, pinned to one core, both before
and with this change, and scaled to the Xeon by the factor of the unchanged
families whose Xeon fit is published. A published fit is the Xeon's time less
the allocation gas of its own run, about 1.6 per element in
`op_bench_do_dedicated.txt`, while allocation now costs 6.6 per element of
these literals, so the Xeon's time per element is the published fit plus 1.6.
Each factor divides that time by the local one: `SliceLit` 1.48, `SliceLit2`
1.41 and the positional `StructLit` 1.52. The largest, rounded up to 1.53,
scales each local time, and today's 6.6 is subtracted from the result, so the
slopes err towards overcharging:

| Family | Published | Before | This change |
| --- | --- | --- | --- |
| `ArrayLit` (int) | 52.5 | 91.3 | 32.4 |
| `ArrayLit` (keyed) | — | — | 33.7 |
| `ArrayLit` (uint8) | 8.8 | 6.1 | 6.0 |
| `StructLit` (positional) | 29.3 | 24.4 | 28.2 |
| `StructLit` (keyed) | 51.0 | 108.0 | 45.6 |

A first pass solved each factor against the published fit plus today's 6.6
instead of its own 1.6, which raised every factor by about 16% and gave 40
and 54. Two reruns on the same model of machine, a Ryzen 7 7840U, pinned over
eight and fourteen alternated rounds, read `SliceLit2` as the largest factor
each time, 1.37 and 1.51, and with it the int array literal at 32.0 and 32.9,
the keyed one at 34.2 and 34.5, the uint8 one at 6.4 and the keyed struct
literal at 42.7 and 44.6.

The keyed array literal, `[N]int{0: x, 1: y}`, has its own benchmark,
`BenchmarkOpArrayLit_Keyed_*`. `doOpArrayLit` converted each key to `int`
again per element, though preprocess already makes every key an `int`
`ConstExpr`, and that made a keyed element cost 49.2, 1.57 times the
positional 31.4 its session measured; it now reads the key with `GetInt`,
and the keyed row, 33.7 and 32.2 over two later sessions, sits next to the
int row, which those sessions read at 30.0 to 40.3, the high end slower than
keyed, which does strictly more, and so taken as noise. Those two sessions
only checked the keyed fix and set no constant: each constant keeps the
largest of its variants over the table and the two reruns, rounded up, so
`OpCPUSlopeArrayLit` goes from 52 to 35 and `OpCPUSlopeStructLit` from 51 to
46. The Before column times the int array and keyed struct literals at 1.7
and 2.1 times their published fit, so before this change they were
undercharged by 43% and 53%; the fill they did per element has since grown a
pkg-ID stamp, which the published fits predate.

`gen_analysis.py` also mapped the `TypeAssert2 (iface hit)` family to no
constant, so its fit of 355.3 per method never competed for
`OpCPUSlopeTypeAssertIface`, which shipped `TypeAssert1`'s 348.9 although
`doOpTypeAssert2` charges it too. Mapped like `TypeAssert1`, the generator
keeps the higher fit, and the constant goes from 349 to 355.

After the change, the `[1000]int` filetest costs 592,922 gas: 40,000 more
for each of its eleven copies, the declaration's included, and 40,000 for the
zero value the declaration builds.

## Alternatives considered

- A charge in `PointerValue.Assign2` only, walking the value being assigned.
  Named results, arguments, `range`, conversions and receivers copy without
  passing through `Assign2`, so each would stay free.
- Threading `*Machine` into `Copy` to call `m.incrCPU`. It changes the
  signature of every caller for a gas meter the allocator already holds.
- Leaving byte arrays at allocation gas alone, since their copy is one
  `memcpy`. The `append` builtin and `[]byte` to `string` conversion already
  charge 4 gas per byte for the same `memcpy`, so a byte array copy would be
  the one place it is free.
- Charging the zero-fill `OpCPUSlopeValueDecl`, 43, whose fit reads 42.9 per
  element of a `[N]int` declaration. That benchmark times `doOpValueDecl`
  whole, the zero value and the copy `Assign2` makes of it, so it is not the
  fill's own price.
- Charging a byte array's zero `Data` 4 per byte. Nothing loops over it, and
  `make([]byte, n)` would then need the same charge to keep one price for the
  same allocation.
- Storing `var a T`'s zero value without the copy `Assign2` makes of it. The
  copy stamps a struct with its declared type's realm and the zero value does
  not, so it would change which realm may write the variable.
- Moving every sliced array variable to the heap, as `&a` moves it. Only a
  for-init variable needs a per-iteration copy, and `buf[:]` on a local array
  is common enough that a heap item for each would cost every such call.

## Consequences

- Every by-value copy of an array or struct costs more gas: 40 per element or
  field, and 4 per byte of a byte array. A `[32]byte` copy pays 128.
- Ranging over an array by value and reading its elements pays 40 per element
  instead of 15, and 4 per byte of a byte array instead of 15; a key-only
  range copies nothing. Ranging over a slice or through an
  array pointer no longer pays 15 per element.
- A repeated copy of a value nested deeply by value, `x = [1]any{x}` in a loop,
  was already quadratic in allocation gas and now pays the quadratic walk too:
  `gas/nested_alloc.gno` moves from 8,559,693,635 to 12,560,317,779, of which
  794,144 comes from collections now visiting the anchored lists and -170,000
  from the refitted array literal slope.
- A zero value costs 40 per element or field it writes: 100 calls of
  `new([1000000]int)` go from 591,438,269 to 4,591,438,269 gas, and
  `make([]int, n)` pays the same. `var a [N]int` pays the fill and the copy
  `doOpValueDecl` makes of it, 80 per element. An interface field is left
  undefined and costs nothing, as an interface element does.
- `gas/compute_map_key_big_struct.gno` holds both 40 MB copies in the count
  and needs a 100 MB cap where it ran under 50 MB.
- `len` and `cap` of an array, or of a pointer to one, without a call cost
  nothing at run time and no longer evaluate the operand.
- An array literal pays 35 per element it names, down from 52, and a struct
  literal 46 per field, down from 51; a missing element or field pays the
  zero-fill.
- An interface type assertion or type switch pays 355 per method checked, up
  from 349, and so does printing, which checks each value against
  `fmt.Stringer` and `error`: 12 more gas per value printed.
- `m[k] = v` on a new key of a map of arrays or structs pays the copy of `v`
  and no zero value, and a range over a map pays nothing for a key or value
  bound to `_`.
- A for-init array or struct whose address is taken, whether by `&`, a
  closure, a pointer method value or a slice of the array it holds, is copied
  at the end of each iteration, 40 per element or field, whether or not
  anything keeps it past the iteration.
- A collection during `make([]T, n)`, `append` into a new array, `new(T)` or a
  struct's zero value counts the value being built, so a cap that refused the
  same memory held by a variable now refuses it here too.
- The `append` and `copy` builtins keep their per-element charge for the
  top-level elements they move; `Copy` charges only the elements of each array
  or struct element it duplicates, so no element is charged twice.
