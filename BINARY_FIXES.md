# Binary Codec Fix Plan

Gaps identified by auditing `tm2/pkg/amino/genproto2/` against
`BINARY_CONDITIONS.md`. Each entry lists the bug, the reflect-side source
of truth, the generator site, the concrete fix, and the test to add.

Severity tiers:

- **CRITICAL** — wire-integrity / consensus divergence. Peers can disagree.
- **LATENT** — code path not reachable by any currently-registered type, but
  the condition is wrong; future types would silently miscompile.
- **BRITTLE** — observationally equivalent today but structurally divergent;
  relies on an unintended composition of two other rules.

---

## 1. CRITICAL — monotonic fnum check too permissive

**Bug.** `gen_unmarshal.go:255` emits `if fnum < lastFieldNum { error }`.
Reflect-side (`binary_decode.go:1009`) uses `fnum <= lastFieldNum`. The
generator lets the same field number appear twice in sequence; a peer that
crafts such a wire image would be accepted by the generator fast path but
rejected by the reflect fallback.

**Source of truth.** `binary_decode.go:1009`:
```go
if fnum <= lastFieldNum { return ... }
```

**Fix.** One-character change: `<` → `<=` at `gen_unmarshal.go:255`.
Regenerate all `pb3_gen.go`.

**Test.** Negative test in `tm2/pkg/amino/tests/`: craft bytes with field
1 appearing twice, assert both `UnmarshalReflect` and `UnmarshalBinary2`
reject. Add a positive fixture that confirms monotonic (strictly
increasing) field numbers still decode.

---

## 2. CRITICAL — top-level non-struct repr decoder ignores BinFixed64/32

**Bug.** `gen_unmarshal.go:161-170` is the primitive decode branch inside
`writeReprUnmarshal` (used when the top-level registered type is a non-
struct AminoMarshaler whose repr is a primitive). It unconditionally emits
`amino.DecodeVarint` / `amino.DecodeUvarint` for every int/uint kind,
without consulting `fopts.BinFixed64` / `fopts.BinFixed32`.

This is the same bug shape as the bare-`int` bug in `typeToTyp3` (already
fixed). The sibling inside `writePrimitiveDecodeFrom` (`gen_unmarshal.go:
785-859`) handles BinFixed64/32 correctly — the repr wrapper does not.

Effect: a top-level `type X int64; func (X) MarshalAmino() (int64, error)`
declared as `amino:"fixed64"` encodes via the fast path as 8 fixed bytes,
but decodes via the generator as varint → wire mismatch.

**Source of truth.** `binary_decode.go:86` switch for `Int64`/`Int32`/`Int`/
`Uint64`/`Uint32`/`Uint` honoring BinFixed64/32.

**Fix.** In `writeReprUnmarshal`'s primitive branch, mirror the BinFixed
dispatch already present in `writePrimitiveDecodeFrom`. Both must decide
from `fopts.BinFixed64` / `fopts.BinFixed32`.

**Test.** Add parity fixture in `tm2/pkg/amino/tests/`: a top-level
AminoMarshaler whose repr is `int64` with `fixed64` tag. Round-trip via
`aminotest.AssertCodecParity`.

---

## 3. CRITICAL (latent, but easy to trip) — no trailing-bytes check on non-struct decode

**Bug.** `gen_unmarshal.go:52-57` decodes a non-struct top-level via
`writeReprUnmarshal`, then returns without asserting `len(bz) == 0`.
`UnmarshalReflect` enforces `n != len(bz) → error` at `amino.go:1054`.

For primitive repr (lines 161-180) the value is decoded byte-by-byte and
trailing input is silently truncated. Malformed or extended inputs decode
successfully under Binary2 and fail under Reflect.

**Fix.** After the final `writeReprUnmarshal` call at `gen_unmarshal.go:56`,
emit `if len(bz) != 0 { return fmt.Errorf(...) }`. Verify the
packed-list repr and unpacked-list repr paths already handle trailing
bytes correctly (they do — packed uses `DecodeByteSlice` length-prefix,
unpacked drains on `for len(bz) > 0`).

**Test.** Negative fixture: valid encoding + one trailing `0xff` byte.
Reflect decoder rejects; generator decoder must also reject.

---

## 4. LATENT — `MarshalAnyBinary2` single-`0x00` elision missing

**Bug.** `amino.go:663` (hand-written helper, not generator-emitted) only
checks `innerLen > 0` when deciding whether to emit the Any.Value field.
Reflect-side (`binary_encode.go:302`) additionally checks
`len(bz2) == 1 && bz2[0] == 0x00` and elides the field in that case too.

