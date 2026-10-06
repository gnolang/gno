package pages

func init() {
	path := "about"
	title := "Gno.land Is A Platform To Write Smart Contracts In Gno"
	// XXX: description := "On Gno.land, developers write smart contracts and other blockchain apps using Gno without learning a language that’s exclusive to a single ecosystem."
	body := `
Gno.land is a next-generation smart contract platform using Gno, an interpreted version of the general-purpose Go
programming language. On Gno.land, smart contracts can be uploaded on-chain only by publishing their full source code,
making it trivial to verify the contract or fork it into an improved version. With a system to publish reusable code
libraries on-chain, Gno.land serves as the “GitHub” of the ecosystem, with realms built using fully transparent,
auditable code that anyone can inspect and reuse.

Gno.land addresses many pressing issues in the blockchain space, starting with the ease of use and intuitiveness of
smart contract platforms. Developers can write smart contracts without having to learn a new language that’s exclusive
to a single ecosystem or limited by design. Go developers can easily port their existing web apps to Gno.land or build
new ones from scratch, making web3 vastly more accessible.

Secured by Proof of Contribution (PoC), a DAO-managed Proof-of-Authority consensus mechanism, Gno.land prioritizes
fairness and merit, rewarding the people most active on the platform. PoC restructures the financial incentives that
often corrupt blockchain projects, opting instead to reward contributors for their work based on expertise, commitment, and
alignment.

One of our inspirations for Gno.land is the gospels, which built a system of moral code that lasted thousands of years.
By observing a minimal production implementation, Gno.land’s design will endure over time and serve as a reference for
future generations with censorship-resistant tools that improve their understanding of the world.
`
	if err := b.NewPost("", path, title, body, "2022-05-20T13:17:22Z", nil, nil); err != nil {
		panic(err)
	}
}
