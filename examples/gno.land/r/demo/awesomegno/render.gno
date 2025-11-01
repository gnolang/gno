package awesomegno

import (
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/ufmt"
	"gno.land/r/sys/users"
)

const (
	itemsPath      = "items"
	categoriesPath = "categories"
	pageSize       = 10
)

func Render(path string) string {
	var out string

	switch {
	case path == "":
		out = renderHome()
	case strings.HasPrefix(path, itemsPath):
		out = renderItems(path)
	case strings.HasPrefix(path, categoriesPath):
		out = renderCategories(path)
	case strings.HasPrefix(path, "d/"):
		id := strings.TrimPrefix(path, "d/")
		out = renderSingleItem(id)
	case strings.HasPrefix(path, "c/"):
		id := strings.TrimPrefix(path, "c/")
		out = renderSingleCategory(id)
	}

	return out
}

func renderHome() string {
	out := md.H1("🚀 AwesomeGno")
	out += md.Paragraph("Welcome to AwesomeGno – a curated list of awesome items built on gno.land. Discover and explore the growing ecosystem of decentralized applications ✨")

	out += md.Paragraph(md.Link("Browse categories →", "/r/demo/awesomegno:categories"))

	out += md.H2("🆕 Recent Additions")
	var recentItems []string
	count := 0
	items.ReverseIterateByOffset(0, 5, func(key string, value interface{}) bool {
		if count >= 5 {
			return true
		}
		item := value.(*Item)
		recentItems = append(recentItems, formatItemListItem(item))
		count++
		return false
	})

	if len(recentItems) > 0 {
		out += md.BulletList(recentItems)
		out += "\n"
		out += md.Paragraph(md.Link("See all items →", "/r/demo/awesomegno:items"))
	} else {
		out += md.Italic("No items added yet. Be the first to add one! 🌱")
		out += "\n\n"
		addItemLink := txlink.NewLink("AddItem").
			AddArgs(
				"title", "",
				"description", "",
				"url", "",
				"author", "",
			).URL()
		out += md.Paragraph(md.Link("➕ Add New Item", addItemLink))
	}

	out += md.H2("📊 Quick Stats")
	stats := []string{
		ufmt.Sprintf("Total Items: %d", items.Size()),
		ufmt.Sprintf("Categories: %d", categories.Size()),
	}
	out += md.BulletList(stats)
	out += "\n"
	out += md.Italic("AwesomeGno is a community-driven project that showcases and promotes the best items built on gno.land. Each item is carefully categorized and reviewed to ensure quality and usefulness. 💎")

	return out
}

func renderItems(path string) string {
	out := md.H1("📦 Items")
	out += md.Paragraph("Browse through all the awesome items built on gno.land.")

	addItemLink := txlink.NewLink("AddItem").
		AddArgs(
			"title", "",
			"description", "",
			"url", "",
			"author", "",
		).URL()
	out += "\n\n"
	out += md.Paragraph(md.Link("➕ Add New Item", addItemLink))
	out += "\n"

	var itemItems []string
	itemPager := pager.NewPager(items, pageSize, false)
	page := itemPager.MustGetPageByPath(path)

	for _, pagerItem := range page.Items {
		if item, ok := pagerItem.Value.(*Item); ok {
			itemItems = append(itemItems, formatItemListItem(item))
		}
	}

	out += renderItemsList(itemItems, page, "No items registered yet 💤")
	return out
}

func renderCategories(path string) string {
	out := md.H1("🗂️ Categories")
	out += md.Paragraph("Browse items by category.")

	addCategoryLink := txlink.NewLink("AddCategory").
		AddArgs(
			"name", "",
			"description", "",
		).URL()
	out += "\n\n"
	out += md.Paragraph(md.Link("➕ Add New Category", addCategoryLink))
	out += "\n"
	var categoryItems []string
	categoryPager := pager.NewPager(categories, pageSize, false)
	page := categoryPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		category := item.Value.(*Category)
		categoryItems = append(categoryItems, formatCategoryListItem(category))
	}

	out += renderItemsList(categoryItems, page, "No categories available 🏷️")
	return out
}

