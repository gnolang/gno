# groups

A `Group` is a set of addresses (the **base set**) plus any number of named
**Roles**, each with its own member set and optional metadata. One `Group`
per DAO, per board, per permissions instance — whatever your realm manages.

```
Group
├── base set:        the plain members (guests, users, council — you decide)
└── roles
    ├── "admin":     member set + meta
    └── "moderator": member set + meta
```

## Quick start

```go
import "gno.land/p/nt/groups/v0"

var group = groups.NewGroup()

func init() {
    // Base members.
    group.Add(address("g1alice..."))
    group.Add(address("g1bob..."))

    // A role with its own members.
    admins, _ := group.AddRole("admin")
    admins.Members().Add(address("g1carol..."))
}
```

## Three kinds of operations

Every method belongs to exactly one family, so a call site always says
which semantic it means — checking the base set and checking "anywhere in
the group" are different questions with different methods.

| Family | Methods | Looks at |
|---|---|---|
| Base set | `Add`, `Remove`, `Has`, `Size`, `Iterate` | base set only |
| Role registry | `AddRole`, `GetRole`, `HasRole`, `RemoveRole`, `RoleCount`, `IterateRoles` | the named roles |
| Aggregated | `HasAny`, `TotalSize`, `IterateAll`, `RolesContaining`, `RemoveFromAll` | base + every role, deduplicated |

So with alice in the base set only and dave in the "council" role only:

```go
group.Has(alice)  // true  — alice is a base member
group.Has(dave)   // false — Has never consults roles
group.HasAny(dave) // true — dave is somewhere in the group
group.RolesContaining(dave) // ["council"]
```

An address may appear in the base set and several roles at once;
`TotalSize` and `IterateAll` count and yield it once. All iterators take
`offset, count` for pagination, and the callback returns `true` to stop.

`RemoveRole` discards only the role itself — its members stay wherever
else they appear. `RemoveFromAll` is the opposite: it purges one address
from the base set and every role.

## Sharing across realms: readonly views

A `*Group` or `*Role` is a **mutable handle**: anyone holding it can change
your data (method calls run with the allocating realm's storage authority).
`Readonly()` returns a view that structurally cannot mutate — no mutator
methods exist on it at all.

Three rules at realm boundaries:

1. **Never accept** a `*Group`/`*Role` from an untrusted caller.
2. **Never return** a `*Group`/`*Role` to one — return
   `group.Readonly()` (a `*ReadonlyGroup`) or `role.Readonly()` instead.
3. **Never trust** a readonly view someone else hands you: it is a live
   window onto *their* data, which they can change between your reads.

## The `meta` slot

`Role.SetMeta(meta any)` stores arbitrary per-role data — permission bits,
a description, a quorum. Store **value types only** (strings, ints, value
structs/slices). Do not store pointers to types with mutator methods (such
as `*avl.Tree` or `*addrset.Set`): `Meta()` returns the value as-is, so a
reader holding a readonly view could call those mutators on it.

See `doc.gno` for the precise security model, and
`filetests/z_readme_filetest.gno` for this README as a running example.
