// PKGPATH: gno.land/r/demo/grc20dupsignal

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
package grc20dupsignal

import (
	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/seqid/v0"
)

var tokenIDSeq seqid.ID

func nextIDSeq() string {
	return tokenIDSeq.String()
}

func main(cur realm) {
	gen := grc20.NewIDGenerator(0, cur, nextIDSeq)

	first, firstLedger := grc20.NewToken("Same", "DUP", 6, gen, cur)
	second, secondLedger := grc20.NewToken("Same", "DUP", 6, gen, cur)

	println("same id:", first.ID() == second.ID())

	// Two independent ledgers, one identifier: these Transfers cannot be
	// attributed to either object from the event stream alone.
	firstLedger.Mint(cur.Address(), 1)
	secondLedger.Mint(cur.Address(), 999999)
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
//         "value": "gno.land/r/demo/grc20dupsignal.DUP.gno.land/r/demo/grc20dupsignal:0000000"
//       },
//       {
//         "key": "name",
//         "value": "Same"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       },
//       {
//         "key": "decimals",
//         "value": "6"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20dupsignal.DUP.gno.land/r/demo/grc20dupsignal:0000000"
//       },
//       {
//         "key": "name",
//         "value": "Same"
//       },
//       {
//         "key": "symbol",
//         "value": "DUP"
//       },
//       {
//         "key": "decimals",
//         "value": "6"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20dupsignal.DUP.gno.land/r/demo/grc20dupsignal:0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1k29lxz0d8gx7m94n032ulum875p7gg0kduus5m"
//       },
//       {
//         "key": "value",
//         "value": "1"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc20dupsignal.DUP.gno.land/r/demo/grc20dupsignal:0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1k29lxz0d8gx7m94n032ulum875p7gg0kduus5m"
//       },
//       {
//         "key": "value",
//         "value": "999999"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc20"
//   }
// ]
