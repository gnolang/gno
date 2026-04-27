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

#### `module`

Defines the canonical **package path** (e.g. `gno.land/r/demo/myapp`). This must
match the path used in the `addpkg` transaction during deployment.

#### `gno`  

Specifies the **Gno language version**. Currently, only version `"0.9"` is supported.

#### `replace` (coming soon)

Used for **local development and testing**. When set, this field allows local
replacement of the package, but will cause `addpkg` to **fail on-chain**. Useful
for overriding dependencies during local testing.

#### `addpkg`

A section containing on-chain metadata, filled by the VM keeper when a module is
added. It is not intended for manual use off-chain.

- **`creator`**: the address of the package creator (origin of the transaction).
  At **genesis (block 0)**, this can override the default deployer address.
  Primarily used in monorepo setups. If not specified, it's automatically set to
  the address that initiated the `addpkg` transaction.
- **`height`**: the block height at which the module was added.

#### `draft`  

A flag intended for **chain creators**. Marks the package as *unimportable*
during normal operation. This flag is **ignored at block 0**, allowing draft
packages to be included at genesis.

#### `private`

Marks the package as private and **unimportable** by any other package. Additionally:
- It can be **re-uploaded** - the new version fully overwrites the old one.
- Memory, pointers, or types defined in this package **cannot be stored elsewhere**.
- Usually, this flag can be used for packages that are meant to be changed,
  such as the home realm of a specific user (i.e. `r/username/home`).
- *This flag _does not_ provide any sort of privacy. All code is still fully
  open-source and visible to everyone, including the transactions that were used for deployments.

#### `ignore`

Marks the module to be **ignored by the Gno toolchain** while still being usable
in development environments.

### Example

```toml
# gnomod.toml
module = "gno.land/r/test"
gno = "0.9"
draft = true
private = true

[[replace]]
  old = "gno.land/r/test"
  new = "gno.land/r/test/v2"

[[replace]]
  old = "gno.land/r/test/v3"
  new = "../.."

[addpkg]
  creator = "g1xyz..."
  height = 123
```

Note that this example isn't realistic because replace directives and addpkg
metadata would not normally coexist: replace directives prevent on-chain deployment.

## Workspaces with `gnowork.toml`

Adding a `gnowork.toml` file to the root of your workspace allows Gno tooling to 
better understand your local environment and resolve dependencies between packages.
This is especially useful during local development and testing with tools like 
`gno test` & `gno lint`. 

Consider the following project structure:

```text
project/
    ├─ p/
    │   └─ library/
    │       ├─ gnomod.toml
    │       └─ lib.gno
    └─ r/
        └─ app/
            ├─ gnomod.toml
            ├─ app.gno
            └─ app_test.gno
```

In this setup, the `app` package cannot use `library` as a dependency during local 
development. The tooling has no way of knowing that both live inside the same
workspace - only dependencies usable are the ones found in `gnohome`, where the 
tooling was initially installed.

By adding a `gnowork.toml` file at the root, Gno tooling can properly link the 
packages together:

```text
project/
    ├─ gnowork.toml
    ├─ p/
    │   └─ library/
    │       ├─ gnomod.toml
    │       └─ lib.gno
    └─ r/
        └─ app/
            ├─ gnomod.toml
            ├─ app.gno
            └─ app_test.gno
```

At the moment, `gnowork.toml` does not have any configuration options and should be
an empty file. In the future, workspace-level configuration options may be added to
support more advanced use cases.

Gno tooling resolves dependencies by first using the packages available in your workspace.
If a dependency cannot be found locally, it will then fall back to other sources,
such as a remote chain.

<!-- TODO: allow configuration of dependency source priority/hierarchy -->

Note: `gnowork.toml` support is a work in progress for `gnodev` and `gnopls`.

#### Cleaning the dependency cache

Downloaded dependencies are stored locally under `$GNOHOME/pkg/mod/`.
If you want to remove them and fetch fresh versions, you can run:

```bash
gno clean -modcache
```
