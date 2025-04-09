package awesomegno

import (
	"strings"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

const (
	dappsPath      = "dapps"
	categoriesPath = "categories"
	pageSize       = 10
)

func Render(path string) string {
	var out string

	switch {
	case path == "":
		out = renderHome()
	case strings.HasPrefix(path, dappsPath):
		out = renderDapps(path)
	case strings.HasPrefix(path, categoriesPath):
		out = renderCategories(path)
	case strings.HasPrefix(path, "d/"):
		id := strings.TrimPrefix(path, "d/")
		out = renderSingleDapp(id)
	case strings.HasPrefix(path, "c/"):
		id := strings.TrimPrefix(path, "c/")
		out = renderSingleCategory(id)
	}

	return out
}

func renderHome() string {
	out := md.H1("AwesomeGno")
	out += md.Paragraph("Welcome to AwesomeGno - A curated list of awesome dapps built on Gno.land. Here you can discover and explore the growing ecosystem of decentralized applications.")

	out += md.H2("Quick Stats")
	stats := []string{
		ufmt.Sprintf("Total Dapps: %d", dapps.Size()),
		ufmt.Sprintf("Categories: %d", categories.Size()),
	}
	out += md.BulletList(stats)

	out += md.H2("Navigation")
	links := []string{
		md.Link("Browse All Dapps", "/r/demo/awesomegno:dapps"),
		md.Link("Browse Categories", "/r/demo/awesomegno:categories"),
	}
	out += md.BulletList(links)

	out += md.H2("About")
	out += md.Paragraph("AwesomeGno is a community-driven project that aims to showcase and promote the best dapps built on Gno.land. Each dapp is carefully categorized and reviewed to ensure quality and usefulness.")

	return out
}

func renderDapps(path string) string {
	out := md.H1("Dapps")
	out += md.Paragraph("Browse through all the awesome dapps built on Gno.land.")

	addDappLink := txlink.NewLink("AddDapp").
		AddArgs(
			"title", "",
			"description", "",
			"url", "",
			"author", "",
		).URL()
	out += md.Paragraph(md.Link("➕ Add New Dapp", addDappLink))

	var dappItems []string
	dappPager := pager.NewPager(dapps, pageSize, false)
	page := dappPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		dapp := item.Value.(*Dapp)
		dappItems = append(dappItems, formatDappListItem(dapp))
	}

	out += renderItemsList(dappItems, page, "No dapps registered yet")
	return out
}

func renderCategories(path string) string {
	out := md.H1("Categories")
	out += md.Paragraph("Browse dapps by category.")

	addCategoryLink := txlink.NewLink("AddCategory").
		AddArgs(
			"name", "",
			"description", "",
		).URL()
	out += md.Paragraph(md.Link("➕ Add New Category", addCategoryLink))

	var categoryItems []string
	categoryPager := pager.NewPager(categories, pageSize, false)
	page := categoryPager.MustGetPageByPath(path)

	for _, item := range page.Items {
		category := item.Value.(*Category)
		categoryItems = append(categoryItems, formatCategoryListItem(category))
	}

	out += renderItemsList(categoryItems, page, "No categories available")
	return out
}

func renderSingleDapp(id string) string {
	v, found := dapps.Get(id)
	if !found {
		return md.Italic("Dapp not found")
	}

	dapp := v.(*Dapp)
	out := md.H1(dapp.Title)
	out += md.Paragraph(dapp.Description)

	dappAuthors := []string{}
	for _, author := range dapp.Authors {
		dappAuthors = append(dappAuthors, author.String())
	}

	metadata := []string{
		ufmt.Sprintf("Authors: %s", strings.Join(dappAuthors, ", ")),
		ufmt.Sprintf("Created: %s", dapp.CreatedAt.Format("2006-01-02")),
		ufmt.Sprintf("URL: %s", md.Link("Visit Dapp", dapp.URL)),
	}
	out += md.BulletList(metadata)

	out += md.H2("Categories")
	var categories []string
	dapp.categories.Iterate("", "", func(key string, value interface{}) bool {
		cat := value.(*Category)
		categories = append(categories, md.Link(cat.Name, "/r/demo/awesomegno:c/"+cat.ID.String()))
		return false
	})

	if len(categories) > 0 {
		out += md.BulletList(categories)
	} else {
		out += md.Italic("No categories assigned")
	}

	out += md.H2("Management")
	out += md.Paragraph("Use these actions to manage the dapp:")

	updateLink := txlink.NewLink("UpdateDapp").
		AddArgs(
			"dappID", dapp.ID.String(),
			"title", "",
			"description", "",
			"url", "",
		).URL()

	deleteLink := txlink.NewLink("DeleteDapp").
		AddArgs("dappID", dapp.ID.String()).URL()

	actions := []string{
		md.Link("📝 Update Dapp", updateLink),
		md.Link("🗑️ Delete Dapp", deleteLink),
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
	out := md.H1(category.Name)
	out += md.Paragraph(category.Description)
	out += md.Paragraph(ufmt.Sprintf("Created: %s", category.CreatedAt.Format("2006-01-02")))

	out += md.H2("Dapps in this category")
	var dappsList []string
	category.dapps.Iterate("", "", func(key string, value interface{}) bool {
		dapp := value.(*Dapp)
		dappsList = append(dappsList, formatDappListItem(dapp))
		return false
	})

	if len(dappsList) > 0 {
		out += md.BulletList(dappsList)
	} else {
		out += md.Italic("No dapps in this category yet")
	}

	out += md.H2("Management")
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

func formatDappListItem(dapp *Dapp) string {
	return ufmt.Sprintf("%s - %s",
		md.Link(dapp.Title, "/r/demo/awesomegno:d/"+dapp.ID.String()),
		md.Italic(strings.Split(dapp.Description, ".")[0]))
}

func formatCategoryListItem(category *Category) string {
	return ufmt.Sprintf("%s (%d dapps) - %s",
		md.Link(category.Name, "/r/demo/awesomegno:c/"+category.ID.String()),
		category.dapps.Size(),
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
