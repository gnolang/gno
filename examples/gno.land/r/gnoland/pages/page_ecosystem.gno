package gnopages

func init() {
	var (
		path  = "ecosystem"
		title = "Discover gno.land Ecosystem Projects & Initiatives"
		// XXX: description = "Dive further into the gno.land ecosystem and discover the core infrastructure, projects, smart contracts, and tooling we’re building."
		body = `
### [Gno Playground](https://play.gno.land)

Gno Playground is a simple web interface that lets you write, test, and experiment with your Gno code to improve your
understanding of the Gno language. You can share your code, run unit tests, deploy your realms and packages, and execute
functions in your code using the repo.

Visit the playground at [play.gno.land](https://play.gno.land)!

### [Gno Studio Connect](https://gno.studio/connect)

Gno Studio Connect provides seamless access to realms, making it simple to explore, interact, and engage
with gno.land’s smart contracts through function calls. Connect focuses on function calls, enabling users to interact
with any realm’s exposed function(s) on gno.land.

See your realm interactions in [Gno Studio Connect](https://gno.studio/connect)

### [Gnoscan](https://gnoscan.io)

Developed by the Onbloc team, Gnoscan is gno.land’s blockchain explorer. Anyone can use Gnoscan to easily find
information that resides on the gno.land blockchain, such as wallet addresses, TX hashes, blocks, and contracts.
Gnoscan makes our on-chain data easy to read and intuitive to discover.

Explore the gno.land blockchain at [gnoscan.io](https://gnoscan.io)!

### Adena

Adena is a user-friendly non-custodial wallet for gno.land. Open-source and developed by Onbloc, Adena allows gnomes to
interact easily with the chain. With an emphasis on UX, Adena is built to handle millions of realms and tokens with a
high-quality interface, support for NFTs and custom tokens, and seamless integration. Install Adena via the [official website](https://www.adena.app/)

### Gnoswap

Gnoswap is currently under development and led by the Onbloc team. Gnoswap will be the first DEX on gno.land and is an
automated market maker (AMM) protocol written in Gno that allows for permissionless token exchanges on the platform.

### Flippando

Flippando is a simple on-chain memory game, ported from Solidity to Gno, which starts with an empty matrix to flip tiles
on to see what’s underneath. If the tiles match, they remain uncovered; if not, they are briefly shown, and the player
must memorize their colors until the entire matrix is uncovered. The end result can be minted as an NFT, which can later
be assembled into bigger, more complex NFTs, creating a digital “painting” with the uncovered tiles.  Play the game at [Flippando](https://gno.flippando.xyz/flip)

### Gno Native Kit

[Gno Native Kit](https://github.com/gnolang/gnonative) is a framework that allows developers to build and port gno.land (d)apps written in the (d)app's native language.


`
	)
	if err := b.NewPost("", path, title, body, "2022-05-20T13:17:23Z", nil, nil); err != nil {
		panic(err)
	}
}
