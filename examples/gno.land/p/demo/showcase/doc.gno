// Package showcase is a kitchen-sink package for documentation demos.
//
// It declares one of every exported symbol kind so gnoweb's source overview
// renders all of its sections — Constants, Variables, Functions and Types —
// and the kind glyph that prefixes each symbol.
//
// # Example
//
// Build a catalog, add an item and summarize it:
//
//	c := showcase.Catalog{}
//	c["widget"] = showcase.New("Widget", 1500)
//	println(showcase.Summarize(c))
//
// # Symbol kinds on display
//
//   - struct, interface and a defined primitive: [Item], [Storer], [Name]
//   - slice, map, func and pointer: [Tags], [Catalog], [Hook], [ItemRef]
//   - package functions, plus methods grouped under their receiver type
//
// See [Item] for the core record and [Catalog.Total] for an example method.
// Online docs live at https://docs.gno.land for more.
package showcase
