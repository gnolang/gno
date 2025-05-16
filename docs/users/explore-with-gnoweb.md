# Exploring gno.land with gnoweb

`gnoweb` is gno.land's universal web interface that lets you browse applications
and source code on any gno.land network. This guide explains how to use gnoweb
to explore the blockchain ecosystem.

## Networks

The main gnoweb instance is available at [gno.land](https://gno.land), which serves the Staging network.

For a complete list of all available networks (testnets and more), see [Networks](../resources/gnoland-networks.md).

## Understanding Code Organization

Before diving into `gnoweb`, we need to cover a fundamental concept in gno.land:
code organization.

gno.land can host two types of code: [realms](../resources/realms.md) (smart contracts),
and [pure packages](../resources/gno-packages.md) (libraries). Realms can
contain and manage state, while pure packages are used for creating reusable
functionality, hence _pure_.

gno.land employs a storage system which is similar to a classic file system - each
package lives on a specific package path. A typical gno.land package path, such
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
- `namespace` is the namespace of the package. Namespaces can be registered using
  the `gno.land/r/gnoland/users` realms, granting a user permission to deploy under
  that specific namespace.
- `package name` represents the name of the package found on the path. This part has
  to match the top-level package declaration in Gno files.

## Viewing Rendered Content

Realms can implement a special `Render()` function that returns HTML-like content:

`gnoweb` is a minimalistic web server that serves as a unified frontend for all
realms in gno.land. It uses ABCI queries to get the latest state of a specific
realm from the gno.land network.

Let's dive into how this works.

### Realm state rendering

In line with minimalistic principles, gno.land encourages developers to implement
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

Based on the provided path, `gnoweb` queries the gno.land network using the
`qrender` ABCI query. It then renders the response data as Markdown.

The realm above can be found on the Staging network at [`gno.land/r/docs/hello`](https://gno.land/r/docs/hello).

While JS/TS clients for Gno exist and developers can create custom websites for their
gno.land applications as they see fit, the approach `gnoweb` takes with `Render()`
is a surefire way for simplicity and ease of development.

:::info `Render()` is optional
Developers can but do not have to provide a `Render()` function in their realms.
Custom getter methods tailored to the specifics of the realm can be built instead.
:::

### Viewing source code

All code uploaded to gno.land is open-source and available for everyone to see,
by design.

Visit the [`gno.land/r/docs/source`](https://gno.land/r/docs/source) realm to learn
how you can do this.
