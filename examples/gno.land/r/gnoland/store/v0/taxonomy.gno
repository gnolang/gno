package store

// Kinds of listing. Register lists apps; Claim lists the others too.
const (
	kindApp     = "app"     // a realm with a Render, shown on the store front
	kindService = "service" // a realm other code calls, shown on api/v1/build
	kindPackage = "package" // a /p/ package, shown on api/v1/build
)

type category struct {
	Key   string
	Label string
}

// categories is the closed taxonomy, in display order.
var categories = []category{
	{"defi", "DeFi"},
	{"social", "Social"},
	{"games", "Games"},
	{"dao-governance", "DAO & Governance"},
	{"identity", "Identity"},
	{"nft-art", "NFT & Art"},
	{"dev-tools", "Dev tools"},
	{"infrastructure", "Infrastructure"},
	{"education", "Education"},
	{"content-media", "Content & Media"},
	{"utilities", "Utilities"},
}

func findCategory(key string) (category, bool) {
	for _, c := range categories {
		if c.Key == key {
			return c, true
		}
	}
	return category{}, false
}

func isKind(k string) bool {
	return k == kindApp || k == kindService || k == kindPackage
}
