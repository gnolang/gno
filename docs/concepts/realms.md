---
id: realms
---

# Realms

A realm refers to a specific instance of a smart contract that can be written
in [Gnolang](./gno-language.md). The most important characteristics of realms are the following:

* Realms are stateful
* Realms can own assets ([coins](todo link concepts/coin))
* Each realm is deployed under a unique package path, i.e. `gno.land/r/blog`, and also has a Gno 
  address derived from it, i.e. `g1n2j0gdyv45aem9p0qsfk5d2gqjupv5z536na3d`
* They are deployed with a package path beginning with `gno.land/r/`
* Realms can import packages from `gno.land/p/demo/` to gain more functionality
* Realms can implement `Render(path string) string` to simplify dApp frontend development by allowing users to request
  markdown renderings from validators and full nodes without a transaction

The potentials of realms are endless - you can create virtually any
application in your mind with built-in composability,
transparency, and censorship resistance. Here are some ideas of what you can build with realms:

* Self-custodial financial exchanges (decentralized exchanges).
* Lending platforms with better rates.
* Transparent insurance systems.
* Fair and accessible voting systems.
* Logistics and supply chain networks. // todo add non-blockchain stuff? ie r/GH, twitter clone, svg generator, gnochess?

Example realms can be found on the Gno monorepo in the [examples folder](https://github.com/gnolang/gno/tree/master/examples/gno.land/r), or on-chain, under the `gno.land/r/` path.

// todo move to a new page? explain how to utilize arg path for muxing?
A notable feature of realms is the `Render()` function.

```go
package demo

func Render(path string) string {
	return "# Hello Gno!"
}
```

Upon calling the realm function above, `# Hello Gno!` will be returned with a string-typed `path` declared in an argument. It should be
noted that while the `path` argument included in the sample code is not utilized, it serves the purpose of
distinguishing the path during the rendering process.
