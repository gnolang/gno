# ADR: refuse a type whose id passes 16 KiB

## Context

A declared type's `TypeID` is its name. Every other type's id writes out
everything it refers to, and an alias refers without adding a name, so eight
aliases of eight fields each write eight to the eighth copies of the bottom type.
`Type.String` does the same. Both build Go strings outside the gas meter and the
allocator, and memoising the id does not shrink it. A free query or a cheap
transaction could make a node allocate until it stopped.

An expression's text multiplies the same text again: `func(A, A, ...)` in an error
message repeats the alias's text once per parameter, four source bytes each.

`go/types` walks the same expansion before any id is built. It compares two unnamed
types and validates a declared type's base field by field, with no memo and outside
the gas meter. This record leaves that walk as it is, under Consequences.

## Decision

- `typeIDWriter` builds every composite id and panics before it writes past
  `maxTypeIDBytes`, 16 KiB. A child's memoised id is already its size, so no size
  walk or stored size is needed, and a list of children near the cap costs one cap.
- Every id under the cap keeps its bytes, so gas and state bytes do not move.
- `writeType`, the one type printer, writes a type's text through a cut section of
  the `meteredWriter` it renders into, at `maxTypeTextBytes`, twice the id's cap,
  since text can run before the id is built. The section drops every byte past the
  cut and ends the walk there, so a printed value's type streams into its render
  with no buffer of its own, and `Type.String` is the same section over a
  `strings.Builder`.
- A type's text and a constant's text inside an expression's text are cut at
  `maxExprTypeTextBytes`, 1 KiB. `TypeValue.String` takes that cut on every route,
  so a query answering with a type value gets at most 1 KiB of its text, marked,
  where it printed whole before.
- A run-time error a realm can recover, a failed assertion, a comparison or a map
  key, prints each type whole, as `Type.String` does. A realm can store the message
  or branch on it, so a cut there would move state and results under the cap. Text
  under the id's cap never reaches the text's cut, so the message keeps the bytes it
  had. A type stored past the id's cap prints its first 32 KiB where its text is
  longer.
- One recoverable error is the exception: a write to another realm's object panics
  with its target expression's text, so each type in that text takes the 1 KiB
  expression cut. Printing them whole would let one wide target, such as
  `f(A(nil), A(nil), ...).F`, copy up to 20 KiB per four source bytes into a
  message a recover loop rebuilds outside gas, and bounding the whole message
  takes a node printer on a writer.
- A type read back from the store keeps its id past the cap, by a flag `fillType`
  sets. Only a binary without the cap could have written one, and refusing it would
  fail every call touching the object holding it, a map holding it as a key on every
  load. Stored ids are unchanged, so such a value behaves as it did.
- The func type `Specify` builds for a uverse generic skips the cap, and so does
  its bound type. It repeats its operands' types, each already under the cap, at
  most three times for `append`, so its id stays under three caps, and capping it
  refused `append` over an unnamed element of 37 percent of the cap.
- `vm/qfuncs` streams its type text into the answer, which is refused past the
  1 MiB cap on every vm query's answer. The longest such text in the scan below is
  about 41 KB.

## How the values were measured

Instrumented copy of the tree, preprocessing and sealing `gnovm/stdlibs`, `examples`
and the packages and runs of `gnolang/tx-exports` `mainnet.gno.land` to height
1308735: longest id 7550 bytes, a struct of unexported fields in a realm with a
60-byte path; longest text 5286 bytes; text at most 1.137 times its id. 16 KiB holds
that struct at the longest deployed path, 87 bytes, an extrapolation. A type holding
unfilled `RefType` parts prints 9 bytes per part past its id, which only a type
built by hand before `fillType` shows.

The expression cap sits past the 1024 bytes a transaction's error keeps, so no
transaction's result bytes change. The longest such text in a golden is 44 bytes.

## Alternatives

- A cap at the alias declaration: an unnamed type repeating an alias under the cap
  never reaches it, and it refused small valid packages in one declaration order.
- A sha256 digest past 1 KiB: built as a prototype and dropped. It bounds the id
  but not the type the id tracked: saving a value copies and amino-encodes its
  type with every alias written out before `AminoEncodePerByte` charges for it,
  so deploying a seven-level alias chain of eight fields under a 10M gas limit
  allocated 743 MB, eight times more per level. The id cap bounds that encoding
  too, since an id grows with the type it writes out.
- A second cap on what a parent copies from a child: see Consequences.
- A separate buffer for each type's text, cut and then copied into the render:
  built first and replaced. Every printed value then built its type's text apart,
  so `println` of a 10000-element `[]int` made 30002 allocations and took about
  three times the base's CPU for the same output gas, where the cut section makes 2,
  as the base does.
- A node count per type before `go/types`: built and removed. A fresh review measured
  it deploying or refusing one package by Go map order, 8 of 80 runs deployed where
  the base deploys 80, which splits consensus. It added nested generic instances,
  `H[H[...H[int]...]]` at depth 10 and 153 bytes costing 14.3 s, and left out union
  terms, 100 costing 17.9 s. A declared interface chain, 20.3 s, and duplicate type
  switch cases, 18.4 s and quadratic, multiply the walk under any cap per type, and
  it refused packages the base accepts by over-counting repeated embeds.
- Repeating the upgrade-height scan over stored values instead of keeping their ids:
  a scan reads a past height, and a 311-byte `MsgRun` planted after it fails every
  later touch of a map holding the value.

## Consequences

- A realm recovering a write to another realm's object gets a message that differs
  from the base's wherever a type in the target expression prints past 1 KiB. The
  message is a pure function of the target's source, so a scan of every deployed
  package and run for an assignment target holding such a type settles whether
  `gnoland-1`'s history holds one. That scan has not been run.
- A package holding a type past the cap fails to deploy, and a node whose state
  holds one panics at boot in `PreprocessAllFilesAndSaveBlockNodes`, which has no
  recover, and the panic names no package. The longest id on `gnoland-1` at height
  1308735 is 7550 bytes.
- Each distinct type node still holds up to one cap, so allocation stays linear in
  source: about 6500 bytes per source byte on a query and 39000 on a package whose
  deposit fails, against about 2000 and 3000 for the same statements over a small
  type. A child cap of 1 KiB, over a measured maximum of 283 bytes, brought those to
  2235 and 4761 in a prototype, and was left as an open call. At full size, a
  480 KB query over such a type allocated 1923 MB with 519 MB at peak, and a 120 KB
  package of `var V *A` lines 5130 MB with 2046 MB at peak, about 45000 bytes per
  source byte, where gas is what stops it. The base measures the same on both.
- An error message about a wide expression sums its 1 KiB cuts before it is cut to
  1024 bytes, at about 6800 bytes per source byte.
- A realm recovering a failed check in a loop builds each message whole, outside
  the gas meter, as on the base: the time grows with the types' text and the gas
  does not.
- `go/types` still walks an unnamed type whole, as on the base: `var b *B9 = a` between
  two chains declared apart, 1372 bytes and 1860784 gas, holds every validator for
  seconds, and each level multiplies that by about eight. Bounding it takes a separate
  change, with its bound inside the checker's walk.
- `maxTypeIDBytes` holds a struct's unexported fields with its package path written
  in each, so a 255-byte path refuses a struct of 63 `int` fields the base accepted.
  The longest deployed path is 87 bytes.
