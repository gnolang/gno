package awesomegno

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/moul/md"
	"gno.land/r/sys/users"
)

func Render(path string) string {
	queryParams := ""
	if idx := strings.Index(path, "?"); idx != -1 {
		queryParams = path[idx:]
		path = path[:idx]
	}

	if path == "" {
		return renderHomePage(queryParams)
	}

	if path == "categories" {
		return renderCategoriesPage(queryParams)
	}

	if path == "contribute" {
		return renderContributePage()
	}

	if strings.HasPrefix(path, "dapp/") {
		idStr := strings.TrimPrefix(path, "dapp/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return "Invalid dApp ID"
		}
		return renderDappPage(id)
	}

	if strings.HasPrefix(path, "categories/") {
		categoryName := strings.TrimPrefix(path, "categories/")
		return renderCategoryPage(categoryName, queryParams)
	}

	if path == "proposals" {
		return renderProposalsPage(queryParams)
	}

	if strings.HasPrefix(path, "proposals/") {
		idStr := strings.TrimPrefix(path, "proposals/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return "Invalid proposal ID"
		}
		return renderProposalPage(id)
	}

	return "Page not found"
}

func renderHomePage(queryParams string) string {
	out := md.H1("Awesome Gno")
	out += "A curated list of awesome Gno resources, libraries, tools and more.\n\n"

	out += md.Link("Browse Categories", "/r/demo/awesomegno:categories") + "\n\n"

	out += md.H2("All Resources")

	dappPager := pager.NewPager(dapps, 5, true)
	dappPage := dappPager.MustGetPageByPath(queryParams)

	if len(dappPage.Items) == 0 {
		out += "No dApps available yet.\n\n"
	} else {
		for _, item := range dappPage.Items {
			dapp := item.Value.(Dapp)
			id, _ := strconv.ParseUint(item.Key, 10, 64)

			out += md.H3(md.Link(dapp.Title, "/r/demo/awesomegno:dapp/"+strconv.FormatUint(id, 10)))
			out += dapp.Description + "\n\n"

			if len(dapp.Categories) > 0 {
				categoryLinks := []string{}
				for _, cat := range dapp.Categories {
					categoryLinks = append(categoryLinks, md.Link(cat, "/r/demo/awesomegno:category/"+cat))
				}
				out += md.Bold("Categories:") + " " + strings.Join(categoryLinks, ", ") + "\n\n"
			}

			out += md.Bold("Author:") + " " + getUserDisplayName(dapp.Author) + "\n\n"
			out += md.Bold("URL:") + " " + md.Link(dapp.URL, dapp.URL) + "\n\n"

			out += md.HorizontalRule()
		}

		dappPagination := dappPage.Picker(dappPage.Pager.PageQueryParam)
		if dappPagination != "" {
			out += dappPagination + "\n\n"
		}
	}

	out += md.H2("More")
	out += md.Link("View Pending Proposals", "/r/demo/awesomegno:proposals") + " | "
	out += md.Link("How to Contribute", "/r/demo/awesomegno:contribute") + "\n\n"

	return out
}

func renderDappPage(id uint64) string {
	dapp, exists := GetDapp(id)
	if !exists {
		return "dApp not found"
	}

	out := md.H1(dapp.Title)
	out += dapp.Description + "\n\n"

	if len(dapp.Categories) > 0 {
		categoryLinks := []string{}
		for _, cat := range dapp.Categories {
			categoryLinks = append(categoryLinks, md.Link(cat, "/r/demo/awesomegno:category/"+cat))
		}
		out += md.Bold("Categories:") + " " + strings.Join(categoryLinks, ", ") + "\n\n"
	}

	out += md.Bold("Author:") + " " + getUserDisplayName(dapp.Author) + "\n\n"
	out += md.Bold("URL:") + " " + md.Link(dapp.URL, dapp.URL) + "\n\n"
	out += md.Bold("Added:") + " " + dapp.CreatedAt.Format(time.RFC3339) + "\n\n"

	return out
}

