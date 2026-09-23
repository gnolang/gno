# Exploring Gno.land with gnoweb

`gnoweb` is Gno.land's universal web interface that lets you browse applications
and source code on any Gno.land network. This guide explains how to use gnoweb
to explore the blockchain ecosystem.

## Networks

The main gnoweb instance is available at [gno.land](https://gno.land), which serves the Betanet (`gnoland1`). The Staging network is accessible at [staging.gno.land](https://staging.gno.land).

For a complete list of all available networks (testnets and more), see [Networks](../resources/gnoland-networks.md).

## Understanding Code Organization

Before diving into `gnoweb`, we need to cover a fundamental concept in Gno.land:
code organization.

Gno.land can host two types of code: [realms](../resources/realms.md) (smart contracts),
and [pure packages](../resources/gno-packages.md) (libraries). Realms can
contain and manage state, while pure packages are used for creating reusable
functionality, hence _pure_.

Gno.land employs a storage system which is similar to a classic file system - each
package lives on a specific package path. A typical Gno.land package path, such
as `gno.land/r/gnoland/home`, contains the following components:

```
  gno.land     /     r     /    gnoland    /      home
chain domain        type       namespace       package name
```

Let's break it down:
- `chain domain` represents the domain of the chain. In this case, the domain is
  simply `gno.land`. In the future, the ecosystem may expand to multiple chains
  which could have different chain domains.
- `type` represents the type of package found on this path. There are two available
  options - `p` & `r` - pure packages and realms, respectively.
- `namespace` is the namespace of the package. Currently only address-prefix
  namespaces are supported, where the namespace matches the deployer's address.
- `package name` represents the name of the package found on the path. This part has
  to match the top-level package declaration in Gno files.

## Viewing Rendered Content

Realms can implement a special `Render()` function that returns Markdown content:

`gnoweb` is a minimalistic web server that serves as a unified frontend for all
realms in Gno.land. It uses ABCI queries to get the latest state of a specific
realm from the Gno.land network.

Let's dive into how this works.

### Realm state rendering

In line with minimalistic principles, Gno.land encourages developers to implement
a `Render()` function within their realms, allowing them to create a Markdown view
for how their realms will be rendered. `gnoweb` utilizes a built-in Markdown renderer
that uses the output of the `Render()` function as its content source.

A simple example of a realm utilizing the Render function can be found below:

```go
package hello

func Render(path string) string {
	if path == "" {
		return "# Hello, 世界！"
	}

	return "# Hello, " + path
}
```

Based on the provided path, `gnoweb` queries the Gno.land network using the
`qrender` ABCI query. It then renders the response data as Markdown.

The realm above can be found on the Staging network at [`gno.land/r/docs/hello`](https://staging.gno.land/r/docs/hello).

While JS/TS clients for Gno exist and developers can create custom websites for their
Gno.land applications as they see fit, the approach `gnoweb` takes with `Render()`
is a surefire way for simplicity and ease of development.

:::info `Render()` is optional
Developers can but do not have to provide a `Render()` function in their realms.
Custom getter methods tailored to the specifics of the realm can be built instead.
:::

### Images and IPFS content

Rendered pages load images only from gnoweb's own assets, `data:` SVG images and
a short list of hosts such as imgur and GitHub (`cspImgHost` in
`gno.land/cmd/gnoweb/main.go`).

To show content stored on IPFS, write an `ipfs://` or `ipns://` URL:

```markdown
![logo](ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi/logo.png)
[whitepaper](ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi/paper.pdf)
```

gnoweb resolves these through its configured IPFS gateway (`-ipfs-gateway`,
`https://ipfs.filebase.io` by default). It also rewrites URLs on retired public
gateways, such as `https://ipfs.io/ipfs/...`, to that gateway. Prefer `ipfs://`
to a gateway URL, so the gateway can change without touching your realm. The
content still has to be pinned somewhere on the IPFS network: a gateway only
serves what some node provides.

:::info
The URL helpers in `p/nt/markdown/sanitize/v0`, which `p/moul/md` uses, reject
`ipfs://` URLs. Write the markdown link or image directly instead.
:::

### Viewing source code

All code uploaded to Gno.land is open-source and available for everyone to see,
by design.

Visit the [`gno.land/r/docs/source`](https://staging.gno.land/r/docs/source) realm to learn
how you can do this.

## Alternative: Terminal UI with gnobro

While `gnoweb` provides a web-based interface for exploring realms, developers
who prefer working in the terminal can use `gnobro` - a terminal user interface
(TUI) for browsing realms.

`gnobro` offers:
- Terminal-based navigation of realms
- Direct connection to `gnodev` for local development
- Real-time updates when connected to a development server
- SSH access for remote browsing

To learn more about `gnobro`, see the [gnobro
documentation](https://github.com/gnolang/gno/blob/master/contribs/gnobro/README.md).
