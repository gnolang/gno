package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type (
	// CategoryLink defines a generated-index link.
	CategoryLink struct {
		Type string `json:"type"`
		Slug string `json:"slug"`
	}

	// Category represents a category in the sidebar.
	Category struct {
		Type      string        `json:"type"`
		Label     string        `json:"label"`
		Items     []interface{} `json:"items,omitempty"`
		Link      *CategoryLink `json:"link,omitempty"`
		Collapsed bool          `json:"collapsed,omitempty"`
	}

	// ExternalLink represents an external link.
	ExternalLink struct {
		Type  string `json:"type"`
		Label string `json:"label"`
		Href  string `json:"href"`
	}

	// Sidebar is the final JSON export structure.
	Sidebar struct {
		Sidebar []interface{} `json:"tutorialSidebar"`
	}

	// RootConfig holds the configuration (path to the README file).
	RootConfig struct {
		IndexPath string
	}
)

var rootFlagSet = flag.NewFlagSet("parser", flag.ExitOnError)

func parseRootConfig(args []string) (*RootConfig, error) {
	var cfg RootConfig
	rootFlagSet.StringVar(&cfg.IndexPath, "path", "", "path to the docs README index file")
	err := ff.Parse(rootFlagSet, args)
	if err != nil {
		return nil, fmt.Errorf("unable to parse flags: %w", err)
	}
	return &cfg, nil
}

func main() {
	cfg, err := parseRootConfig(os.Args[1:])
	if err != nil {
		panic(err)
	}
	parseAndOutput(cfg)
}

// parseAndOutput traverses the root AST and, for each level 2 heading,
// collects all subsequent nodes (typically lists) until the next level 2 heading.
// These lists are merged into a section. Finally, the sidebar is built as an array
// with "README" as the first element and an array of sections as the second element.
func parseAndOutput(cfg *RootConfig) {
	if cfg.IndexPath == "" {
		panic("no index file path provided")
	}
	source, err := os.ReadFile(cfg.IndexPath)
	if err != nil {
		panic(err)
	}
	gm := goldmark.New()
	rootNode := gm.Parser().Parse(text.NewReader(source))

	var sections []interface{}
	// Iterate over the root nodes.
	for n := rootNode.FirstChild(); n != nil; n = n.NextSibling() {
		// Look for a level 2 heading (##).
		if heading, ok := n.(*ast.Heading); ok && heading.Level == 2 {
			label := string(heading.Text(source))
			sec := &Category{
				Type:  "category",
				Label: label,
			}
			// Scan subsequent nodes until the next level 2 heading or end.
			for m := n.NextSibling(); m != nil; m = m.NextSibling() {
				// If another level 2 heading is found, stop collecting.
				if h, ok := m.(*ast.Heading); ok && h.Level == 2 {
					break
				}
				// If it's a list, process it and append its items to the section.
				if list, ok := m.(*ast.List); ok {
					items, err := processList(list, source, 1)
					if err != nil {
						panic(err)
					}
					sec.Items = append(sec.Items, items...)
				}
			}
			sections = append(sections, sec)
		}
	}

	var sb Sidebar
	// Build the final structure with "README" as the first element,
	// then the sections array.
	sb.Sidebar = []interface{}{"README", sections}

	res, err := json.MarshalIndent(&sb, "", "  ")
	if err != nil {
		panic(err)
	}
	output := strings.NewReplacer(
		`\u0026`, `&`,
		`\u003c`, `<`,
		`\u003e`, `>`,
	).Replace(string(res))
	fmt.Println(output)
}

// processList processes a List node and returns its items (categories or links).
func processList(list *ast.List, source []byte, depth int) ([]interface{}, error) {
	var items []interface{}
	for li := list.FirstChild(); li != nil; li = li.NextSibling() {
		if liItem, ok := li.(*ast.ListItem); ok {
			item, err := processListItem(liItem, source, depth)
			if err != nil {
				return nil, err
			}
			if item != nil {
				items = append(items, item)
			}
		}
	}
	return items, nil
}

// processListItem processes a ListItem node:
//   - Recursively searches for the first link in the ListItem's content (excluding sub-lists).
//   - If a sub-list exists in the item, creates a category with a "generated-index" link pointing to the extracted link.
//   - Otherwise, returns simply the link.
func processListItem(li *ast.ListItem, source []byte, depth int) (interface{}, error) {
	var link *ast.Link
	// Recursively search for a link in the children of the ListItem, ignoring sub-lists.
	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.List); ok {
			continue
		}
		link = findLink(child)
		if link != nil {
			break
		}
	}
	if link == nil {
		return nil, nil
	}
	label := string(link.Text(source))
	slug := string(link.Destination)

	// Check for the presence of a sub-list within the same ListItem.
	var hasSublist bool
	var sublist *ast.List
	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		if lst, ok := child.(*ast.List); ok {
			hasSublist = true
			sublist = lst
			break
		}
	}

	if hasSublist {
		cat := &Category{
			Type:      "category",
			Label:     label,
			Collapsed: depth >= 1,
		}
		cat.Link = &CategoryLink{
			Type: "generated-index",
			Slug: slug,
		}
		subItems, err := processList(sublist, source, depth+1)
		if err != nil {
			return nil, err
		}
		cat.Items = subItems
		return cat, nil
	}
	return extractLink(link, source), nil
}

// findLink recursively searches a node for an ast.Link.
func findLink(n ast.Node) *ast.Link {
	if l, ok := n.(*ast.Link); ok {
		return l
	}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if found := findLink(child); found != nil {
			return found
		}
	}
	return nil
}

// extractLink converts a Link node into an external link (if the URL starts with "http://" or "https://")
// or returns a string (the slug without extension).
func extractLink(n ast.Node, source []byte) interface{} {
	astLink, ok := n.(*ast.Link)
	if !ok {
		return nil
	}
	link := string(astLink.Destination)
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return ExternalLink{
			Type:  "link",
			Label: string(astLink.Text(source)),
			Href:  link,
		}
	}
	return strings.Split(link, ".")[0]
}