func renderProposalsPage(queryParams string) string {
	out := md.H1("Pending Proposals")

	pendingProposalsTree := avl.NewTree()
	proposals.Iterate("", "", func(key string, value interface{}) bool {
		proposal := value.(Proposal)
		if proposal.Status == StatusPending {
			pendingProposalsTree.Set(key, proposal)
		}
		return false
	})

	proposalPager := pager.NewPager(pendingProposalsTree, 5, true)
	page := proposalPager.MustGetPageByPath(queryParams)

	if len(page.Items) == 0 {
		out += "No pending proposals.\n\n"
	} else {
		for _, item := range page.Items {
			proposal := item.Value.(Proposal)

			out += md.H2(md.Link(proposal.Dapp.Title, "/r/demo/awesomegno:proposals/"+strconv.FormatUint(proposal.ID, 10)))
			out += md.Bold("Proposer:") + " " + getUserDisplayName(proposal.Proposer) + "\n\n"
			out += md.Bold("Submitted:") + " " + proposal.CreatedAt.Format(time.RFC3339) + "\n\n"
			out += md.HorizontalRule()
		}

		pagination := page.Picker(page.Pager.PageQueryParam)
		if pagination != "" {
			out += pagination + "\n\n"
		}
	}

	return out
}

func renderProposalPage(id uint64) string {
	proposal, exists := GetProposal(id)
	if !exists {
		return "Proposal not found"
	}

	out := md.H1("Proposal: " + proposal.Dapp.Title)
	out += md.Bold("Status:") + " " + proposal.Status + "\n\n"
	out += md.Bold("Proposer:") + " " + getUserDisplayName(proposal.Proposer) + "\n\n"
	out += md.Bold("Submitted:") + " " + proposal.CreatedAt.Format(time.RFC3339) + "\n\n"

	if proposal.Status == StatusRejected && proposal.Reason != "" {
		out += md.Bold("Rejection Reason:") + " " + proposal.Reason + "\n\n"
	}

	out += md.H2("dApp Details")
	out += md.Bold("Title:") + " " + proposal.Dapp.Title + "\n\n"
	out += md.Bold("Description:") + " " + proposal.Dapp.Description + "\n\n"

	if len(proposal.Dapp.Categories) > 0 {
		categoryLinks := []string{}
		for _, cat := range proposal.Dapp.Categories {
			categoryLinks = append(categoryLinks, md.Link(cat, "/r/demo/awesomegno:category/"+cat))
		}
		out += md.Bold("Categories:") + " " + strings.Join(categoryLinks, ", ") + "\n\n"
	}

	out += md.Bold("URL:") + " " + md.Link(proposal.Dapp.URL, proposal.Dapp.URL) + "\n\n"

	return out
}

func renderCategoryPage(categoryName string, queryParams string) string {
	category, err := GetCategory(categoryName)
	if err != nil {
		return "Category not found: " + categoryName
	}

	out := md.H1("Category: " + category.Name)

	if category.Description != "" {
		out += category.Description + "\n\n"
	}

	dappsInCategoryTree := avl.NewTree()
	for _, dappID := range category.DappIDs {
		value, found := dapps.Get(strconv.FormatUint(dappID, 10))
		if found {
			dapp := value.(Dapp)
			if dapp.Approved {
				dappsInCategoryTree.Set(strconv.FormatUint(dappID, 10), dapp)
			}
		}
	}

	dappPager := pager.NewPager(dappsInCategoryTree, 5, true)
	page := dappPager.MustGetPageByPath(queryParams)

	if len(page.Items) == 0 {
		out += "No resources available in this category yet.\n\n"
	} else {
		for _, item := range page.Items {
			dapp := item.Value.(Dapp)
			id, _ := strconv.ParseUint(item.Key, 10, 64)

			out += md.H3(md.Link(dapp.Title, "/r/demo/awesomegno:dapp/"+strconv.FormatUint(id, 10)))
			out += dapp.Description + "\n\n"

			if len(dapp.Categories) > 1 {
				categoryLinks := []string{}
				for _, cat := range dapp.Categories {
					if cat != categoryName {
						categoryLinks = append(categoryLinks, md.Link(cat, "/r/demo/awesomegno:category/"+cat))
					}
				}
				if len(categoryLinks) > 0 {
					out += md.Bold("Also in:") + " " + strings.Join(categoryLinks, ", ") + "\n\n"
				}
			}

			out += md.Bold("Author:") + " " + getUserDisplayName(dapp.Author) + "\n\n"
			out += md.Bold("URL:") + " " + md.Link(dapp.URL, dapp.URL) + "\n\n"

			out += md.HorizontalRule()
		}

		pagination := page.Picker(page.Pager.PageQueryParam)
		if pagination != "" {
			out += pagination + "\n\n"
		}
	}

	return out
}

