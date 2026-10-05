package store

import "chain/runtime"

// Realms and packages deployed before the store cannot call Register, so the
// store starts with a curated catalog of existing ones, as r/sys/namereg
// pre-registers names. Seeds go through upsert, so they obey every Register
// rule except the caller check, and are confirmed. They carry no stars:
// those come from users.
func init() {
	d := runtime.ChainDomain()
	seeds := []listing{
		{Kind: kindApp, Slug: "boards", Path: d + "/r/gnoland/boards2/v0", Title: "Boards", Tagline: "On-chain forums run by their own communities", Category: "social", Palette: 4},
		{Kind: kindApp, Slug: "blog", Path: d + "/r/gnoland/blog", Title: "Gno.land Blog", Tagline: "News, deep dives and release notes from the core team", Category: "content-media", Palette: 7},
		{Kind: kindApp, Slug: "coins", Path: d + "/r/gnoland/coins", Title: "Coin Balances", Tagline: "Check any address's balances without an explorer", Category: "utilities", Palette: 10},
		{Kind: kindApp, Slug: "token-factory", Path: d + "/r/demo/defi/grc20factory", Title: "Token Factory", Tagline: "Launch your own GRC20 token, faucet included", Category: "defi", Palette: 1},
		{Kind: kindApp, Slug: "profiles", Path: d + "/r/demo/profile", Title: "Profiles", Tagline: "A public profile for every address: bio, avatar and links", Category: "identity", Palette: 5},
		{Kind: kindService, Slug: "users", Path: d + "/r/sys/users", Title: "User Registry", Tagline: "The name registry behind every @username", Category: "identity", Palette: 6},
		{Kind: kindApp, Slug: "govdao", Path: d + "/r/gov/dao", Title: "GovDAO", Tagline: "The proposals and votes that govern gno.land", Category: "dao-governance", Palette: 3},
		{Kind: kindApp, Slug: "valopers", Path: d + "/r/gnops/valopers", Title: "Valopers", Tagline: "Profiles of the validator operators securing gno.land", Category: "infrastructure", Palette: 9},
		{Kind: kindPackage, Slug: "avl", Path: d + "/p/nt/avl/v0", Title: "AVL Tree", Tagline: "The ordered key-value tree behind most realm state", Category: "dev-tools", Palette: 6},
		{Kind: kindPackage, Slug: "ufmt", Path: d + "/p/nt/ufmt/v0", Title: "ufmt", Tagline: "Lightweight Sprintf and friends for Gno", Category: "dev-tools", Palette: 2},
		{Kind: kindPackage, Slug: "ownable", Path: d + "/p/nt/ownable/v0", Title: "Ownable", Tagline: "Single-owner access control for realms", Category: "dev-tools", Palette: 3},
		{Kind: kindPackage, Slug: "markdown", Path: d + "/p/moul/md/v0", Title: "Markdown Helpers", Tagline: "Build markdown for Render without string soup", Category: "dev-tools", Palette: 7},
	}
	h := runtime.ChainHeight()
	for _, seed := range seeds {
		l := seed          // its own object, not a pointer into the seeds array
		l.Confirmed = true // curated, like a Register
		// Before upsert, which indexes l: core keeps the core apps out of
		// New, Top and Rediscover. essentials keeps them in seed order for
		// gno.land essentials.
		l.core = true
		essentials = append(essentials, &l)
		if reason := upsert(&l, h); reason != "" {
			panic("store: seed " + l.Path + ": " + reason)
		}
	}
}
