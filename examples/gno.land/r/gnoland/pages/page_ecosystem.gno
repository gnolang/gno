package gnopages

func init() {
	var (
		path  = "ecosystem"
		title = "Discover Gno.land Ecosystem Projects & Initiatives"
		// XXX: description = "Dive further into the Gno.land ecosystem and discover the core infrastructure, projects, smart contracts, and tooling we’re building."
		body = `# Gno Ecosystem

## Gno.land Space

For the best onboarding experience, head over to [Gno.land Space](https://www.gnoland.space/) open ecosystem. Here you can set up your Gno wallet, explore existing community-written Gno smart contracts (realms), and become part of our vibrant community by joining [Gno.land Discord](https://discord.com/invite/x76qK4ttHC).

## Gno Studio (IDE)

Gno IDE is a web-based application helping builders quickly spin up Gno realms and packages right on their browsers. Offering a smooth and intuitive UX for building on Gno, you’ll find multiple modes for customizability with all the features you’d expect from an IDE, such as auto compilation in the editor, debugging, and extensive testing capability.

## Gnoscan

Developed by the Onbloc team, Gnoscan is Gno.land’s blockchain explorer. Anyone can use Gnoscan to easily find information that resides on the Gno.land blockchain, such as wallet addresses, TX hashes, blocks, and contracts. Gnoscan makes our on-chain data easy to read and intuitive to discover. [Go to Gnoscan.](https://gnoscan.io/)

## Adena

Adena is a user-friendly non-custodial wallet for Gno.land. Open-source and developed by Onbloc, Adena currently powers all transactions on Gno.land, allowing gnomes to interact easily with the chain. With an emphasis on UX, Adena is built to handle millions of realms and tokens with a high-quality interface, support for NFTs and custom tokens, and seamless integration. [Get started here.](https://adena.app/)

## Gnoswap

Gnoswap is currently under development and led by the Onbloc team. Gnoswap will be the first DEX on Gno.land and is an automated market maker (AMM) protocol written in Gnolang that allows for permissionless token exchanges on the platform.

## Gno.land Developer Portal

Through the Gno.land Developer Portal, new developers can explore the exciting world of Gnolang (Gno), a novel programming language that powers the Gno.land blockchain. If you want to interact with Gno.land, start writing a realm, build a dApp, or even port a Solidity contract to a Gnolang realm, you’ll find the resources to [get started here](https://docs.onbloc.xyz/).`
	)
	_ = b.NewPost("", path, title, body, nil)
}
