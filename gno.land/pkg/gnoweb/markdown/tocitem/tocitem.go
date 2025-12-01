// Package tocitem contains TocItem type and functions that were moved into
// this subpackage to resolve a circular dependency problem, because they need
// to be imported by both the markdown package and the component package, and
// the markdown package needs to access the component package to render certain
// templates (to date, only ui/command.html but likely others in the future).
//
// We could consider moving all the types into a shared/common package so they
// can be imported by both markdown and component.
package tocitem

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

// CompactItems removes items with no titles
// from the given list of items.
//
// Children of removed items will be promoted to the parent item.
func CompactItems(items []*TocItem) []*TocItem {
	result := make([]*TocItem, 0)
	for _, item := range items {
		if len(item.Title) == 0 {
			result = append(result, CompactItems(item.Items)...)
			continue
		}

		item.Items = CompactItems(item.Items)
		result = append(result, item)
	}

	return result
}
