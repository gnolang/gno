// PKGPATH: gno.land/r/nft_test
package nft_test

import (
	"gno.land/p/testutils"
	"gno.land/r/nft"
)

func main() {
	addr1 := testutils.TestAddress("addr1")
	//addr2 := testutils.TestAddress("addr2")
	grc721 := nft.GetGRC721()
	tid := grc721.Mint(addr1, "NFT#1")
	println(grc721.OwnerOf(tid))
	println(addr1)
}

// Output:
// g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5
// g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5

// Realm:
// switchrealm["gno.land/r/nft"]
// switchrealm["gno.land/r/nft"]
// c[6bde79a5f04d2658d17cbc323e45df5fad216511:6]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.Address"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/grc/grc721.TokenID"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "NFT#1"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6bde79a5f04d2658d17cbc323e45df5fad216511:6",
//         "ModTime": "0",
//         "OwnerID": "6bde79a5f04d2658d17cbc323e45df5fad216511:5",
//         "RefCount": "1"
//     }
// }
// c[6bde79a5f04d2658d17cbc323e45df5fad216511:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/nft.NFToken"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/nft.NFToken"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "2216c79f21a1e1b13cacb160ad0499a77518fb3a",
//                         "ObjectID": "6bde79a5f04d2658d17cbc323e45df5fad216511:6"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6bde79a5f04d2658d17cbc323e45df5fad216511:5",
//         "ModTime": "0",
//         "OwnerID": "6bde79a5f04d2658d17cbc323e45df5fad216511:4",
//         "RefCount": "1"
//     }
// }
// u[6bde79a5f04d2658d17cbc323e45df5fad216511:4]={
//     "Fields": [
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "428f989586370a95fdd37c1a46bbd5fa51cd6e6a",
//                         "ObjectID": "6bde79a5f04d2658d17cbc323e45df5fad216511:5"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "6bde79a5f04d2658d17cbc323e45df5fad216511:4",
//         "ModTime": "4",
//         "OwnerID": "6bde79a5f04d2658d17cbc323e45df5fad216511:2",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/nft"]
// switchrealm["gno.land/r/nft_test"]
