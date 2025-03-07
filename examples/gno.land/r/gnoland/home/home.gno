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

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ultricies ultricies risus. Cras libero est, congue vitae ex nec, sodales rutrum leo. Quisque a blandit ex, at viverra ex.Donec nulla risus, ullamcorper rutrum sapien ut, varius interdum quam. Donec non sapien varius, iaculis nibh at, placerat risus. 

Proin placerat quis odio id maximus.

---

Why Gno?

## A new decentralized operating system (dOS) to power the Internet of Everyone

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ultricies ultricies risus. Cras libero est, congue vitae ex nec, sodales rutrum leo.

<gno-columns>

####

![Social coordination & community](#)

### Social coordination & community

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

####

![Governance & organization](#)

### Governance & Organization

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

####

![New-Economic-Model](#)

### New Economic Model

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

####

![The next killer app](#)

### The next killer app

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

</gno-columns>

---

Ecosystem

## Discover the gno.land universe

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ultricies ultricies risus. 

[/the universe ](#)


<gno-columns>

####

![Wallet](#)

### Wallet

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

####

![Explorers](#)

### Explorers

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

####

![Defi](#)

### Defi

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

####

![Social dApps](#)

### Social dApps

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

- [Lorem ipsum dolor](/)
- [Lorem ipsum dolor](/)

</gno-columns>

## A community of gnomes

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ultricies ultricies risus. Cras libero est, congue vitae ex nec, sodales rutrum leo. Quisque a blandit ex, at viverra ex.

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

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vivamus ultricies ultricies risus. Cras libero est, congue vitae ex nec, sodales rutrum leo. Quisque a blandit ex, at viverra ex. Maecenas quis neque quis augue laoreet ultrices. Integer posuere hendrerit feugiat.

[/read the documentation](docs.gno.land)

<gno-columns>

### Builder Portal

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

[/read more](/)

### Bounties

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

[/read more](/)

### gno.land grants

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec gravida gravida. 

[/read more](/)

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
