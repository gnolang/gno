package home

var body = `# Welcome to **Gno.land**

- [About Gno.land](/about)
- [Blogs](/r/gnoland/blog)
- [Install ` + "`" + `gnokey` + "`" + `](https://github.com/gnolang/gno/tree/master/gno.land/cmd/gnokey)
- [Acquire testnet tokens](/faucet)
- [Game of Realms](/game-of-realms) - An open worldwide competition for developers to build the best Gnolang smart-contracts.

# Explore new packages.

- r/gnoland
  - [/r/gnoland/blog](/r/gnoland/blog)
  - [/r/gnoland/faucet](/r/gnoland/faucet)
- r/system
  - [/r/system/names](/r/system/names)
  - [/r/system/rewards](/r/system/rewards)
  - [/r/system/validators](/r/system/validators)
- r/demo
  - [/r/demo/banktest](/r/demo/banktest)
  - [/r/demo/boards](/r/demo/boards)
  - [/r/demo/foo20](/r/demo/foo20)
  - [/r/demo/nft](/r/demo/nft)
  - [/r/demo/types](/r/demo/types)
  - [/r/demo/users](/r/demo/users)
  - [/r/demo/groups](/r/demo/groups)
- p/demo
  - [/p/demo/avl](/p/demo/avl)
  - [/p/demo/blog](/p/demo/blog)
  - [/p/demo/flow](/p/demo/flow)
  - [/p/demo/gnode](/p/demo/gnode)
  - [/p/demo/grc/exts](/p/demo/grc/exts)
  - [/p/demo/grc/grc20](/p/demo/grc/grc20)
  - [/p/demo/grc/grc721](/p/demo/grc/grc721)

# Other Testnets

- **[staging.gno.land](https://staging.gno.land) (wiped every commit to master)**
- _[test3.gno.land](https://test3.gno.land) (latest)_
- _[test2.gno.land](https://test2.gno.land) (archive)_
- _[test1.gno.land](https://test1.gno.land) (archive)_

**This is a testnet.**
Package names are not guaranteed to be available for production.

# Social

Check out our [community projects](https://github.com/gnolang/awesome-gno).

Official channel: [Discord](https://discord.gg/S8nKUqwkPn)<br />
Other channels: [Telegram](https://t.me/gnoland) [Twitter](https://twitter.com/_gnoland) [Youtube](https://www.youtube.com/@_gnoland)
`

var adminAddr std.Address = std.Address("")

func Render(_ string) string {
	// highest level == highest responsibility
	// teams are responsible for components they don't owne
	// flag : realm maintainers VS facilitators
	// teams
	// committee of trustees to create the directory
	// each directory is a name, has a parent and have groups
	// homepage team - blocks aggregating events
		`
# Directory

* gno.land (owned by group)
  * 
* gnovm
  * gnolang (language)
  * gnovm
    - current challenges / concerns / issues
* tm2
  * amino
  * 

## Contributors
``


	// XXX: dynamic content: contributions
	// XXX: 
	// XXX: community curated picks
	// XXX: 
	return body
}
