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

## We’re building a decentralized operating system where governance is fair, development is open, and innovation is rewarded.

[/about](/about) &nbsp;&nbsp;&nbsp; [/github](/github) &nbsp;&nbsp;&nbsp; [/try gno](/try-gno) 


**/about /github /try gno**

gno.land is a smart contract platform built for Gno, a deterministic implementation of Go, designed for secure, auditable, and reusable decentralized applications (dApps). Unlike historical blockchain development languages, Gno allows Go developers to easily transfer their skills without needing to learn a completely new language. With type safety and composability, gno.land makes building scalable dApps secure, intuitive and reusable.

At its core, gno.land is the ‘GitHub’ of the Gno ecosystem, where transparent, open-source smart contracts are built on top of a native censorship-resistant decentralized operating system (dOS) starting with the Gno Virtual Machine (GnoVM).

## Key Features of Gno

- Open & Transparent: All applications are open-source and fully auditable.
- Composable: Smart contracts are modular, reusable, and customizable to ease development and extend the ecosystem’s capabilities quickly.
- Censorship-Resistant: Enables platforms with fair moderation and freedom of communication without a central authority.
- Developer-Friendly: Intuitive design that reduces onboarding timelines and enables faster iteration.
- Go-Based & Type Safe: Go developers can transition seamlessly to Gno thanks to the well-established Go language, while benefiting from type safety for security.


---

Use Cases

## Why Gno and gno.land?

gno.land’s smart contract platform is a first step in an experiment to build a complete dOS to bypass the centralized and controlled restrictions of the current Web, and power new social platforms, governance frameworks, and economic models that favor contribution-based incentives.

<gno-columns>

####

![The next killer app](#)

### A Platform Designed for Next Generation Apps

A reusable technical architecture for everyone to build the next generation applications of the future that attract the next wave of builders and users.

####

![New-Economic-Model](#)

### New Economic Model

A builder-aligned economy that incentivizes developers, content creators, and maintainers with sustainable rewards.

####

![Decentralized Platform Governance](#)

### Decentralized Governance: GovDAO

A governance model that prioritizes contribution-based participation over capital-weighted control.

####

![Social coordination & community](#)

### Social coordination & community

A decentralized space for uncensored discourse, fair moderation, and community-driven governance with forkability to allow for conflict resolution.




</gno-columns>

---
Beta Mainnet

## Beta Mainnet Coming Soon!

Stay tuned for the first phase of gno.land's launch, the beta mainnet release.  
- [Road to Mainnet and Beyond](https://gno.land/r/gnoland/blog:p/road-to-mainnet)

---

Ecosystem

## Discover the gno.land universe

Explore gno.land's expanding ecosystem of applications, wallets, and explorers.

[/the ecosystem ](https://gno.land/ecosystem)


<gno-columns>

####

![Wallet](#)

### Wallets and Faucet

Execute transactions on gno.land using wallets, or a CLI keychain and client tool.

- [Adena Wallet](https://www.adena.app/)
- [gnokey](https://docs.gno.land/gno-tooling/cli/gnokey/)
- [gno.land Faucet Hub](https://faucet.gno.land/)

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

####

![Social dApps](#)

### Social and Community dApps

Register your gno.land username and be a part of the conversation happening on gno.land social network, r/Boards.  

- [gno.land's user name registry](https://gno.land/r/gnoland/users/v1)
- [Boards - gno.land's social network](coming soon!)
- [Hall of Fame](https://gno.land/r/leon/hof) 
- [Gno.me - gno.land's educational and community development platform](Coming soon!)


</gno-columns>

## A community of gnomes

gno.land is a contribution-driven network and community where dedicated contributors, also known as gnomes, serve as the builders, guardians, and pioneers of gno.land’s future.

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

gno.land builders are at the center of creating the applications and infrastructure for the decentralized operating system needed to power an Internet that is 100% decentralized.

[/read the documentation](docs.gno.land) [Read about the Gno language](https://gno.land/gnolang)

<gno-columns>

### Builders

- [Discover Gno Examples](https://gno.land/r/docs)
- [Gno Playground: Code to Gno](https://play.gno.land/)
- [Gno Studio Connect: A Realm interaction app](https://gno.studio/connect)
- [Gnodev: Efficient local development for Gno](https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev)
- [gno.land Faucet Hub](https://faucet.gno.land/)
- [Discover Demo Packages](https://github.com/gnolang/gno/tree/master/examples)

### Bounties

The Gno bounty program is a good way to find interesting challenges in Gno, and get rewarded for helping us advance the project. 

[/read more](https://github.com/gnolang/bounties)

### gno.land grants

Are you a builder, tinkerer, or researcher? If you’re looking to create awesome dApps, tooling, infrastructure, or smart contract libraries on gno.land, you can apply for a grant. gno.land grants program provides financial support to developers interested in exploring new blockchain technology, and/or are already actively contributing to gno.land.

[/read more](https://github.com/gnolang/grants)

</gno-columns>

---

<gno-columns>

### General

- [About](#)
- [Events](#)
- [Blog](#)
- [Career](#)

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
