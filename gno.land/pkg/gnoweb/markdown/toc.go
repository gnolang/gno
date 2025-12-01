// This file is a minimal version of https://github.com/abhinav/goldmark-toc

package markdown

import (
	ti "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown/tocitem"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

const MaxDepth = 6

type Toc struct {
	Items []*ti.TocItem
}

type TocOptions struct {
	MinDepth, MaxDepth int
}

func TocInspect(n ast.Node, src []byte, opts TocOptions) (Toc, error) {
	// Appends an empty subitem to the given node
	// and returns a reference to it.
	appendChild := func(n *ti.TocItem) *ti.TocItem {
		child := new(ti.TocItem)
		n.Items = append(n.Items, child)
		return child
	}

	// Returns the last subitem of the given node,
	// creating it if necessary.
	lastChild := func(n *ti.TocItem) *ti.TocItem {
		if len(n.Items) > 0 {
			return n.Items[len(n.Items)-1]
		}
		return appendChild(n)
	}

	var root ti.TocItem

	stack := []*ti.TocItem{&root} // inv: len(stack) >= 1
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

		target.Title = util.UnescapePunctuations(nodeText(src, heading))

		if id, ok := n.AttributeString("id"); ok {
			target.ID, _ = id.([]byte)
		}

		return ast.WalkSkipChildren, nil
	})

	root.Items = ti.CompactItems(root.Items)

	return Toc{Items: root.Items}, err
}
