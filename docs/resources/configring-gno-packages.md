# Configuring Gno Packages


## gnomod.toml

used as a package metadata file that can be created with `gno mod init pkgpath`

can specify specific fields:
- uploader - replaces the deployer address, only for genesis (block0, working in the monorepo)
- draft
- private
- gno version - currently version 0.9 is the only supported version.
- pkgpath - has to match the addpkg path during transaction deployment
- replace - helps with local testing; if it's not empty addpkg fails on-chain

## gnowork.toml

? 