Today, no generator-emitted `MarshalBinary2` can return `[0x00]` as its
sole byte, because `writeReprMarshal:82-87` rolls back the single-`0x00`
byte already. So this gap is not reachable through the generator at
present. Still, any future hand-written `MarshalBinary2` (or a change to
the generator) that produces `[0x00]` would regress.

**Fix.** Change `innerLen > 0` guard at `amino.go:663` to also exclude the
single-`0x00`-byte case, mirroring `binary_encode.go:302`.

**Test.** Synthetic test type with `MarshalBinary2` that emits `[0x00]`;
assert cross-parity.

---

## 5. LATENT — Float32/Float64 emission not gated by `fopts.Unsafe`

**Bug.** `gen_marshal.go:727-730` emits `PrependFloat32` / `PrependFloat64`
unconditionally. Reflect (`binary_encode.go:117-118`) errors unless
`fopts.Unsafe` is set.

`ValidateBasic` panics at codec-init time if a non-`Unsafe` float field is
registered, so the gap is not reachable for registered types. But the
generator's contract is to match the reflect codec; a future change that
relaxes `ValidateBasic` would silently emit float bytes.

**Fix.** Add an emission-time check: if `!fopts.Unsafe`, panic during
generation (not at runtime).

**Test.** None needed directly — covered by `ValidateBasic` test.

---

## 6. LATENT — `writePrimitiveEncode` ignores `beOptionByte` for `Uint8`

**Bug.** `gen_marshal.go:703-704` emits `PrependUvarint` for any
`reflect.Uint8` scalar. Reflect (`binary_encode.go:114`) emits
`EncodeByte` (bare, one byte) when `options&beOptionByte != 0`.

Not currently reachable because `beOptionByte` is set only by
`writePackedSliceReprMarshal` / `writeUnpackedListMarshal` (lines 108-113,
456-467), and those helpers inline the bare-byte case before calling
`writePrimitiveEncode`. A scalar `uint8` field whose `fopts` somehow
carried `beOptionByte` would miscompile.

**Fix.** Guard in `writePrimitiveEncode`: when `Uint8` and the caller
indicates `beOptionByte`, emit a single-byte write instead of
`PrependUvarint`. Requires plumbing a flag through; alternatively assert
that `beOptionByte` never reaches `writePrimitiveEncode`.

**Test.** N/A (unreachable). Leave a `panic` assertion at the call site.

---

## 7. BRITTLE — unpacked-list non-pointer zero-element sentinel

**Bug.** Reflect-side (`binary_encode.go:460`) uses
`isNonstructDefaultValue(erv)` to decide whether to write the `0x00`
sentinel for a list element. Generator (`gen_marshal.go:487-498`) only
checks `elem == nil`, which is a pointer-only guard.

For non-pointer scalar/struct elements the generator still produces
`<fnum> 0x00` bytes by way of downstream rollback rules (primitive
`PrependVarint(0)` / struct `dataLen == 0`). The final wire bytes match
reflect, but the control flow is structurally different — a refactor that
changes downstream rollback behavior would regress.

**Fix.** Add an explicit `zeroCheck` guard at the unpacked-list element
site in `writeUnpackedListMarshal`, mirroring reflect's
`isNonstructDefaultValue(erv)` branch. The emission can then be a direct
`<fnum> 0x00` without depending on downstream composition.

**Test.** No regression possible today; keep the current parity fixtures
that cover `[]string` with empty-string elements and `[]int64` with
zero elements.

---

## Execution order

1. **#1** (`<=` → one-liner) — fix and regenerate. Add replay test.
2. **#2** (BinFixed64 in repr decoder) — mirror existing dispatch. Add
   parity fixture.
3. **#3** (trailing-bytes check) — three-line emit. Add negative test.
4. **#4-#6** (latent) — tighten emission-time guards. No regen needed for
   #4 (hand-written helper).
5. **#7** (brittle) — refactor `writeUnpackedListMarshal` to use
   `zeroCheck` directly. Confirm existing fixtures still pass.

