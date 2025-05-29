package awesomegno

import (
	"gno.land/p/moul/md"
)

func GetLatestItemsWidget() string {
	crossing()

	latestIDs := GetLatestItems()
	if len(latestIDs) == 0 {
		return md.Italic("No items available")
	}

	var itemsList []string
	for _, itemID := range latestIDs {
		if itemVal, found := items.Get(itemID.String()); found {
			item := itemVal.(*Item)
			itemsList = append(itemsList, formatItemListItem(item))
		}
	}

	if len(itemsList) == 0 {
		return md.Italic("No items available")
	}

	return md.BulletList(itemsList)
}

func GetLatestCategoriesWidget() string {
	crossing()

	latestIDs := GetLatestCategories()
	if len(latestIDs) == 0 {
		return md.Italic("No categories available")
	}

	var categoriesList []string
	for _, categoryID := range latestIDs {
		if categoryVal, found := categories.Get(categoryID.String()); found {
			category := categoryVal.(*Category)
			categoriesList = append(categoriesList, formatCategoryListItem(category))
		}
	}

	if len(categoriesList) == 0 {
		return md.Italic("No categories available")
	}

	return md.BulletList(categoriesList)
}

func GetLatestItemsWithTitle(title string) string {
	crossing()

	out := md.H3(title)
	out += GetLatestItemsWidget()
	return out
}

func GetLatestCategoriesWithTitle(title string) string {
	crossing()

	out := md.H3(title)
	out += GetLatestCategoriesWidget()
	return out
}
