/*
 DISCLAIMER:
 All exported functions in this realm when called, panic wih the same error message.
 Output returned by using MsgCall tx:

 gnokey maketx call
 / -pkgpath "gno.land/r/demo/awesomegno"
 /-func "AddItem"
 /-args "title"
 /-args "desc"
 /-args "url"
 /-args "g1kjsl2ungmc95mgluq96w8dqlep8d4n8cxdfthk"
 /-gas-fee 1000000ugnot
 /-gas-wanted 500000000
 /-broadcast
 /-chainid "dev"
 /-remote "tcp://127.0.0.1:26657"
 mainwallet

 Enter password.
 TX HASH:    
 INFO:       vm.version=develop
 --= Error =--
 Data: should not happen, should be heapItemType: title<()~VPBlock(1,0)>      
 Msg Traces:
...
*/
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
	)
	Auth = authz.NewWithAuthority(authority)
)

func init() {
	cross(hor.Register)("AwesomeGno", "Awesome Gno is a registry for gnowesome dApps, tools and resources on gno.land")
}

func AddItem(title, description, url, authorAddress string) {
	crossing()

	if title == "" {
		panic("item title cannot be empty")
	}

	addr := std.Address(authorAddress)
	if err := Auth.DoByPrevious("add_item", func() error {
		addItem(title, description, url, addr)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateItem(itemID seqid.ID, title, description, url string, categories ...seqid.ID) {
	crossing()

	emptyAddrs := []std.Address{}
	if err := Auth.DoByPrevious("update_item", func() error {
		updateItem(itemID, title, description, url, emptyAddrs, categories...)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateItemsAuthors(itemID seqid.ID, authors []std.Address) {
	crossing()

	if err := Auth.DoByPrevious("update_items_authors", func() error {
		updateItem(itemID, "", "", "", authors)
		return nil
	}); err != nil {
		panic(err)
	}
}

func DeleteItem(itemID seqid.ID) {
	crossing()

	if err := Auth.DoByPrevious("delete_item", func() error {
		deleteItem(itemID)
		return nil
	}); err != nil {
		panic(err)
	}
}

func AddCategory(name, description string) {
	crossing()

	if name == "" {
		panic("category name cannot be empty")
	}

	if err := Auth.DoByPrevious("add_category", func() error {
		addCategory(name, description)
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateCategory(categoryID seqid.ID, name, description string) {
	crossing()

	if err := Auth.DoByPrevious("update_category", func() error {
		updateCategory(categoryID, name, description)
		return nil
	}); err != nil {
		panic(err)
	}
}

func DeleteCategory(categoryID seqid.ID) {
	crossing()

	if err := Auth.DoByPrevious("delete_category", func() error {
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
		Categories:  []seqid.ID{},
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
		for _, oldCatID := range d.Categories {
			if cat, found := categories.Get(oldCatID.String()); found {
				cat.(*Category).items.Remove(itemID.String())
			}
		}

		d.Categories = categoryIDs
		for _, categoryID := range categoryIDs {
			category, found := categories.Get(categoryID.String())
			if !found {
				panic("category not found: " + categoryID.String())
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

	for _, catID := range item.(*Item).Categories {
		if cat, found := categories.Get(catID.String()); found {
			cat.(*Category).items.Remove(itemID.String())
		}
	}

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

	std.Emit("CategoryDeleted", "id", categoryID.String())
}
