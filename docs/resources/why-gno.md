# Gno.land

Gno.land represents paradigm shift in multi-user programming that no other
solution offers. It is not just a blockchain; it is the world's first viable
language-based multi-user operating system. Its ultimate goal is to host the
world's knowledge base for the new millennium.

Tendermint changed the way blockchain developers think about blockchain
consensus algorithms. Gno.land will change the way developers think about
programming; we won't remember a time without it.

Cosmos Hub (gaia) was the beginning of the Cosmos Network, a web of
blockchains.

## Why Gno.land?

Compare publishing a blog site in Gno.land to all prior smart contract systems.

```go
// Realm: gno.land/r/me/myblog

package gnoblog

import (
	"std"

	"gno.land/p/demo/blog"
)

var b = &blog.Blog{
	Title:  "gno.land's blog",
	Prefix: "/r/gnoland/blog:",
}

func AddComment(postSlug, comment string) {
    crossing()
	assertIsCommenter()
	assertNotInPause()

	caller := std.OriginCaller()
	err := b.GetPost(postSlug).AddComment(caller, comment)
	checkErr(err)
}

func Render(path string) string {
	return b.Render(path)
}
```

Q: Why is everything else so complicated?

A: Because our languages, compilers, interpreters, and programming paradigm is
immature.

## Brief Evolution of Language

Written human language has only been around for a mere 6000 years, a blip in
our evolutionary history. Like living species our language and writing have
evolved along side us and within us. Adam was not the first homo sapiens on
earth, but he may have been the first with written language, and thereby a new
kind of man. 

Programming languages likewise has been evolving rapidly, but only for a
handful of decades; it was in the 1970s when Alan Kay developed Smalltalk, the
first object oriented programming language. In the 1990’s Brendan Eich of
Netscape invented Javascript which forever transformed the World Wide Web; Sun
Microsystem made Java, and industries prospered greatly by these and similar
language technologies. 

## Gno vs Previous

Our languages, compilers & interpreters, and programs are today:
 - Nondeterministic - randomness is the norm in concurrent programming, but
   even Go randomizes map iteration. 
 - Disk Bound - programs need to be designed to save to disk —> SQL solutions;
   NOT native language
 - Dependent - running programs are owned by an owner; dependent on
   individuals, not self-sustaining
 - Ephemeral -  running programs are expected to fail; no guarantee of
   presence.
 - Single User Realm - import of internal libraries are native, but
   interactions with external programs are NOT native; generally no `import
   “gno.land/r/external/realm”`, but leaky abstractions synthesized ie GRPC

Gno, GnoVM, and Gno.land is in contrast:
 - Deterministic - gno routines not yet supported, but even these will be
   deterministic.
 - Auto Persistent - all changes to instantiated Gno objects in the transaction
   are persisted transparently.
 - Self Sustaining - every transaction locks GNOT up for new storage allocated;
   CPU gas fees paid in any language.
 - Immortal - every Gno object that is referenced (not GC’d) remains forever.
 - Multi User Realm - all objects are stored in realm packages (namespaces). 

## Gno Language Innovation

Gno the language has the same semantics as Gno, but due to the multi-user
nature of Gno there are extra semantics for inter-realm interoperability.
 - `cross(fn)(…)` calls `fn(…)` where fn is another realm.
 - `std.CurrentRealm()` and `std.PreviousRealm()` changes upon cross-calls.
 - `func fn() { crossing(); … }` signifies that fn is a crossing-function where
   std.CurrentRealm() returns the realm in which the function is declared.
 - Gno2 proposed syntax: `@fn(…)`, `@func @fn() { … }`
 - These are like verb (function) modifiers in honorifics in Korean and
   Japanese: https://en.wikipedia.org/wiki/Honorifics_(linguistics) 
 - Type checking cross-calls isn’t strictly necessary, but is e.g. for
   financial systems, aids development.
 - While all data is readable by other realms, dot.selector access
   across realms get tainted with 'readonly' attribute.
 - Function/method return implies access without readonly taint.
 - Inter-realm type conversion limitations to prevent exploits.
 - More and refinements to come in Gno2.

