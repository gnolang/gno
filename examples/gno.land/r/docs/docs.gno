package docs

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/addrset"

	lc "gno.land/r/leon/config"
	mc "gno.land/r/moul/config"
)

func Render(_ string) string {
	content += `
# Gno Examples Documentation

Welcome to the Gno examples documentation index.
Explore various examples to learn more about Gno functionality and usage.

## Examples

- [Hello World](/r/docs/hello) - A simple introductory example.
- [Adder](/r/docs/adder) - An interactive example to update a number with transactions.
- [Source](/r/docs/source) - View realm source code.
- [Buttons](/r/docs/buttons) - Add buttons to your realm's render.
- [Optional Render](/r/docs/optional_render) - Render() is optional in realms.
- [AVL Pager](/r/docs/avl_pager) - Paginate through AVL tree items.  
- [AVL Pager + Render paths](/r/docs/avl_pager_params) - Handle render arguments with pagination.
- [Routing for Render paths](/r/docs/routing) - Route Render paths with the p/demo/mux package.
- [Object Apps](/r/docs/objectapps) - Create "mini-apps" with Gno objects.
- [Safe Objects](/r/docs/safeobjects) - A Gno-specific pattern allowing you to expose objects with admin privileges preserved.
- [Img Embed](/r/docs/img_embed) - Demonstrates how to embed an image.
- [Calling non-primitive argument functions](/r/docs/complexargs) - Learn how to call functions that take in complex arguments.
- [MiniSocial](/r/docs/minisocial) - Minimalistic social media app for learning purposes.
- [Markdown](/r/docs/markdown) - Documentation for Gno Flavored Markdown syntax and features.
- [Resolving usernames and addresses](/r/docs/users) - How to resolve usernames and addresses via the r/sys/users realm.
- ...
<!-- meta issue with suggestions: https://github.com/gnolang/gno/issues/3292 -->

## Other resources

- [Official documentation](https://github.com/gnolang/gno/tree/master/docs) <!-- should be /docs with gnoweb embedding the docs/ folder. -->
`
	content := ""
	content += frontMatter() // XXX: try to remove this
	return content
}

func frontMatter() string {
	return `---
Title: Welcome to the Gno examples documentation index 
Description: Welcome to the Gno examples documentation index. Explore various examples to learn more about Gno functionality and usage.
---`
}

var (
	idGen         seqid.ID
	alphabetical  = avl.NewTree() // title > *Entry
	chronological = avl.NewTree() // seqid.ID > *Entry

	admins = addrset.Set{}
)

func init() {
	admins.Add(lc.OwnableMain.Owner()) // @leohhhn
	admins.Add(mc.Addr())              // @moul
}

const docsPrefix = "gno.land/r/docs"

type Entry struct {
	id          seqid.ID
	pkgpath     string
	title       string
	description string
}

// AddEntry adds a docs entry.
// Entry needs to be under r/docs to be valid.
// Title and description are required.
func AddEntry(title, description string) {
	entryPath := std.PreviousRealm().PkgPath()

	// Disallow non-code calls and non-docs realms
	// This will automatically reject non-code calls as pkgpath of a user realm is ""
	if !strings.HasPrefix(entryPath, docsPrefix) {
		return
	}

	// Need both title and description
	if title == "" || description == "" {
		return
	}

	id := idGen.Next()
	entry := &Entry{
		id:          id,
		pkgpath:     entryPath,
		title:       title,
		description: description,
	}

	chronological.Set(id.String(), entry)
	alphabetical.Set(entryPath, entry)
}
