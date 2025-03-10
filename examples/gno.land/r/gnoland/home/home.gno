package home

import (
	"gno.land/p/demo/ownable"
	"gno.land/p/moul/dynreplacer"
	blog "gno.land/r/gnoland/blog"
	events "gno.land/r/gnoland/events"
)

// XXX: p/demo/ui API is crappy, we need to make it more idiomatic
// XXX: use an updatable block system to update content from a DAO
// XXX: var blocks avl.Tree

var (
	override string
	Admin    = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

func Render(_ string) string {
	r := dynreplacer.New()
	r.RegisterCallback(":latest-blogposts:", func() string {
		return blog.RenderLastPostsWidget(4)
	})
	r.RegisterCallback(":upcoming-events:", func() string {
		out, _ := events.RenderEventWidget(events.MaxWidgetSize)
		return out
	})

	template := `# Welcome to gno.land

## We're building a user-owned internet where governance is fair, development is open, and innovation is rewarded.

[/about](/about) &nbsp;&nbsp;&nbsp; [/github](/github) &nbsp;&nbsp;&nbsp; [/try gno](/try-gno) 

We’re building gno.land, set to be the leading open-source smart contract platform for the Internet of the future. At its core, gno.land is powered by Gno, an interpreted and fully deterministic variation of the Go programming language, designed for succinct and composable smart contracts.

The gno.land blockchain is the first experiment showcasing what’s possible with Gno, providing a secure, auditable, and reusable architecture for building and iterating on decentralized applications. With transparent and timeless code, gno.land serves as the “GitHub” of the Gno ecosystem, enabling anyone to inspect, reuse, and contribute to the development of Gno-based dApps. Intuitive and easy to use, gno.land makes anti-censorship-resistant web3 applications the standard for the future. Join us in building an open-source-first, fairer, and freer Internet for everyone

- Reusability - smart contracts built with Gno are designed to encourage reusability, composability, and forkability when needed. 
- Open and transparent - all smart contracts and applications are fully open-source and auditable, ensuring verifiability. 
- Decentralized Governance – governance is maintained by trusted contributors and the community. 
- Developer-Friendly – Go developers can easily transition to Gno and start building without learning a new, restrictive language.


---

Why Gno?

## A new decentralized operating system (dOS) to power the Internet of Everyone

We are building an operating system and network for everyone who wants to contribute and change the way the human component of the web currently operates. Find out more how gno.land is making this happen. 

<gno-columns>

####

![Social coordination & community](#)

### Social coordination & community

Censorship-resistant, community-organized social impact platforms that provide coordination, conversation, conflict resolution, and a builder-owned governance sstructure. 


####

![Governance & organization](#)

### Governance & Organization

Redefining governance frameworks, negotiation systems, and decentralized autonomous organizations (DAOs) to prioritize a common ecosystem alignment, active participation and ownership, and application specific implementations in lieu of a solely capital-based control system.

####

![New-Economic-Model](#)

### New Economic Model

A contribution-driven economic model that incentivizes creators, innovators, moderators, and maintainers to ensure a sustainable, builder-aligned governance and organizational framework for an alternative model to the current speculation or advertising based internet. 

####

![The next killer app](#)

### The next killer app

A reusable technical architecture for everyone to build the killer applications of the future that attract the next wave of builders and users to the Internet of Everyone.   


</gno-columns>

TODO: Add 'Beta Mainnet Coming Soon!' section

Stay tuned for the first phase of gno.land's launch, the beta mainnet release, and airdrop distribution. [Road to Mainnet and Beyond](https://gno.land/r/gnoland/blog:p/road-to-mainnet)

TODO: Include the gno.land roadmap dynamic infographic

**A moment in time (May 20th, 2020)*: Cosmos snapshot 
**Milestone complete (July 2024)**: A multi-node, permanent and stable testnet with 1st implementation of the GovDAO
**Milestone complete (November 2024)**: Extended validator set, testnet improvements and GovDAO v2
**Milestone 'Beta Mainnet Release' (March 2025)**: GovDAO v3, r/Boards, Token distribution, Security audit complete
**Milestone 'Build the Future' (2025)**: Bridging Cosmos and gno.land, the foundations of a network of everyone
**Milestone 'Mainnet' (2025)**: A complete mainnet with token transfers and Cosmos bridge with ICS and IBC

---

Ecosystem

## Discover the gno.land universe

Explore gno.land's expanding ecosystem of applications, wallets, and explorers.

[/the universe ](#)


<gno-columns>

####

![Wallet](#)

### Wallets

Execute transactions on gno.land using wallets, or a CLI keychain and client tool.

- [Adena Wallet](https://www.adena.app/)
- [gnokey](https://docs.gno.land/gno-tooling/cli/gnokey/)

####

![Explorers](#)

### Explorers

Monitor and track the status of on-chain transactions in real time. 

- [gnoscan](https://gnoscan.io/)


####

![Defi](#)

### Defi and Games

Try out the DeFi platforms and gaming applications on gno.land. 

- [gnoswap - a gno.land DEX](http://beta.gnoswap.io/)
- [gnochess](TODO)
- [Flippando](https://flippando.xyz/#)
- [Shifumi] (TODO)
- [Conway's Game of Life](TODO)

####

![Social dApps](#)

### Social and Community dApps

Register your gno.land username and be a part of the conversation happening on gno.land social network, r/Boards.  

- [r/users - gno.land's user name registry](https://gno.land/r/gnoland/users/v1)
- [r/Boards - gno.land's social network](TODO)


</gno-columns>

## A community of gnomes

gno.land is a contribution-driven network and community where gnomes, its dedicated contributors, serve as the builders, guardians, and pioneers of a future Internet of Everyone.

<gno-columns>
##

![Blog](#)

### Our Blog 

:latest-blogposts:

##

![Events](#)

### Past Events

:upcoming-events:

</gno-columns>

---

Builders

## Build on gno.land

gno.land builders are at the center of creating a user-owned network where applications remain open and reusable, governance aligns with contributors, and innovation extends beyond financial speculation.

[/read the documentation](docs.gno.land)

<gno-columns>

### Builder Portal

The Builder Portal is a collection of educational resources and tools to help you learn, experiment, and develop your Gno skills. 

[TODO](/)

### Bounties

The Gno bounty program is a good way to find interesting challenges in Gno, and get rewarded for helping us advance the project. 

[/read more](https://github.com/gnolang/bounties)

### gno.land grants

gno.land grants program provides financial support to developers interested in exploring new blockchain technology, and/or are already actively contributing to gno.land.

[/read more](https://github.com/gnolang/grants)

</gno-columns>

---

<gno-columns>

### General

- [About](#)
- [Events](#)
- [Blog](#)
- [Carreer](#)

### Advanced

- [Builder Guides](#)
- [Tokenomics](#)
- [Demo list](#)
- [Docs](#)

### Explore

- [Faucet](#)
- [Ecosystem](#)
- [Partners](#)
- [Gnops](#)

### Social

- [GitHub](#)
- [Discord](#)
- [YouTube](#)
- [Twitter](#)

</gno-columns>
`

	if override != "" {
		template = override
	}
	result := r.Replace(template)
	return result
}

func AdminSetOverride(content string) {
	Admin.AssertCallerIsOwner()
	override = content
}
