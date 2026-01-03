package home

import "strings"

func Render(_ string) string {
	content := `# Gno Examples Documentation

Welcome to the Gno examples documentation index.
Explore various examples to learn more about Gno functionality and usage.

## Getting Started
- [Hello World](/r/docs/hello) - A simple introductory example.
- [Adder](/r/docs/adder) - An interactive example to update a number with transactions.
- [Source](/r/docs/source) - View realm source code.

## Development & Testing
- [Crossing](/r/docs/crossing) - Understanding realm boundary crossing.
- [Common Mistakes](/r/docs/mistakes) - Pitfalls to avoid: overflow, errors, OriginCaller attacks.
- [Calling non-primitive argument functions](/r/docs/complexargs) - Learn how to call functions that take in complex arguments.
- [Test-Driven Development (TDD)](/r/docs/tdd) - Writing tests before code in Gno.

## Data & Serialization
- [JSON](/r/docs/json) - Working with JSON data using ^p/onbloc/json^.
- [AVL Pager](/r/docs/avl_pager) - Paginate through AVL tree items.
- [AVL Pager + Render paths](/r/docs/avl_pager_params) - Handle render arguments with pagination.

## Design Patterns
- [Composition](/r/docs/composition) - Composing functionality from multiple realms.
- [Factory Pattern](/r/docs/factory) - Create and manage multiple object instances.
- [Registry Pattern](/r/docs/registry) - Register and discover entities in a central directory.
- [Safe Objects](/r/docs/safeobjects) - A Gno-specific pattern allowing you to expose objects with admin privileges preserved.
- [Common Solidity Patterns](/r/docs/soliditypatterns) - A list of common Solidity coding practices, converted to Gno equivalents.

## Rendering & UI
- [Buttons](/r/docs/buttons) - Add buttons to your realm's render.
- [Transaction Links](/r/docs/txlink) - Create clickable transaction links in your realm's render!
- [Optional Render](/r/docs/optional_render) - Render() is optional in realms.
- [Routing for Render paths](/r/docs/routing) - Route Render paths with the ^p/demo/mux^ package.
- [Embed images](/r/docs/img_embed) - Demonstrates how to embed an image in a realm render.
- [Markdown](/r/docs/markdown) - Documentation for Gno Flavored Markdown syntax and features.
- [p/moul/md Package Demo](/r/docs/moul_md) - Learn how to programmatically generate markdown using the p/moul/md package.
- [Chart Packages Demo](/r/docs/charts) - Create charts to display data in your realm's render.

## Users & Profiles
- [Resolving usernames and addresses](/r/docs/resolveusers) - How to resolve usernames and addresses via the ^r/sys/users^ realm.
- [User & Profile Integration](/r/docs/userprofile) - Using ^r/sys/users^ and ^r/demo/profile^ together.

## Advanced Topics
- [Emitting Gno Events](/r/docs/events) - Emit Gno Events to make life off-chain easier.
- [Private Realms](/r/docs/private) - Documentation for private realms in Gno.
- [MiniSocial](/r/docs/minisocial) - Minimalistic social media app for learning purposes.

<!-- meta issue with suggestions: https://github.com/gnolang/gno/issues/3292 -->

## Other resources

- [Official documentation](https://docs.gno.land) <!-- should be /docs with gnoweb embedding the docs/ folder. -->
`
	return strings.ReplaceAll(content, "^", "`")
}
