package home

import "strings"

// Render returns the r/docs index page.
// The path argument is ignored — this is a static landing page.
func Render(_ string) string {
	content := `# Gno by Example — r/docs

A hands-on tour of Gno realm development. Each link below is a live realm; click the
**[source]** button on any page to read its on-chain code. Reading top-to-bottom gives
you a coherent course from first principles through advanced patterns.

## Getting Started

Start here — what a realm is and how its page is produced.

- [Hello World](/r/docs/hello) — the minimal realm: ^Render(path)^, path routing, safe output.
- [Viewing Source](/r/docs/source) — read any realm's on-chain code from gnoweb.
- [Testing realms (TDD)](/r/docs/tdd) — write ^_test.gno^ tests: fake the caller and coins, red-green-refactor a realm.

## Rendering & gnoweb

Produce rich gnoweb pages and let users act from them.

- [Markdown](/r/docs/markdown) — Gno-flavored markdown and every gnoweb extension (alerts, columns, forms, images, mentions).
- [Programmatic Markdown (p/moul/md)](/r/docs/moul_md) — build that same markdown in code.
- [Routing Render paths](/r/docs/routing) — dispatch sub-pages with the ^p/nt/mux/v0^ package.
- [Buttons & Forms](/r/docs/buttons) — turn Render output into transactions with gno-form and txlink.
- [Transaction Links (p/moul/txlink)](/r/docs/txlink) — build clickable tx links, including cross-realm calls.

## State & Data Structures

Persist data that grows, and make it observable.

- [Emitting Events](/r/docs/events) — expose state changes to off-chain consumers with ^chain.Emit^.
- [Service Registry](/r/docs/registry) — register and look up entries with owner-namespaced ^avl.Tree^ keys, returning copies to protect stored state.

## Interrealm & Access Control

The cur capability, crossing functions, and who is allowed to call what.

- [Crossing functions](/r/docs/crossing) — how ^cross(cur)^, ^cur realm^, and ^cur.Previous()^ identify the caller, and when a secondary ^rlm^ needs ^IsCurrent()^.
- [Complex arguments: MsgCall vs MsgRun](/r/docs/complexargs) — when you need MsgRun and the ^cross(cur)^ call shape.
- [Realm composition](/r/docs/composition) — a parent realm cross-calls sub-realms that each restrict writes to their allowed caller.
- [Resolving users](/r/docs/resolveusers) — read identity from ^r/sys/users^.
- [User profiles](/r/docs/userprofile) — per-user profile data on top of ^r/sys/users^ and ^r/demo/profile^.
- [Private realms](/r/docs/private) — the gnomod private flag and realm isolation.

## Advanced / Patterns

Capstones that combine everything — and the Gno-specific reflexes that differ from Solidity.

- [MiniSocial v1](/r/docs/minisocial/v1) — a social feed backed by a slice.
- [MiniSocial v2](/r/docs/minisocial/v2) — same feed rewritten with avl + pager + ownable + sys/users.
- [Factory pattern](/r/docs/factory) — a constructor that returns different token types behind one interface.
- [Common mistakes](/r/docs/mistakes) — overflow, error-vs-panic reverts, middleman calls, address validation.
- [Solidity patterns in Gno](/r/docs/soliditypatterns) — porting [counter](/r/docs/soliditypatterns/counter), [ownable](/r/docs/soliditypatterns/ownable), [state-lock](/r/docs/soliditypatterns/statelock), [banker/payments](/r/docs/soliditypatterns/banker), and [reentrancy](/r/docs/soliditypatterns/reentrancy); why msg.sender does NOT map to a stack-walk.

## Other resources

- [Official documentation](https://docs.gno.land)
- [Suggestions / meta issue](https://github.com/gnolang/gno/issues/3292)
`
	return strings.ReplaceAll(content, "^", "`")
}
