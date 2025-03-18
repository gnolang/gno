package awesomegno

import (
	"strings"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/moul/md"
)

func Render(path string) string {
	if path == "" {
		return renderHomePage()
	}
	
	if strings.HasPrefix(path, "dapp/") {
		idStr := strings.TrimPrefix(path, "dapp/")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return "Invalid dApp ID"
		}
		return renderDappPage(id)
	}
	
	if strings.HasPrefix(path, "category/") {
		categoryName := strings.TrimPrefix(path, "category/")
		return renderCategoryPage(categoryName)
	}
	
	if path == "proposals" {
		return renderProposalsPage()
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

func renderHomePage() string {
	out := md.H1("Awesome Gno")
	out += "A curated list of awesome Gno resources, libraries, tools and more.\n\n"
	
	out += md.H2("Categories")
	
	categoriesTree := avl.NewTree()
	categories.Iterate("", "", func(key string, value interface{}) bool {
		categoriesTree.Set(key, value)
		return false
	})
	
	categoryPager := pager.NewPager(categoriesTree, 10, true)
	categoryPage := categoryPager.MustGetPageByPath("")
	
	if len(categoryPage.Items) == 0 {
		out += "No categories available yet.\n\n"
	} else {
		categoryLinks := []string{}
		for _, item := range categoryPage.Items {
			category := item.Value.(Category)
			dappCount := len(category.DappIDs)
			categoryLinks = append(categoryLinks, md.Link(category.Name+" ("+strconv.Itoa(dappCount)+")", "/r/demo/awesomegno:category/"+category.Name))
		}
		out += md.BulletList(categoryLinks)
		
		categoryPagination := categoryPage.Picker()
		if categoryPagination != "" {
			out += categoryPagination + "\n\n"
		}
	}
	
	out += md.H2("All Resources")
	
	dappPager := pager.NewPager(dapps, 5, true)
	dappPage := dappPager.MustGetPageByPath("")
	
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
			
			out += md.Bold("Author:") + " " + GetUserDisplayName(dapp.Author) + "\n\n"
			out += md.Bold("URL:") + " " + md.Link(dapp.URL, dapp.URL) + "\n\n"
			
			out += md.HorizontalRule()
		}
		
		dappPagination := dappPage.Picker()
		if dappPagination != "" {
			out += dappPagination + "\n\n"
		}
	}
	
	out += md.H2("Contribute")
	out += "Want to add your project? Call `ProposeNewDapp()` to submit a proposal.\n\n"
	out += md.Link("View Pending Proposals", "/r/demo/awesomegno:proposals") + "\n\n"
	
	out += md.H2("API Documentation")
	out += "This realm can be imported by other realms to display curated resources:\n\n"
	out += md.LanguageCodeBlock("go", `import "gno.land/r/demo/awesomegno"

// Get a specific dApp
dapp, exists := awesomegno.GetDapp(1)

// List all dApps
allDapps := awesomegno.ListDapps()

// Get dApps by category
categoryDapps := awesomegno.GetDappsByCategory("Tools")

// Get all categories
allCategories := awesomegno.ListCategories()`)
	
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
	
	out += md.Bold("Author:") + " " + GetUserDisplayName(dapp.Author) + "\n\n"
	out += md.Bold("URL:") + " " + md.Link(dapp.URL, dapp.URL) + "\n\n"
	out += md.Bold("Added:") + " " + dapp.CreatedAt.Format(time.RFC3339) + "\n\n"
	
	out += md.Link("Back to Home", "/r/demo/awesomegno") + "\n"
	
	return out
}

func renderProposalsPage() string {
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
	page := proposalPager.MustGetPageByPath("")
	
	if len(page.Items) == 0 {
		out += "No pending proposals.\n\n"
	} else {
		for _, item := range page.Items {
			proposal := item.Value.(Proposal)
			
			out += md.H2(md.Link(proposal.Dapp.Title, "/r/demo/awesomegno:proposals/"+strconv.FormatUint(proposal.ID, 10)))
			out += md.Bold("Proposer:") + " " + GetUserDisplayName(proposal.Proposer) + "\n\n"
			out += md.Bold("Submitted:") + " " + proposal.CreatedAt.Format(time.RFC3339) + "\n\n"
			out += md.HorizontalRule()
		}
		
		pagination := page.Picker()
		if pagination != "" {
			out += pagination + "\n\n"
		}
	}
		
	out += md.Link("Back to Home", "/r/demo/awesomegno") + "\n"
	
	return out
}

func renderProposalPage(id uint64) string {
	proposal, exists := GetProposal(id)
	if !exists {
		return "Proposal not found"
	}
	
	out := md.H1("Proposal: " + proposal.Dapp.Title)
	out += md.Bold("Status:") + " " + proposal.Status + "\n\n"
	out += md.Bold("Proposer:") + " " + GetUserDisplayName(proposal.Proposer) + "\n\n"
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
	
	out += md.Link("Back to Proposals", "/r/demo/awesomegno:proposals") + "\n"
	
	return out
}

func renderCategoryPage(categoryName string) string {
	category, found := GetCategory(categoryName)
	if !found {
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
	page := dappPager.MustGetPageByPath("")
	
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
			
			out += md.Bold("Author:") + " " + GetUserDisplayName(dapp.Author) + "\n\n"
			out += md.Bold("URL:") + " " + md.Link(dapp.URL, dapp.URL) + "\n\n"
			
			out += md.HorizontalRule()
		}
		
		pagination := page.Picker()
		if pagination != "" {
			out += pagination + "\n\n"
		}
	}
	
	out += md.Link("Back to Home", "/r/demo/awesomegno") + "\n"
	
	return out
}