After each fix: run the full parity-test suite
(`go test ./tm2/pkg/amino/... ./tm2/pkg/bft/... ./tm2/pkg/crypto/...
./tm2/pkg/std/... ./gnovm/pkg/gnolang/... ./gno.land/...`). Recompile
`pb3_gen.go` where the generator itself changed (#1, #2, #3, #5, #6).

---

# Second-pass findings

A second audit pass, with three agents focused on (A) cross-cutting helpers,
(B) list/byte/array edges, (C) struct-field emission and the
`writeFieldIfNotEmpty` rollback contract, surfaced the following additional
gaps. Items below are ordered by severity, not discovery order.

## 8. CRITICAL — `ertIsStruct` keyed off `ReprType.Type.Kind()` in generator; reflect keys off `einfo.Type.Kind()`

**Bug.** Three sites:
- `gen_marshal.go:487` — `ertIsStruct := einfo.ReprType.Type.Kind() == reflect.Struct`
- `gen_size.go:396` — same
- `gen_unmarshal.go:589` — `isStructLike` uses `einfo.ReprType.Type`

Reflect (`binary_encode.go:399`) keys off `einfo.Type.Kind()`. For a
`[]*X` where `X` is a Go struct that implements AminoMarshaler with a
non-struct repr (e.g. `func (X) MarshalAmino() (string, error)`), the
generator sees `Repr=String` and emits the `0x00` sentinel without
requiring `nil_elements`. Reflect sees `Type=Struct` and errors unless
`nil_elements` is set. Two peers using different codec paths will
disagree on whether to accept a nil element.

**Fix.** Change all three sites to `einfo.Type.Kind() == reflect.Struct`
(and keep the companion check `ert.Kind() == reflect.Ptr` for the
pointer predicate). Regenerate `pb3_gen.go`.

**Test.** Synthetic AminoMarshaler with struct Go type + string repr, used
as `[]*X` without `nil_elements`. Confirm both codecs reject a nil
element in the slice.

## 9. CRITICAL — `writePrimitiveDecodeFrom` byte-array branch does not enforce length

**Bug.** `gen_unmarshal.go:890-898` (`reflect.Array` / `Uint8` branch).
After `amino.DecodeByteSlice`, it emits `copy(accessor[:], v)` with no
check that `len(v) == N`. A 20-byte typed address decoding a 19-byte or
21-byte payload silently truncates/zero-pads.

The struct-field top-level byte-array decoder (`gen_unmarshal.go:377-391`)
does enforce the length correctly — but the shared
`writePrimitiveDecodeFrom` is also reached from list-element decoding,
AminoMarshaler repr decoding, and implicit-struct nested-list element
decoding. Those call sites silently accept malformed input.

**Reflect ground truth.** `binary_decode.go:551-555`:
`if len(byteslice) != length { err = "mismatched byte array length" }`.

**Fix.** Before the `copy`, emit
`if len(v) != N { return fmt.Errorf("mismatched byte array length: expected %d, got %d", N, len(v)) }`.

**Test.** Decode a `[]crypto.Address` where one element has 19 bytes on
the wire. Reflect rejects; generator currently accepts.

## 10. CRITICAL — unpacked-list array path silently accepts short input

**Bug.** `gen_unmarshal.go:280-293` (struct decode of a `[N]T` field via
unpacked list). Breaks out on any wire fnum change without asserting that
`N` elements were consumed. Also exits cleanly on a higher fnum
mid-array. For a fixed-size `[N]T` array field, a peer sending fewer
than `N` entries will leave the tail at Go zero and no error is raised.

**Reflect ground truth.** `binary_decode.go:625-644` — the packed-array
path iterates `for i := range length` and errors at `:637-640` if the
field number changes. The bare unpacked array path at `:988` requires
exactly `length` entries.

**Fix.** Track the consumed count; emit
`if consumed != arrayLen { return fmt.Errorf("expected %d array entries, got %d", arrayLen, consumed) }`
after the break. Change the inner-loop `break` to an error when the
count is still short.

**Test.** Fixed-length `[4]uint64` field with three wire entries.
Reflect rejects; generator currently accepts.

## 11. CRITICAL — `UnmarshalBinary2` does not reset absent-from-wire fields to default

**Bug.** `gen_unmarshal.go:230-310`. The generator drives decode from
the wire (`for len(bz) > 0 { switch fnum }`). Registered fields whose
`BinFieldNum` never appears on the wire are left UNTOUCHED on the
receiver. Non-pointer `time.Time` fields are force-initialized at line
239-248, and pointer fields are post-loop initialized at line 317-329 —
but every other kind is implicitly assumed to be the Go zero on entry.

**Reflect ground truth.** `binary_decode.go:971-974` and `:999-1002`:
for each registered field, if input is exhausted OR the next wire fnum
exceeds this field's BinFieldNum, the decoder actively calls
`frv.Set(defaultValue(frv.Type()))` — resetting to zero (or 1970 for
`time.Time`).

**Why this matters.** Any caller that reuses a receiver (a pool, a
repeated `UnmarshalBinary2` into the same target) will see stale values
in fields absent from `bz` under the generator, but zeroed under
reflect. Consensus paths that recycle allocations would be vulnerable.

**Fix.** Emit a per-field zero-assignment at the top of the generated
`UnmarshalBinary2` body (before the wire loop), for every non-pointer
field — mirroring the existing `time.Time` initialization. Pointers
stay on the post-loop path (already correct).

**Test.** Call `UnmarshalBinary2(cdc, bz1, 0)` followed by
`UnmarshalBinary2(cdc, bz2, 0)` on the same receiver, where `bz2` is a
subset of fields. Confirm the fields absent from `bz2` are zero, not
the values from `bz1`.

## 12. CRITICAL — `writeReprUnmarshal` primitive & packed-slice branches don't slide `bz`, invalidating the gap #3 trailing-bytes check

**Bug (extension of gap #3).** `gen_unmarshal.go:161-180` (primitive
repr) calls `amino.DecodeVarint(bz)` etc. with the consumed count bound
to `_`. `bz` is not re-sliced. Similarly, `:97` (packed-slice repr)
binds `_` from `amino.DecodeByteSlice(bz)`. The outer trailing-bytes
check proposed in fix #3 (added at `gen_unmarshal.go:56`) therefore sees
the uncut post-key remainder and always rejects valid input, OR always
accepts if the inner decode never touches `bz` — either way the fix as
worded is inert.

**Fix.** Before landing fix #3, in every branch of
`writeReprUnmarshal`, capture the consumed count (`n` instead of `_`)
and slide `bz = bz[n:]`. Then the outer `len(bz) != 0` check becomes
meaningful. Packed-slice branch also needs explicit
`if len(bz) != 0 { error }` after consuming the inner length-prefixed
payload, independent of the outer slide.

**Test.** Valid encoding + one trailing `0xff` byte. Reflect rejects;
generator must also reject, for each of: top-level primitive, top-level
packed-list, top-level unpacked-list reprs.

## 13. CRITICAL — Interface struct-field emission bypasses single-`0x00` outer rollback (coupled with #4)

**Bug.** `gen_marshal.go:660-669` `writeInterfaceFieldMarshal` guards
only on `accessor != nil`. When non-nil, it unconditionally emits
`<anyLen varint> <field key>`. It never consults whether the encoded Any
bytes are empty or a lone `0x00`.

Today, gap #4 keeps `MarshalAnyBinary2` from returning `[0x00]` (which
would already be a separate divergence from reflect). If fix #4 widens
`MarshalAnyBinary2` to return empty bytes in the single-`0x00` case, the
outer generator divergence becomes reachable: generator emits
`<key> 0x00` (length-prefix = 0), reflect emits nothing.

