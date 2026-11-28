package impl

type FilterByTier struct {
	Tier string
}

func NewFilterByTier(tier string) FilterByTier {
	return FilterByTier{Tier: tier}
}
