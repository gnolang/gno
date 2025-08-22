// This file is a minimal version of https://github.com/abhinav/goldmark-toc

package markdown

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

const MaxDepth = 6

type Toc struct {
	Items []*TocItem
}

type TocItem struct {
	// Title of this item in the table of contents.
	//
	// This may be blank for items that don't refer to a heading, and only
	// have sub-items.
	Title []byte

	// ID is the identifier for the heading that this item refers to. This
	// is the fragment portion of the link without the "#".
	//
	// This may be blank if the item doesn't have an id assigned to it, or
	// if it doesn't have a title.
	//
	// Enable AutoHeadingID in your parser if you expected these to be set
	// but they weren't.
	ID []byte

	// Items references children of this item.
	//
	// For a heading at level 3, Items, contains the headings at level 4
	// under that section.
	Items []*TocItem
}

func (i TocItem) Anchor() string {
	return "#" + string(i.ID)
}

type TocOptions struct {
	MinDepth, MaxDepth int
}

func TocInspect(n ast.Node, src []byte, opts TocOptions) (Toc, error) {
	// Appends an empty subitem to the given node
	// and returns a reference to it.
	appendChild := func(n *TocItem) *TocItem {
		child := new(TocItem)
		n.Items = append(n.Items, child)
		return child
	}

	// Returns the last subitem of the given node,
	// creating it if necessary.
	lastChild := func(n *TocItem) *TocItem {
		if len(n.Items) > 0 {
			return n.Items[len(n.Items)-1]
		}
		return appendChild(n)
	}

	var root TocItem

	stack := []*TocItem{&root} // inv: len(stack) >= 1
	err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Skip non-heading node
		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		if opts.MinDepth > 0 && heading.Level < opts.MinDepth {
			return ast.WalkSkipChildren, nil
		}

		if opts.MaxDepth > 0 && heading.Level > opts.MaxDepth {
			return ast.WalkSkipChildren, nil
		}

		// The heading is deeper than the current depth.
		// Append empty items to match the heading's level.
		for len(stack) < heading.Level {
			parent := stack[len(stack)-1]
			stack = append(stack, lastChild(parent))
		}

		// The heading is shallower than the current depth.
		// Move back up the stack until we reach the heading's level.
		if len(stack) > heading.Level {
			stack = stack[:heading.Level]
		}

		parent := stack[len(stack)-1]
		target := lastChild(parent)
		if len(target.Title) > 0 || len(target.Items) > 0 {
			target = appendChild(parent)
		}

		target.Title = util.UnescapePunctuations(ExtractText(heading, src))
		if id, ok := n.AttributeString("id"); ok {
			target.ID, _ = id.([]byte)
		}

		return ast.WalkSkipChildren, nil
	})

	root.Items = compactItems(root.Items)

	return Toc{Items: root.Items}, err
}

// compactItems removes items with no titles
// from the given list of items.
//
// Children of removed items will be promoted to the parent item.
func compactItems(items []*TocItem) []*TocItem {
	result := make([]*TocItem, 0)
	for _, item := range items {
		if len(item.Title) == 0 {
			result = append(result, compactItems(item.Items)...)
			continue
		}

		item.Items = compactItems(item.Items)
		result = append(result, item)
	}

	return result
}