**Reflect ground truth.** `binary_encode.go:592` (`writeFieldIfNotEmpty`)
rolls back key+value for ANY field kind, including Interface.

**Fix.** Wrap the outer `anyLen + PrependFieldNumberAndTyp3` in the same
rollback shape used by `writeReprMarshal` (gen_marshal.go:78-87). Must
land together with fix #4.

**Test.** Interface field whose concrete `MarshalBinary2` returns
`[0x00]` (synthetic test type). Confirm both codecs elide the field.

## 14. LATENT — `UnmarshalAnyBinary2` (and sibling) miss `IsASCIIText(typeURL)` check

**Bug.** `amino.go:738` decodes `typeURL` via `DecodeString` and passes
it directly to the registry lookup. Reflect (`binary_decode.go:420`)
rejects non-ASCII typeURLs first. The sibling reflect-fallback entry at
`amino.go:1157` has the same omission.

**Fix.** Add `if !IsASCIIText(typeURL) { return fmt.Errorf(...) }` at
both sites immediately after `DecodeString` returns.

**Test.** Craft an Any-wrapped payload with a non-ASCII typeURL.
Both codecs must error with parallel messages.

## 15. LATENT — `UnmarshalAnyBinary2` assignability check runs once, reflect runs three times

**Bug.** `amino.go:791-793` checks `irvSet.Type().AssignableTo(rv.Type())`
once before decode. `decodeReflectBinaryAny` checks at three locations
(`binary_decode.go:441`, `:495`, `:514`).

Today the checks are equivalent because the registry is fixed at
startup. Future pluggable registry would diverge.

**Fix.** Add a second check after `pbm2.UnmarshalBinary2(...)` returns,
just before `rv.Set(irvSet)`. Zero runtime cost.

## 16. LATENT — `writeFieldValueMarshal` default primitive branch lacks inline rollback

