# Limit native Go slices to visible length

## Context

`Gno2GoValue` allocated Go slices using a Gno slice's `Maxcap`. For byte-backed
slices it also copied through `Maxcap`. A zero-length slice with a large
capacity therefore caused an unbounded allocation and copy before a native
call, even though the native function could only observe the slice's visible
elements.

## Decision

Native Go slices expose only the Gno slice's visible `Length`. Both byte-backed
and list-backed conversions allocate with `len == cap == Length`, and
byte-backed conversions copy only `Length` bytes. Capacity beyond `Length` is
not allocated or copied.

## Alternatives considered

- Metering conversion by `Maxcap` prevents unpaid work but changes gas costs for
  ordinary native calls and requires coordinated activation.
- Metering only `Maxcap - Length` avoids charging work already represented by
  native pricing, but still performs unnecessary allocation and copying.
- A transient-memory budget models temporary Go allocations directly, but adds
  accounting complexity while retaining work that natives do not need.

## Consequences

- The Go slice handed to a native function always has `cap == len`. Natives must
  not rely on a parameter slice's capacity: no reslicing past `len`, no in-place
  `append` into the spare capacity, no aliasing assumptions. No currently
  registered native does.
- Hidden Gno capacity is no longer allocated or copied during conversion.
- Native gas pricing is unchanged, so the fix needs no coordinated gas-schedule
  upgrade.
