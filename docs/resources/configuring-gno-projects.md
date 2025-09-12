# Configuring Gno Projects

Gno supports package configuration through a metadata file called `gnomod.toml`.

The `gnomod.toml` file is typically generated using:

```bash
gno mod init <pkgpath>
```

It enables advanced control over how your package is interpreted, deployed, and 
used on-chain.

> Note: Previously, package configuration was done via the `gno.mod` file -
similar to how Go projects are configured with a `go.mod` file. If you have a 
project with a `gno.mod` file, you can use the `gno mod tidy` subcommand to
auto-convert it to a `gnomod.toml`.

## `gnomod.toml`

This file defines metadata for your Gno package and can include the following fields:

#### `gno`  

Specifies the **Gno language version**. Currently, only version `"0.9"` is supported.

#### `pkgpath`  

Defines the canonical **package path**. Must exactly match the path used in the
`addpkg` transaction during deployment (must it right now?).

#### `replace` (coming soon)

Used for **local development and testing**. When set, this field allows local 
replacement of the package, but will cause `addpkg` to **fail on-chain**. Useful
for overriding dependencies during local testing.

#### `creator`  

Specifies the address that will be set as the creator of the package (origin 
of the transaction). This field is used only during **genesis (block 0)** and
replaces the default deployer address if set. Primarily used in monorepo setups.
If not specified, it's automatically set to the address that initiated the `addpkg`
transaction.

#### `draft`  

A flag intended for **chain creators**. Marks the package as *unimportable*
during normal operation. This flag is **ignored at block 0**, allowing draft
packages to be included at genesis.

#### `private` (coming soon)

Marks the package as private and **unimportable** by any other package. Additionally:
- It can be **re-uploaded** - the new version fully overwrites the old one.
- Memory, pointers, or types defined in this package **cannot be used or stored elsewhere**.
- Usually, this flag can be used for packages that are meant to be changed,
  such as the home realm of a specific user (i.e. `r/username/home`).
- *This flag _does not_ provide any sort of privacy. All code is still fully
  open-source and visible to everyone, including the transactions that were used for deployments.

#### `ignore` 

Coming soon - follow progress [here](https://github.com/gnolang/gno/pull/4413).

### Example

```toml
# gnomod.toml
module = "gno.land/r/test"
gno = "0.9"
draft = true
private = true

[replace]
  old = "gno.land/r/test"
  new = "gno.land/r/test/v2"
[replace]
  old = "gno.land/r/test/v3"
  new = "../.."

[addpkg]
    creator = "g1xyz..."
    height = 123
```

Note that this example isn't realistic because we should either replace,
configure addpkg settings, or do neither, but never both at the same time.

## `gnowork.toml`

`gnomod.toml` is fine for working on a single package but when you want to work on multiple packages depending on each other, you will need a `gnowork.toml`.

For now it is only used to delimit your workspace root and is empty. So in most cases, just `touch gnowork.toml` at your project root and you're good to go.

Packages that import other packages present in your workspace will use the ones present in your workspace instead of attempting to download them.

There is no rules on the project hierarchy for dependencies resolution, so you can freely move your packages around.

Packages in directories that contain other `gnowork.toml`s down your hierarchy will be ignored.