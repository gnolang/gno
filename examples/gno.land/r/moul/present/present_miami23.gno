package present

import "time"

func init() {
	date, _ := time.Parse("2006-01-02", "2023-10-15")

	_ = Set(
		"miami23",                       // id
		"Portal Loop Demo (Miami 2023)", // title
		"Miami Gno Tech Retreat 2023",   // event
		"@moul",                         // author
		"2023-10-15",                    // date
		`Rendered by Gno.

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
- Connect and have fun with colleagues.`,
	)
}
