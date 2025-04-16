package awesomegno

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/authz"
	"gno.land/r/leon/hor"
)

var (
	items      = avl.NewTree()
	categories = avl.NewTree()

	Items      = rotree.Wrap(items, nil)
	Categories = rotree.Wrap(categories, nil)

	nextItemID     seqid.ID
	nextCategoryID seqid.ID

	authority = authz.NewMemberAuthority(
		std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // leon
		std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), // moul
		std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"), // matijamarjanovic (for testing)
	)
	Auth = authz.NewWithAuthority(authority)
)

func init() {
	hor.Register("AwesomeGno", "Awesome Gno is a registry for gnowesome dApps, tools and resources on gno.land")
}

func AddItem(title, description, url, authorAddress string) seqid.ID {
	if title == "" {
		panic("item title cannot be empty")
	}

	var itemID seqid.ID
	if err := Auth.Do("add_item", func() error {
		itemID = addItem(title, description, url, std.Address(authorAddress))
		return nil
	}); err != nil {
		panic(err)
	}

	return itemID
}

func UpdateItem(itemID seqid.ID, title, description, url string, categories ...seqid.ID) {
	if err := Auth.Do("update_item", func() error {
		updateItem(itemID, title, description, url, []std.Address{}, categories...)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateItemsAuthors(itemID seqid.ID, authors []std.Address) {
	if err := Auth.Do("update_items_authors", func() error {
		updateItem(itemID, "", "", "", authors, []seqid.ID{}...)
		return nil
	}); err != nil {
		panic(err)
	}
}

func DeleteItem(itemID seqid.ID) {
	if err := Auth.Do("delete_item", func() error {
		deleteItem(itemID)
		return nil
	}); err != nil {
		panic(err)
	}
}

func AddCategory(name, description string) seqid.ID {
	if name == "" {
		panic("category name cannot be empty")
	}

	var categoryID seqid.ID
	if err := Auth.Do("add_category", func() error {
		categoryID = addCategory(name, description)
		return nil
	}); err != nil {
		panic(err)
	}

	return categoryID
}

func UpdateCategory(categoryID seqid.ID, name, description string) {
	if err := Auth.Do("update_category", func() error {
		updateCategory(categoryID, name, description)
		return nil
	}); err != nil {
		panic(err)
	}
}
func DeleteCategory(categoryID seqid.ID) {
	if err := Auth.Do("delete_category", func() error {
		deleteCategory(categoryID)
		return nil
	}); err != nil {
		panic(err)
	}
}

func addItem(name, description, url string, author std.Address) seqid.ID {
	item := &Item{
		ID:          nextItemID.Next(),
		Title:       name,
		Description: description,
		Authors:     []std.Address{author},
		URL:         url,
		CreatedAt:   time.Now(),
		categories:  avl.NewTree(),
	}

	items.Set(item.ID.String(), item)

	std.Emit("ItemAdded",
		"id", item.ID.String(),
		"name", name,
	)

	return item.ID
}

func updateItem(itemID seqid.ID, title, description, url string, authors []std.Address, categoryIDs ...seqid.ID) {
	item, found := items.Get(itemID.String())
	if !found {
		panic("item not found: " + itemID.String())
	}

	d := item.(*Item)

	if title != "" {
		d.Title = title
	}
	if description != "" {
		d.Description = description
	}
	if url != "" {
		d.URL = url
	}
	if len(authors) > 0 {
		d.Authors = authors
	}

	if len(categoryIDs) > 0 {
		for _, categoryID := range categoryIDs {
			category, found := categories.Get(categoryID.String())
			if !found {
				panic("category not found: " + categoryID.String())
			}
			category.(*Category).items.Set(itemID.String(), item)
			d.categories.Set(categoryID.String(), category)
		}
	}

	categoriesstr := ""
	for _, categoryID := range categoryIDs {
		categoriesstr += categoryID.String() + ","
	}

	std.Emit("ItemUpdated",
		"id", itemID.String(),
		"title", d.Title,
		"description", d.Description,
		"url", d.URL,
		"authors", authorsToString(d.Authors),
		"categories", categoriesstr,
	)
}

func authorsToString(authors []std.Address) string {
	result := ""
	for i, author := range authors {
		if i > 0 {
			result += ","
		}
		result += string(author)
	}
	return result
}

func deleteItem(itemID seqid.ID) {
	item, found := items.Get(itemID.String())
	if !found {
		panic("item not found: " + itemID.String())
	}

	item.(*Item).categories.Iterate("", "", func(categoryID string, value any) bool {
		category := value.(*Category)
		category.items.Remove(itemID.String())
		return false
	})

	items.Remove(itemID.String())

	std.Emit("ItemDeleted", "id", itemID.String())
}

func addCategory(name, description string) seqid.ID {
	category := &Category{
		ID:          nextCategoryID.Next(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		items:       avl.NewTree(),
	}

	categories.Set(category.ID.String(), category)

	std.Emit("CategoryAdded",
		"id", category.ID.String(),
		"name", name,
	)

	return category.ID
}

func updateCategory(categoryID seqid.ID, name, description string) {
	category, found := categories.Get(categoryID.String())
	if !found {
		panic("category not found: " + categoryID.String())
	}

	c := category.(*Category)

	if name != "" {
		c.Name = name
	}
	if description != "" {
		c.Description = description
	}

	std.Emit("CategoryUpdated",
		"id", categoryID.String(),
		"name", c.Name,
		"description", c.Description,
	)
}

func deleteCategory(categoryID seqid.ID) {
	category, found := categories.Get(categoryID.String())
	if !found {
		panic("category not found: " + categoryID.String())
	}

	category.(*Category).items.Iterate("", "", func(itemID string, value any) bool {
		item := value.(*Item)
		item.categories.Remove(categoryID.String())
		return false
	})

	categories.Remove(categoryID.String())

	std.Emit("CategoryDeleted", "id", categoryID.String())
}
