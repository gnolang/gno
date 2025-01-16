package present

func init() {
	path := "miami23"
	title := "Portal Loop Demo (Miami 2023)"
	body := `
Rendered by Gno.

[Source (WIP)](https://github.com/gnolang/gno/pull/1176)

## Portal Loop

- DONE: Dynamic homepage, key pages, aliases, and redirects.
- TODO: Deploy with history, complete worxdao v0.
- Will replace the static gno.land site.
- Enhances local development.

[GitHub Issue](https://github.com/gnolang/gno/issues/1108)

## Roadmap

- Crafting the roadmap this week, open to collaboration.
- Combining onchain (portal loop) and offchain (GitHub).
- Next week: Unveiling the official v0 roadmap.

## Teams, DAOs, Projects

- Developing worxDAO contracts for directories of projects and teams.
- GitHub teams and projects align with this structure.
- CODEOWNER file updates coming.
- Initial teams announced next week.

## Tech Team Retreat Plan

- Continue Portal Loop.
- Consider dApp development.
- Explore new topics [here](https://github.com/orgs/gnolang/projects/15/).
- Engage in workshops.
- Connect and have fun with colleagues.
`
	_ = b.NewPost(adminAddr, path, title, body, "2023-10-15T13:17:24Z", []string{"moul"}, []string{"demo", "portal-loop", "miami"})
}
