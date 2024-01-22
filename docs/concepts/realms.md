---
id: realms
---

# Realms

A realm refers to a specific instance of a smart contract that can be written
in [Gno](./gno-language.md). The potentials of realms are endless - you can create virtually any
application in your mind with built-in composability,
transparency, and censorship resistance. Here are some ideas of what you can build with realms:

* Self-custodial financial exchanges (decentralized exchanges).
* Lending platforms with better rates.
* Transparent insurance systems.
* Fair and accessible voting systems.
* Logistics and supply chain networks.

## Packages vs Realms

#### [**Pure Packages**](https://github.com/gnolang/gno/tree/master/examples/gno.land/p)

* A unit that contains functionalities and utilities that can be used in realms.
* Packages are stateless.
* The default import path is `gno.land/p/~~~`.
* Can be imported to other realms or packages.
* Cannot import realms.

#### [**Realms**](https://github.com/gnolang/gno/tree/master/examples/gno.land/r)

* Smart contracts in Gno.
* Realms are stateful.
* Realms can own assets (tokens).
* The default import path is `gno.land/r/~~~`.
* Realms can implement `Render(path string) string` to simplify dapp frontend development by allowing users to request
  markdown renderings from validators and full nodes without a transaction.

A notable feature of realms is the `Render()` function.

```go
package demo

func Render(path string) string {
	return "# Hello Gno!"
}
```

Upon calling the realm above, `# Hello Gno!` is printed with a string-typed `path` declared in an argument. It should be
noted that while the `path` argument included in the sample code is not utilized, it serves the purpose of
distinguishing the path during the rendering process.
