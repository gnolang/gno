package pages

func init() {
	path := "contribute"
	title := "Contributor Ecosystem: Call for Contributions"
	body := `

Gno.land puts at the center of its identity the contributors that help to create and shape the project into what it is; incentivizing those who contribute the most and help advance its vision. Eventually, contributions will be incentivized directly on-chain; in the meantime, this page serves to illustrate our current off-chain initiatives.

Gno.land is still in full-steam development. For now, we're looking for the earliest of adopters; curious to explore a new way to build smart contracts and eager to make an impact. Joining Gno.land's development now means you can help to shape the base of its development ecosystem, which will pave the way for the next generation of blockchain programming.

As an open-source project, we welcome all contributions. On this page you can find some pointers on where to get started; as well as some incentives for the most valuable and important contributions.

## Where to get started

If you are interested in contributing to Gno.land, you can jump in on our [GitHub monorepo](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md) - where most development happens.

A good place to start are the issues tagged ["good first issue"](https://github.com/gnolang/gno/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22). They should allow you to make some impact on the Gno repository while you're still exploring the details of how everything works.

## Gno Bounties

> **Note: The Gno Bounty Program is currently paused until further notice.**

Two reasons prompted this pause:

- **Focus on Mainnet.** The team is fully focused on the Gno.land Beta Mainnet launch — the highest priority — as well as the critical post-beta work required to deliver a complete Mainnet.
- **Decline in submission quality.** The bounty program was created to encourage innovative, creative, and technical contributions that meaningfully advance Gno.land. However, with the rise of generative AI, the overall quality of submissions has declined significantly. A large portion of entries are clearly AI-generated; others, though less obviously so, are often inaccurate upon review due to overreliance on AI. Reviewing and validating these submissions has become increasingly time-consuming, diverting focus from development work.

Community contributions remain strongly encouraged, particularly:

- Tooling and infrastructure projects
- Application-focused development
- Longer-term, high-quality initiatives

These types of contributions are better aligned with Gno's goals than short-term bounties, and are very welcome provided they follow the [contributing guidelines](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md). Sustained, high-quality contributions will be rewarded over time via the Gno.land GovDAO.

## Gno.land Grants

The Gno.land grants program is to encourage and support the growth of the Gno.land contributor community, and build out the usability of the platform and smart contract library. The program provides financial resources to contributors to explore the Gno tech stack, and build dApps, tooling, infrastructure, products, and smart contract libraries in Gno.land.

For more details on Gno.land grants, suggested topics, and how to apply, visit our grants [repository](https://github.com/gnolang/grants).

## Join Game of Realms

Game of Realms is the overarching contributor network of gnomes, currently running off-chain, and will eventually transition on-chain. At this stage, a Game of Realms contribution is comprised of high-impact contributions identified as ['notable contributions'](https://github.com/gnolang/game-of-realms/tree/main/contributors).

These contributions are not linked to immediate financial rewards, but are notable in nature, in the sense they are a challenge, make a significant addition to the project, and require persistence, with minimal feedback loops from the core team.

The selection of a notable contribution or the sum of contributions that equal 'notable' is based on the impact it has on the development of the project. For now, it is focused on code contributions, and will evolve over time. The Gno development teams will initially qualify and evaluate notable contributions, and vote off-chain on adding them to the 'notable contributions' folder on GitHub.

You can always contribute to the project, and all contributions will be noticed. Contributing now is a way to build your personal contributor profile in Gno.land early on in the ecosystem, and signal your commitment to the project, the community, and its future.

There are a variety of ways to make your contributions count:

- Core code contributions
- Realm and pure package development
- Validator tooling
- Developer tooling
- Tutorials and documentation

To start, we recommend you create a PR in the Game of Realms [repository](https://github.com/gnolang/game-of-realms) to create your profile page for all your contributions.`

	if err := b.NewPost("", path, title, body, "2024-09-05T00:00:00Z", nil, nil); err != nil {
		panic(err)
	}
}
