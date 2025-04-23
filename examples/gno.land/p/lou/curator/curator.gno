package curator

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/lou/card"
	"gno.land/p/moul/md"
)

var (
	ErrCollectionExists   = errors.New("Collection already exists")
	ErrNotCollectionOwner = errors.New("Only the owner can modify this collection")
	ErrCollectionMissing  = errors.New("Collection not found")
	ErrItemAlrExists      = errors.New("An item with this title already exists in the collection")
	ErrItemAlrLinked      = errors.New("An item with this path already exists in the collection")
)

type Collection struct {
	id          seqid.ID
	title       string
	description string
	owner       std.Address
	items       []*card.Card
	tags        []string

	gridMode bool
}

func (c Collection) ID() seqid.ID {
	return c.id
}

func (c Collection) Title() string {
	return c.title
}

func (c Collection) Description() string {
	return c.description
}

func (c Collection) Owner() std.Address {
	return c.owner
}

func (c Collection) Items() []*card.Card {
	return c.items
}

func (c Collection) Tags() []string {
	return c.tags
}

func (c Collection) GridMode() bool {
	return c.gridMode
}

func (c *Collection) ToggleGridMode() {
	c.gridMode = !c.gridMode
}

func AddToCollection(c *Collection, title, path string, tags []string) (*Collection, error) {
	if c.owner != std.PreviousRealm().Address() {
		return nil, ErrNotCollectionOwner
	}
	formattedPath := strings.TrimPrefix(path, "gno.land")

	for _, item := range c.Items() {
		if item.Header() == title {
			return nil, ErrItemAlrExists
		}

		for _, line := range item.Content() {
			if strings.Contains(line, formattedPath) {
				return nil, ErrItemAlrLinked
			}
		}
	}
	item := createCard(title, path, tags, c.gridMode)
	c.items = append(c.items, item)

	return c, nil
}

func createCard(title, path string, tags []string, gridMode bool) *card.Card {
	namespace := strings.Split(path, "/")[2]
	formattedPath := strings.TrimPrefix(path, "gno.land")
	linkTitle := md.Link(title, formattedPath)
	formattedTags := formatTags(tags, gridMode)

	return card.NewCard(linkTitle, []string{"by " + namespace, formattedPath, formattedTags, "\n"}, 0, 0)
}

func NewCollection(title,
	description string,
	tags []string,
	id seqid.ID,
	owner std.Address,
) (*Collection, error) {
	return &Collection{
		id:          id,
		title:       title,
		description: description,
		owner:       owner,
		items:       []*card.Card{},
		tags:        tags,
		gridMode:    false,
	}, nil
}

func CreateCollection(
	id seqid.ID,
	title,
	description string,
	tags []string,
	owner std.Address,
) (*Collection, error) {
	col, err := NewCollection(title, description, tags, id, owner)
	if err != nil {
		return col, err
	}
	return col, nil
}

func formatTags(tags []string, gridMode bool) string {
	out := ""

	if gridMode {
		out += "- Tags: "
	} else {
		out += "Tags: "
	}

	for i, tag := range tags {
		out += md.InlineCode(tag)
		if i != len(tags)-1 {
			out += ", "
		}
	}
	return out + "\n\n"
}

func (c Collection) PreviewCollection(gridMode bool) string {
	path := ufmt.Sprintf("/r/lou/collections:%s", c.ID().String())
	linkTitle := md.Link(c.title, path)
	tags := formatTags(c.tags, gridMode)

	if gridMode {
		card := card.NewCard(linkTitle, []string{c.description, tags, "\n"}, 36, 0)
		return card.Render()
	}
	card := card.NewCard(linkTitle, []string{c.description, tags, "\n"}, 0, 0)
	return card.Render()
}

func RenderCollection(c *Collection) string {
	colCount := 0
	out := ""

	for _, card := range c.items {
		if c.gridMode {
			card.SetWidth(24)
		}
		out += card.Render()
		out += md.HorizontalRule()

		colCount++
		if colCount%3 == 0 {
			out += "\n</gno-columns>\n\n"
			if colCount != len(c.items) {
				out += "<gno-columns>\n"
			}
		} else if c.gridMode {
			out += "\n|||\n"
		}
	}

	return out
}
