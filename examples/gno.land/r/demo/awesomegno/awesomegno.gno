package awesomegno

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/authz"
	"gno.land/r/leon/hof"
)

var (
	dapps      = avl.NewTree()
	categories = avl.NewTree()

	Dapps      = rotree.Wrap(dapps, nil)
	Categories = rotree.Wrap(categories, nil)

	nextDappID     seqid.ID
	nextCategoryID seqid.ID

	authority = authz.NewMemberAuthority(
		std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // leon
		std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), // moul
		std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"), // matijamarjanovic (for testing)
	)
	Auth = authz.NewWithAuthority(authority)
)

func init() {
	hof.Register("AwesomeGno", "Awesome Gno is a registry for gnowesome dApps on gno.land")
}

func AddDapp(title, description, url, authorAddress string) seqid.ID {
	if title == "" {
		panic("dapp title cannot be empty")
	}

	if !std.Address(authorAddress).IsValid() {
		panic("invalid author address")
	}

	var dappID seqid.ID
	if err := Auth.Do("add_dapp", func() error {
		dappID = addDapp(title, description, url, std.Address(authorAddress))
		return nil
	}); err != nil {
		panic(err)
	}

	return dappID
}

func UpdateDapp(dappID seqid.ID, title, description, url string, categories ...seqid.ID) {
	if err := Auth.Do("update_dapp", func() error {
		updateDapp(dappID, title, description, url, categories...)
		return nil
	}); err != nil {
		panic(err)
	}
}

func DeleteDapp(dappID seqid.ID) {
	if err := Auth.Do("delete_dapp", func() error {
		deleteDapp(dappID)
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

func addDapp(name, description, url string, author std.Address) seqid.ID {
	dapp := &Dapp{
		ID:          nextDappID.Next(),
		Title:       name,
		Description: description,
		Author:      author,
		URL:         url,
		CreatedAt:   time.Now(),
		categories:  avl.NewTree(),
	}

	dapps.Set(dapp.ID.String(), dapp)

	std.Emit("DappAdded",
		"id", dapp.ID.String(),
		"name", name,
	)

	return dapp.ID
}

func updateDapp(dappID seqid.ID, title, description, url string, categoryIDs ...seqid.ID) {
	dapp, found := dapps.Get(dappID.String())
	if !found {
		panic("dApp not found: " + dappID.String())
	}

	d := dapp.(*Dapp)

	if title != "" {
		d.Title = title
	}
	if description != "" {
		d.Description = description
	}
	if url != "" {
		d.URL = url
	}

	if len(categoryIDs) > 0 {
		for _, categoryID := range categoryIDs {
			category, found := categories.Get(categoryID.String())
			if !found {
				panic("category not found: " + categoryID.String())
			}
			category.(*Category).dapps.Set(dappID.String(), dapp)
			d.categories.Set(categoryID.String(), category)
		}
	}

	categoriesstr := ""
	for _, categoryID := range categoryIDs {
		categoriesstr += categoryID.String() + ","
	}

	std.Emit("DappUpdated",
		"id", dappID.String(),
		"title", d.Title,
		"description", d.Description,
		"url", d.URL,
		"categories", categoriesstr,
	)
}

func deleteDapp(dappID seqid.ID) {
	dapp, found := dapps.Get(dappID.String())
	if !found {
		panic("dApp not found: " + dappID.String())
	}

	dapp.(*Dapp).categories.Iterate("", "", func(categoryID string, value any) bool {
		category := value.(*Category)
		category.dapps.Remove(dappID.String())
		return false
	})

	dapps.Remove(dappID.String())

	std.Emit("DappDeleted", "id", dappID.String())
}

func addCategory(name, description string) seqid.ID {
	category := &Category{
		ID:          nextCategoryID.Next(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		dapps:       avl.NewTree(),
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

	category.(*Category).dapps.Iterate("", "", func(dappID string, value any) bool {
		dapp := value.(*Dapp)
		dapp.categories.Remove(categoryID.String())
		return false
	})

	categories.Remove(categoryID.String())

	std.Emit("CategoryDeleted", "id", categoryID.String())
}