func renderCategoriesPage(queryParams string) string {
	out := md.H1("Categories")
	out += "Browse resources by category:\n\n"

	categoriesTree := avl.NewTree()
	categories.Iterate("", "", func(key string, value interface{}) bool {
		categoriesTree.Set(key, value)
		return false
	})

	categoryPager := pager.NewPager(categoriesTree, 30, true)
	categoryPage := categoryPager.MustGetPageByPath(queryParams)

	if len(categoryPage.Items) == 0 {
		out += "No categories available yet.\n\n"
	} else {
		categoryItems := []string{}
		for _, item := range categoryPage.Items {
			category := item.Value.(Category)
			dappCount := len(category.DappIDs)
			categoryItems = append(categoryItems, md.Link(category.Name+" ("+strconv.Itoa(dappCount)+")", "/r/demo/awesomegno:categories/"+category.Name))
		}

		out += md.BulletList(categoryItems)

		categoryPagination := categoryPage.Picker(categoryPage.Pager.PageQueryParam)
		if categoryPagination != "" {
			out += "Page: " + categoryPagination + "\n\n"
		}
	}

	return out
}

func renderContributePage() string {
	out := md.H1("Contributing to Awesome Gno")

	out += md.H2("How to Submit a Resource")
	out += "To submit a new resource to Awesome Gno, call the `ProposeNewDapp()` function with the following parameters:\n\n"
	out += md.BulletList([]string{
		"title: The name of your resource",
		"description: A brief description of what it does",
		"url: The URL where people can find your resource",
		"categories: A comma-separated list of categories (e.g., \"Tools,DeFi,Games\")",
	})

	out += "\nExample using a realm import:\n"
	out += md.LanguageCodeBlock("go", `// Submit a proposal for a new dApp
awesomegno.ProposeNewDapp(
    "My Awesome Tool", 
    "A tool that helps developers build on Gno", 
    "https://example.com/my-tool",
    "Tools,Developer,Utility"
)`)

	out += "\nOr alternatively using `gnokey`:\n"
	out += md.LanguageCodeBlock("bash", `gnokey maketx call \
-pkgpath "gno.land/r/demo/awesomegno" \
-func "ProposeNewDapp" \
-args "My Awesome Tool" \
-args "A tool that helps developers build on gno" \
-args "https://example.com/my-tool" \
-args "Tools,Developer,Utility" \
-gas-fee 1000000ugnot \
-gas-wanted 5000000 \
-broadcast \
-chainid "dev" \
-remote "tcp://127.0.0.1:26657" \
YOUR-ADDRESS`)

	out += "\n"

	out += md.H2("Contributing Guidelines")
	out += "Your contributions are always welcome!\n\n"

	out += md.H3("Guidelines")
	out += md.BulletList([]string{
		"Submit one resource at a time using the ProposeNewDapp() function",
		"Provide a clear and concise title for your resource",
		"Explain why your resource is valuable to the Gno ecosystem",
		"Include a detailed description that ends with a period",
		"Keep descriptions concise and informative",
		"Select appropriate categories for your resource",
		"Check if your resource already exists before submitting",
		"Don't mention Gno in the description as it's implied",
		"Ensure your submission has correct spelling and grammar",
		"Provide a valid and working URL",
	})

	out += "\n\nFor more information about the Awesome Gno project, visit the " + md.Link("Awesome Gno GitHub repository", "https://github.com/gnoverse/awesome-gno") + ".\n\n"

	return out
}

func getUserDisplayName(addr std.Address) string {
	userData := users.ResolveAddress(addr)
	if userData != nil {
		return "@" + userData.Name()
	}
	return addr.String()
}