These language innovations/extensions allow for safer multi-user application
development where many users are collaboratively programming a single immortal
communal program.

## The Logoverse

Ἐν ἀρχῇ ἦν ὁ Λόγος καὶ ὁ Λόγος ἦν πρὸς τὸν Θεόν καὶ Θεὸς ἦν ὁ Λόγος.  In the
beginning was the Word (Logos), and the Word was with God, and the Word was
God. - John 1:1

Logos means “word, discourse; reason”, and shares its root with the word
“logic”. 

With these elements altogether you can derive a new property:
 - Gno expressions become "real" on Gno.land.
 - Ethereum comes close but isn't object-oriented and Solidity has no pronouns.
 - TBL's WWW, DOM model, HTTP verbs, Plan 9, Ethereum, and FB Meta are all
   attempts to arrive at the logoverse.
 - Gno.land is the first complete logoverse.

## Gno.land License

Anyone can make Gno VM powered chains derived from Gno.land according to the
viral copyleft license terms and strong attribution requirement. The Strong
Attribution clause of the Gno Network GPL license preserves the spirit of the
GNU AGPL license for the blockchain world.

GNOT is the storage lock-up token, so Gno.land is to Gno England like GNOT is
to fungible land rights in Gno England, where total storage is kept finite for
very-long-term existential purposes, and value is derived from the Gno
artifacts created by its users, and some new users competing for attention from
many existing users.

## Tokenomics

XXX include tokenomics

Gno.land may migrate to AtomOne ICS once it is support hard-fork upgrades.
There Gno.land would be one ICS shard, and many Gno VM shards may also exist,
each with their own namespace and probably each their own storage token unless
separate treaties are made between the main Gno.land chain (ICS shard) and
other Gno VM shards.

## Adoption Strategy

Gno.land and its associated network of Gno VM chains, and AtomOne if it hosts
it, will become the nexus of human to human, human to machine, and machine to
machine coordination; but only after it finds a self-sustaining organic growth
cycle.

The best way to ensure success and to accelerate adoption is to seed the
initial community with the right community. There are many types of
communities, such as crypto community, ethereum community, student community,
but since Bitcoin has gone mainstream these communities aren't always in
agreement about the purpose of blockchain technology; because they aren't aware
of the history and fabric of the hidden power structures that run the
narrative--both mainstream AND controlled oppositions. They do not feel that
they need something, so their habits are not as obvious to change.

But the "free-thinking" and "conspiracy" and "anti-war" and "anti-Covid19-vax"
and even the "true Christian" communities feel an urgent need for
censorship-proof coordination and communication tools. These communities have
influencers who are kept hidden from the general public; they have suffered
deplatforming, defamations, and even death.

Build tools, connections, and relations with these particular communities and
especially those influencers who are nuanced in their research and speech.
Even those that don't promote crypto will see the benefits uniquely offered by
Gno.land.

## Team

Jae Kwon before and after creating Tendermint and Cosmos always had a passion
for programming languages and wrote multiple parsers and interpreters, and
initially also wrote an EVM on top of the framework which became the Cosmos
SDK. Gno.land is the result of two decades of search for the logoverse.

Manfred Touron, builder focused on open-source and resilient technologies;
co-founded scaleway (cloud) and berty (p2p messaging), with contributions to
900+ open-source projects.

Morgan Bazalgette - Senior Go engineer; bringing the joy of developing Go to
Gno.

Miloš Živković - Sr. distributed systems engineer; passion for solving
protocol-level problems in the blockchain space.

Marc Vertes - Sr. VM dev and hardware; more than 3 decades of experience,
Co-founder of 3 companies (1 acquired by IBM), author of 34 patents, author of
the Yaegi Go interpreter.

Ìlker Öztürk - Sr. software architect; 17 years in building and designing
products, distributed p2p systems, leadership and strategic vision.

XXX Ask everyone to fill in...
