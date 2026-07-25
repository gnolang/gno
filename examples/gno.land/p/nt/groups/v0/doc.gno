// Package groups provides Groups containing a base address set plus named
// Roles, each with their own member set and metadata.
//
// A Group is the top-level container — one per DAO, one per permissions
// instance, etc. A Role is a named subset within a Group with arbitrary
// per-role metadata.
//
// The API separates three concerns explicitly, so each call site picks the
// right semantic:
//
//   - base-only operations: Add, Remove, Has, Size, Iterate;
//   - role registry operations: AddRole, GetRole, HasRole, RemoveRole,
//     RoleCount, IterateRoles;
//   - aggregated operations across base + all roles: HasAny, TotalSize,
//     IterateAll, RolesContaining, RemoveFromAll.
//
// # Security model
//
// A Group, and the *Role values it hands out, are meant to be allocated and
// held by the consuming realm. Two rules apply at realm boundaries:
//
//  1. Do not ACCEPT a *Group or *Role from an external/untrusted caller —
//     subsequent mutations would route to the allocating (attacker)
//     realm's authority, and a poisoned Group could cause DoS or
//     unexpected state.
//
//  2. Do not RETURN a *Group or *Role from any method or function callable
//     by untrusted realms. Return *ReadonlyGroup or *ReadonlyRole instead.
//     Exposing a mutable handle is exactly as dangerous as accepting one.
//
//  3. Do not TRUST a *ReadonlyGroup or *ReadonlyRole received from an
//     untrusted caller. A readonly view is a live handle over its creator's
//     data, not a snapshot: the sender controls the contents and can mutate
//     them between reads. Base authorization and accounting decisions only
//     on views derived from a Group you allocated yourself.
//
// The Readonly() views are the only safe handles to cross a realm boundary —
// safe to hand out, per rule 3 not blindly safe to consume.
//
// # Metadata: do not store mutable pointers
//
// Each Role has a free-form "meta any" slot. Meta() returns the stored value
// as-is, so a pointer stored in meta can be retrieved by an untrusted reader
// holding a Readonly() view. A direct
// field write through that pointer is still blocked by the realm-ownership
// gate, but invoking a MUTATOR METHOD on it (or passing it into a function
// that mutates by argument) runs under whatever realm allocated it (borrow
// rule #2) and commits the write. This includes common /p/ types such as
// *addrset.Set and *avl.Tree — they are mutable pointers, not "just data".
// Therefore store only:
//
//   - value types (ints, strings, value structs/slices with NO internal
//     pointer reaching a mutator-bearing type), or
//   - a wrapper whose only exported methods are read-only and which holds no
//     externally-mutable pointer.
//
// # Readonly views
//
// Group and Role each expose a Readonly() method returning a typed
// read-only view (ReadonlyGroup, ReadonlyRole; role member sets surface as
// *addrset.ReadonlySet). The views are concrete structs with unexported
// fields and only read-side exported methods, so cross-package callers
// cannot mutate through them.
package groups
