// PKGPATH: gno.land/r/demo/grc721dupsignal

// This filetest is the reference for the detection rule the NewToken event
// enables. The realm below reuses seqid 0 twice, so both tokens end up with the
// same Token.ID() and their Transfer events are indistinguishable — the
// long-standing event-provenance ambiguity.
//
// What changes is that the ambiguity is now *announced*. Two NewToken events
// carry the same "token" value, which is a complete signal: NewToken is the only
// way a Token can exist (Token's fields are unexported), so an indexer watching
// this event sees every token that will ever emit. On the second duplicate
// announcement it should flag the realm and stop trusting its token events.
package grc721dupsignal

import (
	"gno.land/p/demo/tokens/grc721"
)

func main(cur realm) {
	first, firstLedger := grc721.NewToken("Same", "DUP", 0, cur)
	second, secondLedger := grc721.NewToken("Same", "DUP", 0, cur)

	println("same id:", first.ID() == second.ID())

	// Two independent ledgers, one identifier: these Transfers cannot be
	// attributed to either object from the event stream alone.
	firstLedger.Mint(cur.Address(), "1")
	secondLedger.Mint(cur.Address(), "2")
}

// Output:
// same id: true

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc721dupsignal.DUP.0000000"
//       },
//       {
//         "key": "name",
//         "value": "Same"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc721"
//   },
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc721dupsignal.DUP.0000000"
//       },
//       {
//         "key": "name",
//         "value": "Same"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc721"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc721dupsignal.DUP.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1f2p7p8jhjmg6atgukt8kveqgsrrk3qqhe6vg9m"
//       },
//       {
//         "key": "tokenId",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc721"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc721dupsignal.DUP.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1f2p7p8jhjmg6atgukt8kveqgsrrk3qqhe6vg9m"
//       },
//       {
//         "key": "tokenId",
//         "value": "2"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc721"
//   }
// ]
