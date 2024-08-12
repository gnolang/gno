/*
Package openocean is a library that will help you create your own NFT markeplace, following the GRC721

Example:

	import (
		"gno.land/p/demo/openocean"
		"gno.land/p/demo/avl"
		"std"
	)
	
	var collections *avl.Tree
	var collectionID uint64
	var nftsID uint64


	func CreateMyFirstCollection(
		collectionID int,
	) {
		caller := std.GetOrigCaller()
		symbol := "TST"
		name := "test"
		description := "this is the test collection"
		logo := "ipfs link..."
		var avaiableNfts uint64 = 5

		clt, err := openocean.CreateCollection(
			collectionId,
			&nftsID,
			name,
			symbol,
			caller,
			description,
			logo,
			avaiableNfts,
		)

		if err != nil {
			panic(err)
		}
	}
*/

package openocean // import "gno.land/p/demo/openocean"
