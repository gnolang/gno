// PKGPATH: gno.land/r/demo/grc721veto

// Extension hooks run before the Transfer event is emitted, so a hook that vetoes
// a movement leaves nothing on the event stream. Only the second mint, which no
// hook refuses, is announced.
package grc721veto

import (
	"gno.land/p/demo/tokens/grc721"
)

type vetoExtension struct {
	refuse grc721.TokenID
}

func (e *vetoExtension) ExtensionKind() string { return "veto" }

func (e *vetoExtension) OnMint(to address, tid grc721.TokenID) {
	if tid == e.refuse {
		panic("veto: mint refused")
	}
}

func (e *vetoExtension) OnTransfer(from, to address, tid grc721.TokenID) {}

func (e *vetoExtension) OnBurn(tid grc721.TokenID) {}

func main(cur realm) {
	_, ledger := grc721.NewToken("Veto", "VETO", 0, cur)
	ledger.RegisterExtension(&vetoExtension{refuse: "1"})

	func() {
		defer func() { println("recovered:", recover()) }()
		ledger.Mint(cur.Address(), "1")
	}()

	ledger.Mint(cur.Address(), "2")
}

// Output:
// recovered: veto: mint refused

// Events:
// [
//   {
//     "type": "NewToken",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc721veto.VETO.0000000"
//       },
//       {
//         "key": "name",
//         "value": "Veto"
//       },
//       {
//         "key": "symbol",
//         "value": "VETO"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc721"
//   },
//   {
//     "type": "Transfer",
//     "attrs": [
//       {
//         "key": "token",
//         "value": "gno.land/r/demo/grc721veto.VETO.0000000"
//       },
//       {
//         "key": "from",
//         "value": ""
//       },
//       {
//         "key": "to",
//         "value": "g1dywcn8v2jdpl6fuks8tw3l8gnzer0hv8mfkraa"
//       },
//       {
//         "key": "tokenId",
//         "value": "2"
//       }
//     ],
//     "pkg_path": "gno.land/p/demo/tokens/grc721"
//   }
// ]