**Bug.** `gen_marshal.go:417-422` (default branch) emits
`writePrimitiveEncode` + `PrependFieldNumberAndTyp3` unconditionally. The
caller's `writeEmpty` argument is accepted but ignored. Every existing
call site wraps this branch in an outer `zeroCheck`, so it's not
reachable today.

A future refactor that adds a new call site without the outer
`zeroCheck` would silently emit stray `<key> 0x00` bytes.

**Fix.** Make the default branch respect `writeEmpty` via an inline
rollback (shape of `writeReprMarshal`). Defensive hardening.

## 17. BRITTLE — unpacked-list pointer-to-zero non-struct element (sibling of #7)

**Bug.** `gen_marshal.go:486-498` gates sentinel emission on
`elem == nil`. For `[]*int{&0}` (pointer to zero), reflect takes the
`isNonstructDefaultValue` branch and emits `<fnum> 0x00` directly. The
generator falls through to `writePrimitiveEncode("(*elem)", ...)` which
produces `PrependVarint(0)` → the same single byte.

Wire bytes match today because `varint(0) == [0x00]`. Any change to the
length-prefix emission would regress.

**Fix.** Extend the `writeUnpackedListMarshal` guard to call
`zeroCheck` on the dereferenced value when
`ert.Kind() == reflect.Ptr && !ertIsStruct`, emitting the same
`<fnum> 0x00` shortcut reflect uses.

## 18. BRITTLE — `IsStructOrUnpacked` call sites pass zero FieldOptions

**Bug.** `gen_marshal.go:67`, `gen_size.go:60`, `gen_unmarshal.go:68`
all call `rinfo.IsStructOrUnpacked(amino.FieldOptions{})` with a zero
FieldOptions. Reflect's `UnmarshalReflect:1015` does the same.
`IsStructOrUnpacked` internally consults `rinfo.Elem.GetTyp3(fopt)` for
list kinds — today the zero fopt is safe, but future top-level list
types whose element needs non-empty fopts to disambiguate typ3 would
quietly miscompile.

**Fix.** Lift a shared helper `isTopLevelStructOrUnpacked(rinfo)` used
by both reflect and generator, or document the assumption at both call
sites.

## 19. BRITTLE — unpacked-list `writeImplicit` predicate defensive `einfo.Elem != nil` guard

**Bug.** `gen_marshal.go:477`, `gen_size.go:390`, `gen_unmarshal.go:569`
guard `writeImplicit` with `einfo.Elem != nil`. Reflect assumes
`einfo.Elem` is non-nil by TypeInfo invariant.

**Fix.** Drop the guard OR convert to a panic assertion, so generator
and reflect agree on the contract.

## 20. BRITTLE — AminoMarshaler-returning-byte-slice repr can bypass outer rollback

**Bug.** `gen_marshal.go:244-258` (AminoMarshaler non-struct branch). The
generator applies `reprZeroCheck` on the Go-side repr value. A repr that
is non-zero in Go-land but encodes to a single `0x00` on the wire (e.g.
`[]byte{0x00}`) would be emitted; reflect's outer
`writeFieldIfNotEmpty` would still roll it back.

No registered AminoMarshaler today returns such a repr. Either add a
post-emission rollback, or document the invariant.

---

# Revised execution order

Critical wire-integrity bugs (must land before any release):

1. **#1** — `<` → `<=` fnum monotonic check.
2. **#2** — BinFixed64/32 in `writeReprUnmarshal` primitive branch.
3. **#3 + #12** — trailing-bytes check AND slide `bz` in every
   `writeReprUnmarshal` branch. Land together.
4. **#8** — `ertIsStruct` keyed off `einfo.Type`.
5. **#9** — byte-array length check in `writePrimitiveDecodeFrom`.
6. **#10** — unpacked-list array short-input rejection.
7. **#11** — absent-field reset in `UnmarshalBinary2`.

Coupled-critical (land together):

8. **#4 + #13** — `MarshalAnyBinary2` single-`0x00` elision + outer
   Interface field rollback.

Latent / hardening:

9. **#5, #6, #14, #15, #16** — defensive guards; no wire impact today.

Structural cleanup:

10. **#7, #17** — list-element sentinel structural directness.
11. **#18, #19, #20** — helper/contract tightening.

After critical fixes: regen every `pb3_gen.go`; run full parity suite
and the txtar gas tests. Expect additional gas recalibration for the
absent-field reset change (receiver fields now carry explicit zero
assignment; marginal byte cost if any).

---

# Third-pass findings

