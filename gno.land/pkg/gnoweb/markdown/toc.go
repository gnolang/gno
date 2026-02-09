// This file is a minimal version of https://github.com/abhinav/goldmark-toc

package markdown

import (
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

const MaxDepth = 6

type Toc struct {
	Items []*components.TocItem
}

// compactItems removes items with no titles from the given list of items.
// Children of removed items will be promoted to the parent item.
func compactItems(items []*components.TocItem) []*components.TocItem {
	result := make([]*components.TocItem, 0, len(items))
	for _, item := range items {
		if item.Title == "" {
			result = append(result, compactItems(item.Items)...)
			continue
		}
		item.Items = compactItems(item.Items)
		result = append(result, item)
	}
	return result
}

type TocOptions struct {
	MinDepth, MaxDepth int
}

func TocInspect(n ast.Node, src []byte, opts TocOptions) (Toc, error) {
	// Appends an empty subitem to the given node
	// and returns a reference to it.
	appendChild := func(n *components.TocItem) *components.TocItem {
		child := new(components.TocItem)
		n.Items = append(n.Items, child)
		return child
	}

	// Returns the last subitem of the given node,
	// creating it if necessary.
	lastChild := func(n *components.TocItem) *components.TocItem {
		if len(n.Items) > 0 {
			return n.Items[len(n.Items)-1]
		}
		return appendChild(n)
	}

	var root components.TocItem

	stack := []*components.TocItem{&root} // inv: len(stack) >= 1
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
		if target.Title != "" || len(target.Items) > 0 {
			target = appendChild(parent)
		}

		target.Title = string(util.UnescapePunctuations(nodeText(src, heading)))

		if id, ok := n.AttributeString("id"); ok {
			if idBytes, ok := id.([]byte); ok {
				target.ID = string(idBytes)
			}
		}

		return ast.WalkSkipChildren, nil
	})

	root.Items = compactItems(root.Items)

	return Toc{Items: root.Items}, err
}