func renderSingleItem(id string) string {
	v, found := items.Get(id)
	if !found {
		return md.Italic("Item not found")
	}

	item := v.(*Item)
	out := md.H1("📌 " + item.Title)
	out += md.Paragraph(item.Description)

	itemAuthors := []string{}
	for _, author := range item.Authors {
		displayName := author.String()
		if userData, _ := users.ResolveAny(author.String()); userData != nil && userData.Name() != "" {
			displayName = userData.Name()
		}

		authorLink := md.Link("👤 "+displayName, "/r/gnoland/users/v1:"+author.String())
		itemAuthors = append(itemAuthors, authorLink)
	}

	metadata := []string{
		ufmt.Sprintf("👥 Authors: %s", func() string {
			if len(itemAuthors) == 0 || (len(itemAuthors) == 1 && itemAuthors[0] == "") {
				return md.Italic("Not available")
			}
			return strings.Join(itemAuthors, ", ")
		}()),
		ufmt.Sprintf("📅 Created: %s", item.CreatedAt.Format("02 Jan 2006")),
		ufmt.Sprintf("🔗 URL: %s", md.Link("Visit Item", item.URL)),
	}
	out += md.BulletList(metadata)

	out += md.H2("Categories")
	var categoryLinks []string
	for _, catID := range item.Categories {
		if cat, found := categories.Get(catID.String()); found {
			category := cat.(*Category)
			categoryLinks = append(categoryLinks, md.Link(category.Name, "/r/demo/awesomegno:c/"+category.ID.String()))
		}
	}

	if len(categoryLinks) > 0 {
		out += md.BulletList(categoryLinks)
	} else {
		out += md.Italic("No categories assigned")
		out += "\n"
	}

	out += md.H2("🛠️ Management")
	out += "\n"
	out += md.Paragraph("Use these actions to manage the item:")

	updateLink := txlink.NewLink("UpdateItem").
		AddArgs(
			"itemID", item.ID.String(),
			"title", "",
			"description", "",
			"url", "",
		).URL()

	deleteLink := txlink.NewLink("DeleteItem").
		AddArgs("itemID", item.ID.String()).URL()

	actions := []string{
		md.Link("📝 Update Item", updateLink),
		md.Link("🗑️ Delete Item", deleteLink),
	}
	out += md.BulletList(actions)

	return out
}

func renderSingleCategory(id string) string {
	v, found := categories.Get(id)
	if !found {
		return md.Italic("Category not found")
	}

	category := v.(*Category)
	out := md.H1("🏷️ " + category.Name)
	out += md.Paragraph(category.Description)
	out += md.Paragraph(ufmt.Sprintf("📅 Created: %s", category.CreatedAt.Format("02 Jan 2006")))

	out += md.H2("📚 Items in this category")
	var itemsList []string
	category.items.Iterate("", "", func(key string, value interface{}) bool {
		item := value.(*Item)
		itemsList = append(itemsList, formatItemListItem(item))
		return false
	})

	if len(itemsList) > 0 {
		out += md.BulletList(itemsList)
	} else {
		out += md.Italic("No items in this category yet 💤")
		out += "\n"
	}

	out += md.H2("🛠️ Management")
	out += "\n"
	out += md.Paragraph("Use these actions to manage the category:")

	updateLink := txlink.NewLink("UpdateCategory").
		AddArgs(
			"categoryID", category.ID.String(),
			"name", "",
			"description", "",
		).URL()

	deleteLink := txlink.NewLink("DeleteCategory").
		AddArgs("categoryID", category.ID.String()).URL()

	actions := []string{
		md.Link("📝 Update Category", updateLink),
		md.Link("🗑️ Delete Category", deleteLink),
	}
	out += md.BulletList(actions)

	return out
}

func formatItemListItem(item *Item) string {
	return ufmt.Sprintf("%s - %s",
		md.Link(item.Title, "/r/demo/awesomegno:d/"+item.ID.String()),
		md.Italic(strings.Split(item.Description, ".")[0]))
}

func formatCategoryListItem(category *Category) string {
	return ufmt.Sprintf("%s (%d items) - %s",
		md.Link(category.Name, "/r/demo/awesomegno:c/"+category.ID.String()),
		category.items.Size(),
		md.Italic(category.Description))
}

func renderItemsList(items []string, page *pager.Page, emptyMessage string) string {
	var out string
	if len(items) == 0 {
		out += md.Italic(emptyMessage)
		out += "\n"
		return out
	}

	out += md.BulletList(items)
	out += "\n"
	out += md.HorizontalRule()

	picker := page.Picker(page.Pager.PageQueryParam)
	if picker != "" {
		out += md.Paragraph(picker)
	}

	return out
}
