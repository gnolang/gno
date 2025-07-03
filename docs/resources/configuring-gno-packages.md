# Configuring Gno Packages

Gno supports package configuration through a metadata file called `gnomod.toml`.
This file is typically generated using:

```bash
gno mod init <pkgpath>
```

It enables advanced control over how your package is interpreted, deployed, and 
used on-chain.

## `gnomod.toml`

This file defines metadata for your Gno package and can include the following fields:

#### `private` (coming soon)

Marks the package as private and **unimportable** by any other package. Additionally:
- It can be **re-uploaded** - the new version fully overwrites the old one.
- Memory, pointers, or types defined in this package **cannot be used or stored elsewhere**.
- Usually, this flag can be used for packages that are meant to be changed,
such as the home realm of a specific user (i.e. `r/username/home`). 
- *This flag _does not_ provide any sort of privacy. All code is still fully
open-source and visible to everyone, including the transactions that were used for deployments.

#### `gno`  

Specifies the **Gno language version**. Currently, only version `"0.9"` is supported.

#### `pkgpath`  

Defines the canonical **package path**. Must exactly match the path used in the
`addpkg` transaction during deployment (must it right now?).

#### `replace` (coming soon)

Used for **local development and testing**. When set, this field allows local 
replacement of the package, but will cause `addpkg` to **fail on-chain**. Useful
for overriding dependencies during local testing.

#### `uploader`  

Specifies the address that will be set as the uploader of the package (origin 
of the transaction). This field is used only during **genesis (block 0)** and
replaces the default deployer address if set. Primarily used in monorepo setups.
If not specified, it's automatically set to the address that initiated the `addpkg`
transaction.

#### `draft`  

A flag intended for **chain creators**. Marks the package as *unimportable*
during normal operation. This flag is **ignored at block 0**, allowing draft
packages to be included at genesis.

### Example

```toml
module = "gno.land/r/test/test"
gno = "0.9"

[upload_metadata]
    uploader = "g1t43aega4j3t6szv0d3zt9uhpa5g6k7h8x4vvxe"
```

---

## `gnowork.toml` (Coming Soon)

A future configuration file for specifying build and workspace-level metadata. Stay tuned!
