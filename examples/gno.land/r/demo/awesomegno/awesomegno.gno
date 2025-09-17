package awesomegno

import (
	"std"
	"time"

	"gno.land/p/moul/authz"
	"gno.land/p/moul/fifo"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/rotree"
	"gno.land/p/nt/seqid"

	"gno.land/r/leon/hor"
)

var (
	items      = avl.NewTree()
	categories = avl.NewTree()

	Items      = rotree.Wrap(items, nil)
	Categories = rotree.Wrap(categories, nil)

	latestItems      = fifo.New(5)
	latestCategories = fifo.New(5)

	nextItemID     seqid.ID
	nextCategoryID seqid.ID

	authority = authz.NewMemberAuthority(
		std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // leon
		std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), // moul
	)
	Auth = authz.NewWithAuthority(authority)
)

func init() {
	hor.Register(cross, "AwesomeGno", "Awesome Gno is a registry for gnowesome dApps, tools and resources on gno.land")
}

func AddItem(cur realm, title, description, url, authorAddress string) {
	if title == "" {
		panic(ErrItemTitleEmpty)
	}

	author := std.Address(authorAddress)
	if !author.IsValid() {
		panic(ErrItemInvalidAddress + ": " + authorAddress)
	}

	if err := Auth.DoByPrevious("add_item", func() error {
		addItem(title, description, url, author)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateItemWithCategories(cur realm, itemID seqid.ID, title, description, url string, categories ...seqid.ID) {
	emptyAddrs := []std.Address{}
	if err := Auth.DoByPrevious("update_item", func() error {
		updateItem(itemID, title, description, url, emptyAddrs, categories...)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateItem(cur realm, itemID seqid.ID, title, description, url string) {
	emptyAddrs := []std.Address{}
	emptyCategories := []seqid.ID{}
	if err := Auth.DoByPrevious("update_item_without_categories", func() error {
		updateItem(itemID, title, description, url, emptyAddrs, emptyCategories...)
		return nil
	}); err != nil {
		panic(err)
	}
}

func AddItemToCategory(cur realm, itemID seqid.ID, categoryID seqid.ID) {
	if err := Auth.DoByPrevious("add_item_to_category", func() error {
		addItemToCategory(itemID, categoryID)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateItemsAuthors(cur realm, itemID seqid.ID, authors []std.Address) {
	if len(authors) == 0 {
		panic(ErrItemAuthorsEmpty)
	}

	for _, author := range authors {
		if !author.IsValid() {
			panic(ErrItemInvalidAddress + ": " + author.String())
		}
	}

	if err := Auth.DoByPrevious("update_items_authors", func() error {
		updateItem(itemID, "", "", "", authors)
		return nil
	}); err != nil {
		panic(err)
	}
}

func DeleteItem(cur realm, itemID seqid.ID) {
	if err := Auth.DoByPrevious("delete_item", func() error {
		deleteItem(itemID)
		return nil
	}); err != nil {
		panic(err)
	}
}

func AddCategory(cur realm, name, description string) {
	if name == "" {
		panic(ErrCategoryNameEmpty)
	}

	if err := Auth.DoByPrevious("add_category", func() error {
		addCategory(name, description)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateCategory(cur realm, categoryID seqid.ID, name, description string) {
	if err := Auth.DoByPrevious("update_category", func() error {
		updateCategory(categoryID, name, description)
		return nil
	}); err != nil {
		panic(err)
	}
}

func DeleteCategory(cur realm, categoryID seqid.ID) {
	if err := Auth.DoByPrevious("delete_category", func() error {
		deleteCategory(categoryID)
		return nil
	}); err != nil {
		panic(err)
	}
}

func SetLatestItemsSize(cur realm, size int) {
	if err := Auth.DoByPrevious("set_latest_items_size", func() error {
		latestItems.SetMaxSize(size)
		return nil
	}); err != nil {
		panic(err)
	}
}

func SetLatestCategoriesSize(cur realm, size int) {
	if err := Auth.DoByPrevious("set_latest_categories_size", func() error {
		latestCategories.SetMaxSize(size)
		return nil
	}); err != nil {
		panic(err)
	}
}

func GetLatestItems(cur realm) []seqid.ID {
	entries := latestItems.Entries()
	result := make([]seqid.ID, len(entries))
	for i, entry := range entries {
		result[i] = entry.(seqid.ID)
	}
	return result
}

func GetLatestCategories(cur realm) []seqid.ID {
	entries := latestCategories.Entries()
	result := make([]seqid.ID, len(entries))
	for i, entry := range entries {
		result[i] = entry.(seqid.ID)
	}
	return result
}

func addItem(name, description, url string, author std.Address) seqid.ID {
	item := &Item{
		ID:          nextItemID.Next(),
		Title:       name,
		Description: description,
		Authors:     []std.Address{author},
		URL:         url,
		CreatedAt:   time.Now(),
		Categories:  []seqid.ID{},
	}

	items.Set(item.ID.String(), item)

	latestItems.Append(item.ID)

	std.Emit("ItemAdded",
		"id", item.ID.String(),
		"name", name,
	)

	return item.ID
}

func updateItem(itemID seqid.ID, title, description, url string, authors []std.Address, categoryIDs ...seqid.ID) {
	item, found := items.Get(itemID.String())
	if !found {
		panic(ErrItemNotFound + ": " + itemID.String())
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
		for _, oldCatID := range d.Categories {
			if cat, found := categories.Get(oldCatID.String()); found {
				cat.(*Category).items.Remove(itemID.String())
			}
		}

		d.Categories = categoryIDs
		for _, categoryID := range categoryIDs {
			category, found := categories.Get(categoryID.String())
			if !found {
				panic(ErrCategoryNotFound + ": " + categoryID.String())
			}
			category.(*Category).items.Set(itemID.String(), item)
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

func addItemToCategory(itemID seqid.ID, categoryID seqid.ID) {
	itemVal, found := items.Get(itemID.String())
	if !found {
		panic(ErrItemNotFound + ": " + itemID.String())
	}
	item := itemVal.(*Item)

	for _, cat := range item.Categories {
		if cat == categoryID {
			return
		}
	}

	newCategories := append(item.Categories, categoryID)

	updateItem(itemID, "", "", "", nil, newCategories...)
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
		panic(ErrItemNotFound + ": " + itemID.String())
	}

	for _, catID := range item.(*Item).Categories {
		if cat, found := categories.Get(catID.String()); found {
			cat.(*Category).items.Remove(itemID.String())
		}
	}

	items.Remove(itemID.String())

	for i := 0; i < latestItems.Size(); i++ {
		if latestItems.Get(i).(seqid.ID) == itemID {
			latestItems.Delete(i)
			break
		}
	}

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

	latestCategories.Append(category.ID)

	std.Emit("CategoryAdded",
		"id", category.ID.String(),
		"name", name,
	)

	return category.ID
}

func updateCategory(categoryID seqid.ID, name, description string) {
	category, found := categories.Get(categoryID.String())
	if !found {
		panic(ErrCategoryNotFound + ": " + categoryID.String())
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
		panic(ErrCategoryNotFound + ": " + categoryID.String())
	}

	category.(*Category).items.Iterate("", "", func(itemID string, value any) bool {
		if item, ok := value.(*Item); ok {
			newCategories := make([]seqid.ID, 0)
			for _, catID := range item.Categories {
				if catID != categoryID {
					newCategories = append(newCategories, catID)
				}
			}
			item.Categories = newCategories
		}
		return false
	})

	categories.Remove(categoryID.String())

	for i := 0; i < latestCategories.Size(); i++ {
		if latestCategories.Get(i).(seqid.ID) == categoryID {
			latestCategories.Delete(i)
			break
		}
	}

	std.Emit("CategoryDeleted", "id", categoryID.String())
}
