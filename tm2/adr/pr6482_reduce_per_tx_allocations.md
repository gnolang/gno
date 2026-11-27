# Reduce per-transaction allocations in object encode and decode

## Context

GnoVM stores every realm object as an amino `Any`: `defaultStore.SetObject`
encodes it with `amino.MustMarshalAny`, and `loadObjectSafe` decodes it with
`amino.MustUnmarshal`, which on master goes through the reflect decoder. A
write-heavy transaction (read-modify-write on an avl tree of 10,000 keys,
pebble backend) allocates about 1.85 MiB and 29,300 heap objects. An
allocation profile of the whole transaction (`MemProfileRate=1`, 500
transactions) put the top sites at, per transaction:

| site | KiB | objects | handled here |
|---|--:|--:|---|
| `amino.DecodeByteSlice` | 371 | 5,651 | yes, the intermediate copies (see 3) |
| `reflect.New` (decoder construction, `MethodByName`) | 182 | 3,486 | no |
| `amino.newTypeInfoUnregisteredWLocked` (param reads) | 95 | 385 | no, separate change |
| `strings.genSplit` | 88 | 2,795 | yes (see 4) |
| `reflect.growslice` (slice decode) | 86 | 520 | no |
| `gnolang.newBlockWithValueCap` (VM blocks) | 67 | n/a | no |
| `hex.EncodeToString` | 66 | 1,397 | the `ObjectID` part (see 1) |
| `amino.DecodeString` | 65 flat | 2,312 | yes (see 5) |
| `cache.(*cacheStore).setCacheValue` | 62 | 794 | no |
| `ObjectID.MarshalAmino` (`fmt.Sprintf`) | n/a | 666 flat | yes (see 1) |

## Decision

Five local changes, each with output identical to before by construction:

1. `ObjectID.MarshalAmino` builds `hex:newtime` in a stack buffer and
   allocates once, instead of `hex.EncodeToString` plus `fmt.Sprintf`. The
   generated encoder calls it twice per ObjectID (size pass and write pass).
2. `Codec.MarshalAny` for a `T` or `*T` with native genproto2 methods sizes
   the whole envelope with `SizeAnyBinary2` and writes it backward into one
   buffer with `MarshalAnyBinary2`, instead of a value buffer, a pooled
   envelope buffer and a final copy. Both helpers already exist and are what
   nested `Any` fields use. `**T` and non-genproto2 types keep the old path.
3. The reflect decoder reads nested structs, `Any` values, implicit list
   structs and skipped unknown fields as sub-slices of the input
   (`decodeByteSliceNoCopy`) instead of copying them first. Those bytes are
   decoded further and then dropped; every decoder that keeps bytes in the
   result (`DecodeByteSlice`, `DecodeString`, byte arrays) still copies, so a
   decoded value never aliases its input.
4. `typeURLtoFullname` uses `strings.LastIndexByte` and
   `ObjectID.UnmarshalAmino` uses `strings.Cut`, instead of `strings.Split`.
   Same accepted inputs and same errors.
5. `DecodeString` converts straight from the input, one allocation instead of
   a `[]byte` copy plus a string copy.

## Alternatives considered

- Teaching the genproto2 generator to size an `ObjectID` without building the
  string. Removes the remaining second `MarshalAmino` call per ObjectID, but
  changes generated code across the repo for a few percent more.
- Caching the `UnmarshalAmino` method index in `TypeInfo` instead of
  `MethodByName` per decode (about 1,500 objects per transaction). Left out to
  keep the diff contained; the decoder fast path in flight elsewhere makes the
  reflect path for stored objects less hot anyway.
- Folding in the param type-info memo: it is its own change.

## Consequences

Measured on master `d78cddd90`, paired runs on the same cores, 2000
transactions each: allocation per transaction falls from 1,894 KiB / 29,301
objects to 1,361 KiB / 19,710 objects (-28% bytes, -33% objects), GC cycles
per 2000 transactions from 22 to 15-16, GC CPU from 7.6-7.8 s to 5.4-5.8 s,
tx/s +13 to +14%. Gas and the app hash are identical in every run. With
`GOGC=off` the throughput gain is +8 to +16%, so most of it is the
allocation and copy work itself rather than collector pressure.

`TestDecodeDoesNotAliasInput` pins the no-aliasing invariant of (3): making
the leaf `DecodeByteSlice` alias its input turns it red.