Three more agents, with angles: (A) inspect GENERATED `pb3_gen.go` output
directly, (B) audit `BINARY_CONDITIONS.md` itself for completeness, (C)
audit WKT handlers + Sized-encoding + Any helpers (categories the prior
passes didn't cover in depth).

## 21. CRITICAL (couples with #4) — `SizeAnyBinary2` cannot elide single-`0x00` inner value

**Bug.** `amino.go:704` computes `s += fks(2) + UvarintSize(innerSize) + innerSize` when `innerSize > 0`. Purely arithmetic — no way to inspect buffer content. If fix #4 widens `MarshalAnyBinary2` (`amino.go:663`) to elide `innerLen == 1 && buf[offset] == 0x00`, `Marshal` emits fewer bytes than `Size` reports.

**Effect.** `Size`-driven pre-allocation (gas measurement, batch writers, any consumer that sizes-before-marshals) would over-allocate or mis-gate on size equality. Reflect has no such split path.

**Fix.** Either (a) document that no currently-registered concrete type's inner `MarshalBinary2` can produce `[0x00]` (because `writeReprMarshal:82-87` rolls that back) — in which case fix #4 is wire-defensive only and `SizeAnyBinary2` stays arithmetic; or (b) require `SizeAnyBinary2` to do a speculative marshal probe (expensive, changes the perf model). Recommendation: (a) + a comment at `amino.go:704` pointing at the coupling.

## 22. CRITICAL — sibling `marshalAnyBinary2` (buffer-based) also omits single-`0x00` elision

**Bug.** `amino.go:615-623` gates field-2 emission on `len(valueBz) > 0` only. Does not handle `len(valueBz) == 1 && valueBz[0] == 0x00`.

**Source of truth.** `binary_encode.go:302`: both conditions trigger elision.

**Reachability.** Called by `MarshalAny` at `amino.go:553` on the genproto2 fast path. Any pointer-to-registered-type routed through `MarshalAny` with an inner `[0x00]` marshal output would wire-differ from the reflect fallback.

**Fix.** Tighten guard to `len(valueBz) > 0 && !(len(valueBz) == 1 && valueBz[0] == 0x00)`. Apply at the same time as fix #4 (sibling site).

## 23. LATENT — `unmarshalAnyBinary2` (non-Depth variant) lacks `IsASCIIText` check

**Bug.** `amino.go:1157` decodes `typeURL` and passes directly to registry lookup without ASCII validation. Sibling of gap #14 at `amino.go:738`.

**Fix.** Add `if !IsASCIIText(typeURL) { return fmt.Errorf(...) }`. Land as part of #14.

## 24. BRITTLE — `unmarshalAnyBinary2` depth-limit check lives only in the Depth variant

**Bug.** `amino.go:1141` (non-Depth entry) has no `anyDepth > maxAnyDepth` guard. The limit fires only when nested recursion enters `unmarshalAnyBinary2Depth` at `amino.go:722`. Structure relies on every inner Any going through the Depth variant — a future refactor that changes the call chain silently disables the limit.

**Fix.** Mirror the check at the top of the non-Depth variant. Zero wire cost.

## 25. BRITTLE — `MarshalAnyBinary2` does not assert `innerLen >= 0`

**Bug.** `amino.go:660-666`. If a buggy emitted `MarshalBinary2` writes forward past `offset`, `innerLen := before - offset` is negative; `PrependUvarint(buf, offset, uint64(innerLen))` casts it and corrupts the wire silently. Reflect has no analogous failure (buffer-based).

**Fix.** `if innerLen < 0 { panic(...) }`. Zero-cost invariant guard.

## 26. CRITICAL (latent by current reachability) — struct-repr field emission gates on `dataLen > 0` but not on single-`0x00` payload

**Bug.** `gen_marshal.go:427-440` `writeLengthPrefixedField` (used for struct-repr fields, Time/Duration, nested structs). The rollback is `if dataLen > 0 { prepend key and length } else { offset = before }`. It does not match reflect's `writeFieldIfNotEmpty:592` which rolls back when the value is a single `0x00` byte too.

Today no nested struct can produce a single-byte `[0x00]` output because per-field rollback in its own `MarshalBinary2` precludes it. But if any future nested struct emission (or hand-written MarshalBinary2) produces `[0x00]`, the generator would emit `<key> 0x01 0x00`; reflect would emit nothing.

**Same-shape siblings.** This is a sibling of #13 (Interface field) and #20 (AminoMarshaler byte-slice repr) at a third emission position: struct-repr fields. The three together would make the emission-site rollback contract uniform across all field kinds.

**Fix.** Widen `writeLengthPrefixedField`'s gate to `dataLen > 1 || (dataLen == 1 && buf[offset] != 0x00)`. Land with #13 and #20.

## 27. LATENT — non-struct top-level ARRAY types delegate decode to reflect (`cdc.UnmarshalReflect`)

**Observation (not a bug).** Generator audit of 22 `pb3_gen.go` files shows every registered non-struct type whose Go Kind is `Array` (`PubKeyEd25519 [32]byte`, `PrivKeyEd25519 [64]byte`, `Hashlet [20]byte`, `IntAr`, etc.) has `UnmarshalBinary2` emit only `return cdc.UnmarshalReflect(bz, goo)`. The encode side is native; the decode side is not.

**Implication.** Gaps #1 (`<=`), #2 (BinFixed), #3 (trailing bytes), #11 (absent-field reset), #12 (bz slide) are NOT reachable for these types through the generator today — they reach reflect directly. But non-struct **non-array** types (`IntDef`, `StringValue`, `EventString`, `PrimitiveType`, `MemPackageType`, `MemPackageFilter`) DO go through `writeReprUnmarshal` and expose the bugs.

**Action.** No source fix. Document that non-struct-array top-level types trade off encode-speed for decode-correctness-via-reflect. Consider whether to emit native decode for these shapes in a follow-up.

## 28. DOC GAP — `BINARY_CONDITIONS.md` is materially incomplete

**Findings from completeness audit.** Nine categories of reflect-side conditionals are missing from `BINARY_CONDITIONS.md`:

1. **Well-known-type handlers** (`wellknown.go`, `encoder.go`, `decoder.go`). Time/Duration emit seconds+nanos with independent subfield zero-skip. `decodeSecondsAndNanos` state machine rejects duplicate fields, out-of-order fields, unknown fields, nanos outside `[-1e9, 1e9]`. None enumerated.
2. **Hand-written Any helpers** in `amino.go`: `MarshalAny`, `MarshalAnyBinary2`, `SizeAnyBinary2`, `UnmarshalAnyBinary2`, `unmarshalAnyBinary2`, `unmarshalAny2Depth`. ~12 conditionals per helper; doc cites zero.
3. **`UnmarshalSized` / `UnmarshalSizedReader`** framing: varint header overflow, truncated, trailing, `maxSize` guard. Not mentioned.
4. **`parseStructInfoWLocked` / `parseFieldOptions`** (codec.go:682-797). Determines `UnpackedList`, `BinFieldNum` (assigned by exported position, not struct position), tag-to-FieldOptions mapping. Legal tag combinations not enumerated.
5. **`getTypeInfoFromFullnameRLock` Timestamp/Duration redirection** at `codec.go:543, 548`. Affects Any-wrapped WKT lookup.
6. **`defaultValue` full branch table** (`reflect.go:116-151`): `*time.Time` → `&emptyTime{1970}`, `*Struct` → nil, `*OtherPtr` → `new(T)`, scalar → `reflect.Zero`. Doc cites only in passing.
7. **`isNonstructDefaultValue` specials**: `durationType` returns false (Duration is struct-like for zero-skip), pointer-element recursion (`*int` with value 0 is "default").
8. **`consumeAny` per-typ3 dispatch table** (`binary_decode.go:1058-1068`) — doc says "switch on typ3" without enumerating the four arms.
9. **`DecodeByteSlice` overflow guard** (`decoder.go:341`): `count > uint64(len(bz))` error. Only guard against unbounded allocation from a crafted varint length prefix.

**Action.** Revise `BINARY_CONDITIONS.md` to add four new sections: "Well-known types (time, duration, Empty)", "Any envelope (reflect + Binary2)", "Framing (Sized)", "TypeInfo construction". Until revised, any future audit using the doc as ground truth risks missing the above. Not a wire-format gap itself, but a meta-gap that enabled the earlier ones.

**Most likely to hide live generator divergences** (worth targeted parity fixtures):
- Time/Duration subfield zero-skip on size vs marshal: confirm `TimeSize`/`DurationSize` match `PrependTime`/`PrependDuration` for all four quadrants (seconds zero + nanos zero / non-zero × 4).
- `[]*int` with nil element: does generator match reflect's `defaultValue(*int)` = `new(int)` (allocated zero) vs `reflect.Zero(*int)` = nil? No current fixture confirms.
- `*int` field with value 0 at struct position: does generator's `zeroCheck` treat non-nil-pointer-to-zero as default (recurses through pointer) the way reflect does?

---

# Further revised execution order

All critical wire-integrity fixes (land before any release):

1. **#1** (`<` → `<=`).
2. **#2** (BinFixed in `writeReprUnmarshal`).
3. **#3 + #12** (trailing-bytes + bz-slide).
4. **#8** (`ertIsStruct` keyed off `einfo.Type`).
5. **#9** (byte-array length check in `writePrimitiveDecodeFrom`).
6. **#10** (unpacked-list array short-input rejection).
7. **#11** (absent-field reset in `UnmarshalBinary2`).

Coupled-critical — land together as "Any + single-0x00 elision":

8. **#4** (`MarshalAnyBinary2` single-`0x00` elision).
9. **#13** (Interface field outer rollback).
10. **#20** (AminoMarshaler byte-slice repr rollback).
11. **#21** (couple with #4: document Size/Marshal coupling).
12. **#22** (sibling `marshalAnyBinary2` same fix).
13. **#26** (struct-repr field single-`0x00` rollback).

Latent / hardening (can ship independently):

14. **#5, #6, #14+#23, #15, #16, #24, #25**.

Structural cleanup:

15. **#7, #17, #18, #19**.

Documentation:

16. **#28** — revise `BINARY_CONDITIONS.md` before the next audit cycle.

Deferred design:

17. **#27** — decide whether non-struct-array top-level types should emit native decode or keep delegating to `UnmarshalReflect`.

After all fixes: regen `pb3_gen.go`, run full parity suite, recalibrate
gas txtars. Consider adding parity fixtures suggested by the
completeness audit (`[]*int` with nil element; `*int` zero-value field;
Time/Duration subfield zero-skip matrix).

---

# Audit-loop findings

Items discovered during the systematic BINARY_AUDIT.md walkthrough
(per-condition cross-check with two independent agents).

## 30. CRITICAL (reachable only via `amino:"unsafe"` Float) — generator's `zeroCheck` for Float diverges from reflect

**Bug.** `gen_marshal.go:770-771` (`zeroCheck` function) returns `"%s != 0"` for `reflect.Float32` / `reflect.Float64`. Reflect-side `isNonstructDefaultValue` (`reflect.go:101`) returns `false` for Float (falls through to default), meaning Float fields are NEVER considered "default" — always emitted.

**Effect.** For a struct field of type `float32`/`float64` with `amino:"unsafe"` tag:
- Reflect: always emits the field. A zero-value Float writes 4 or 8 zero bytes (Typ34Byte / Typ38Byte).
- Generator: skips emission when value is `0`. Produces different wire bytes than reflect.

**Source of truth.** `reflect.go:101`:
```go
default:  // Float, Array, Complex, etc.
    return false  // NOT default → always emit
```

**Reachability.** Float fields require `amino:"unsafe"` to survive `ValidateBasic`. If a type has such a field (no currently-registered type does, per audit), encoder outputs diverge between codecs, round-trip with the peer fails.

**Fix.** Remove the Float cases from `gen_marshal.go:770-771` (return `""` like Struct, meaning "no zero-skip, always emit"). Array/Complex already fall through correctly.

**Test.** Synthetic test type with `Value float64 amino:"unsafe"` tag, round-trip with value 0.0 through both codecs. Before fix: generator emits `<no bytes>`; reflect emits `<key> 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00`.

---

## 29. BRITTLE — `unmarshalAnyBinary2Depth` empty `bz` leaves receiver untouched; reflect zeroes it

**Bug.** `amino.go:725-727` returns `nil` immediately when
`len(bz) == 0` without writing to the target. Reflect-side
`decodeReflectBinaryInterface` (`binary_decode.go:325`, condition
`decode.28`) explicitly sets `rv.Set(iinfo.ZeroValue)` on empty input.

**Effect.** If the caller reuses a receiver that previously held a
non-nil interface value, reflect zeroes it; generator leaves the stale
value. Observable divergence only on receiver reuse (pool/recycle
patterns). Similar in shape to finding #11 (struct-field reset) but
scoped to the Any envelope path.

**Source of truth.** `binary_decode.go:319-330`:
```go
if len(bz) == 0 {
    rv.Set(iinfo.ZeroValue)
    return
}
```

**Fix.** In `unmarshalAnyBinary2Depth` at `amino.go:725`, on
`len(bz) == 0`, dereference the pointer and set the target interface
to zero via `reflect.ValueOf(ptr).Elem().SetZero()` (or equivalent)
before returning. Zero-wire-cost invariant tightening.

**Test.** Unmarshal into a pre-populated `*iface` with empty bytes;
confirm reflect nils the interface AND generator now nils the
interface (currently does not).
