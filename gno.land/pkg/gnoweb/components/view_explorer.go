package components

const ExplorerViewType ViewType = "explorer-view"

type ExplorerData struct {
	PkgPath        string
	Paths          []string
	PackageCounter int
	CardsListTitle string // Type of the card list (realm or pure)
	CardsList      *UserCardList
	PackageType    UserContributionType
}

// for pure and realm, we want to pluralize the title if there are multiple items
// Not universal, but works for now
func (data *ExplorerData) GetCardsListTitle() string {
	if data.PackageCounter > 1 {
		return data.CardsListTitle + "s"
	}
	return data.CardsListTitle
}

// enrichExplorerCardList enriches the explorer card list with the data from paths
func enrichExplorerCardList(data *ExplorerData) {
	switch data.PackageType {
	case UserContributionTypeRealm:
		data.CardsListTitle = "Realm"
	case UserContributionTypePure:
		data.CardsListTitle = "Pure"
	}

	items := make([]UserContribution, len(data.Paths))

	// Convert paths to UserContribution items
	for i, path := range data.Paths {
		items[i] = UserContribution{
			Title: path,
			URL:   path,
			Type:  data.PackageType, // Use the correct type
		}
	}

	// Set the title with proper pluralization
	data.CardsListTitle = data.GetCardsListTitle()

	data.CardsList = &UserCardList{
		Title:             data.CardsListTitle,
		Items:             items,
		Categories:        []UserCardListCategory{}, // No categories for explorer
		TotalCount:        data.PackageCounter,
		SearchPlaceholder: "Search " + data.CardsListTitle,
	}
}

func ExplorerView(pkgPath string, paths []string, packageCounter int, packageType UserContributionType) *View {
	viewData := ExplorerData{
		PkgPath:        pkgPath,
		Paths:          paths,
		PackageCounter: packageCounter,
		PackageType:    packageType,
	}

	enrichExplorerCardList(&viewData)

	return NewTemplateView(ExplorerViewType, "renderExplorer", viewData)
}
