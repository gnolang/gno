# ADR: a library method a caller picks acts for a realm only if the realm imports it

## Status

Proposed. Stacked on `prxxxx_explicit_entry.md`.

## Context

The explicit-entry rule says a realm's state is written only by code named in
that realm's source. A function value is covered: called through a value it
runs view-only. A `/p/` method reached through an interface is not:

```go
type Runner interface{ Run(o *Obj) }
func Use(r Runner) { r.Run(gObj) } // a caller's r writes gObj
```

A first fix spared receivers the realm owns. It failed: a struct value is
copied under the realm, a pointer stored in one transaction is the realm's in
the next, and `f := r.Run; f(gObj)` is no interface call. Ownership tells where
an object lives, not who chose the code.

## Decision

A `/p/` method reached through an interface call or a method value runs
view-only unless the current realm imports its package, directly or through
another import. Exempt: the realm's own package, the calling package, built-in
methods (`cur.Previous()`) and stdlib methods.

- The import tree is built from stored package nodes, which never change, and
  cached on the realm's package node (not serialized, rebuilt after restart).
  It reads only the node cache, which holds every package from load or node
  start on, so it charges no gas and does not depend on cache state.
- A direct call on a concrete type is unchanged: the realm's source names it.

## Alternatives rejected

- **Receiver ownership** (the first fix): bypassed three ways, see Context.
- **Refuse every interface call:** breaks stdlib readers (`io`, `regexp`) and
  ~40 example packages.
- **Direct imports only:** breaks `commondao` (`bptree/list` calls `seqid`) and
  `ghverify` (`gnorkle/ingesters/single` calls `storage/simple`); the realm
  reaches those libraries through ones it imports.
- **Stdlib checked too:** `sort` calls a `regexp/syntax` `Swap`, `ufmt` writes a
  fresh `strings.Builder`; no realm imports those. A caller cannot author
  stdlib code.

## Consequences

- A caller's library type that the realm's import tree does not contain cannot
  write the realm's state through the realm's interfaces or method values.
- A refused call aborts the transaction: a panic in a view-only frame cannot
  be recovered by the realm.

## Known open

- A library the realm's import tree contains is trusted, so a caller can pick
  any of its types (an `avl.Tree` passed as the realm's `Setter`).
- A stdlib method that writes its argument (`bytes.Reader.Read(p)`) can be
  picked through an interface such as `io.Reader`.

## Tests

`gnovm/tests/files/zrealm_entry_iface_*.gno` pin each refused shape (primitive,
pointer, struct value, stored, method value, assertion, switch, map, embedding,
defer) and the allowed ones (an imported library, the realm's own wrapper).
`gno.land/pkg/integration/testdata/iface_storedptr_2tx.txtar` pins the
two-transaction stored pointer.
