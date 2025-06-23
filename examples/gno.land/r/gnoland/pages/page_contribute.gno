package gnopages

func init() {
	path := "contribute"
	title := "Contributor Ecosystem: Call for Contributions"
	body := `

Gno.land puts at the center of its identity the contributors that help to create and shape the project into what it is; incentivizing those who contribute the most and help advance its vision. Eventually, contributions will be incentivized directly on-chain; in the meantime, this page serves to illustrate our current off-chain initiatives.

Gno.land is still in full-steam development. For now, we're looking for the earliest of adopters; curious to explore a new way to build smart contracts and eager to make an impact. Joining Gno.land's development now means you can help to shape the base of its development ecosystem, which will pave the way for the next generation of blockchain programming.

As an open-source project, we welcome all contributions. On this page you can find some pointers on where to get started; as well as some incentives for the most valuable and important contributions.

## Where to get started

If you are interested in contributing to Gno.land, you can jump on in on our [GitHub monorepo](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md) - where most development happens.

A good place where to start are the issues tagged ["good first issue"](https://github.com/gnolang/gno/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22). They should allow you to make some impact on the Gno repository while you're still exploring the details of how everything works.

## Gno Bounties

Additionally, you can look out to help on specific issues labeled as bounties. The Gno bounty program is a good way to find interesting challenges in Gno, and get rewarded for helping us advance the project. We will maintain open and rewardable bounties in the gnolang/gno repository, and you can search all available bounties by using the ["bounty" label](https://github.com/gnolang/gno/labels/bounty). For more detals on the categories and types of bounties, we have a [bounties README](https://github.com/gnolang/bounties).  

Recommendations on participating in the Gno.land Bounty Program:

- Identify the bounty you want to work on, and join in the discussion on the issue for anything that is unclear; or where you want to more clearly define the work to be done. At this stage, you can also start working on an initial implementation in your local enviornment.
- Once you have spent time on the code related to the bounty, we recommend submitting a 'draft' PR as soon as possible.
    - The draft PR doesn't indicate that the bounty has been assigned to you, others are free to work on other draft PRs for the bounty.
    - Make sure to reference the bounty issue on the PR description you're writing.
    - After submitting the 'draft' PR, continue working until you are ready to mark the PR as "ready for review".
    - The core team will review the bounty PR submission after the work on the bounty has been completed, and determine if it qualifies for the bounty reward.
- Ask for clarification early if an element on the requirements or implementation design is unclear.
    - Aside from publishing the PR early, keeping regular updates with the core team on the bounty issue is key to being on the right track.
    - As part of the requirements, you must adhere to the [contributing guidelines](https://github.com/gnolang/gno/blob/master/CONTRIBUTING.md); additionally, it is expected that any newly added code or functionality is properly documented, tested and covered, at least in 80% of added code.
    - You're welcome to propose additional features and work on an issue should you envision a plausible expansion or change in scope. The core team may assign a bounty to the additional work, or change the bounty with respect to the changed scope.

You may make your submission at any time; however we invite you to publish your draft PR very early in the development process. This will make your work public, so you can easily get help by the core team and other community members. Additionally, your work can be continued by other people should you get stuck or no longer be willing to work on the bounty. Likewise, you can continue the abandoned or stuck work that someone else worked on.

Don't fear your work being "stolen": if a submission is the result of multiple people's efforts, we will look to split the bounty in a way that is fair and recognises each participant in creating the final outcome. Here are some examples of how that can happen:

- If Alice does most of the work and abandons it; then Bob comes around and finishes the job, then Bob's PR will be merged. But the core team will propose a split like 70% for Alice and 30% for Bob (depending, of course, on the relative effort undertaken by both).
- If Alice makes a PR that does only 50% of the work outlined in the requirements for the original issue, she will get 50%. Someone can still come up and finish the job; and claim the remaining part.
	- If you, for instance, cannot complete the entirety of the task or, as a non-developer, can only contribute a part of the specification/implementation, you may still be awarded a bounty for your input in the contribution.
- If Alice makes a PR that aside from implementing what's required, also undertakes creating useful tools among the way, she may qualify for an "outstanding contribution"; and may be awarded up to 25% more of the original bounty's value. Or she may also ask if the team would be willing to offer a different bounty for the implementation of the tools.

Participants in the Gno.land Bounty Program must meet the legal Terms and Conditions referenced [here](https://docs.google.com/document/d/e/2PACX-1vSUF-JwIXGscrNsc5QBD7Pa6i83mXUGogAEIf1wkeb_w42UgL3Lj6jFKMlNTdwEMUnhsLkjRlhe25K4/pub).

### Bounty sizes

Each bounty is associated with a size, to which corresponds the maximum compensation for the work involved on the bounty. A bounty size may under rare occasion be revisited to a bigger or smaller size; hence why it's important to talk about your proposed solution with the core team ahead of time.

In some cases, the work associated with a bounty may be outstanding. When that happens, the core team can decide to award up to 25% of the bounty's value to the recipient.

The value of the bounty, aside from the material completion of the task, considers the involved time in managing the created pull request and iterating on feedback.


t-shirt size | expected compensation
-------------|-----------------------
[XS]         | $ 500
[S]          | $ 1000
[M]          | $ 2000
[L]          | $ 4000
[XL]         | $ 8000
_[XXL]_ \*   | $ 16000
_[3XL]_ \*   | $ 32000

[XS]:  https://github.com/gnolang/gno/labels/bounty%2FXS
[S]:   https://github.com/gnolang/gno/labels/bounty%2FS
[M]:   https://github.com/gnolang/gno/labels/bounty%2FM
[L]:   https://github.com/gnolang/gno/labels/bounty%2FL
[XL]:  https://github.com/gnolang/gno/labels/bounty%2FXL
[XXL]: https://github.com/gnolang/gno/labels/bounty%2FXXL
[3XL]: https://github.com/gnolang/gno/labels/bounty%2F3XL

\*: XXL and 3XL bounties are exceptional. Almost no issues will have these sizes; most will be broken down into smaller bounties.

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
