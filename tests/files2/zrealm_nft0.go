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
//                 "@type": "/gno.tref",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "std.Address"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tref",
//                 "ID": "gno.land/r/nft.TokenID"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
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
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/nft.NFToken"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/nft.NFToken"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "d6f02aad674df014ea8e7057d28f345b8185fd91",
//                         "ObjectID": "6bde79a5f04d2658d17cbc323e45df5fad216511:6"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/p/avl.Tree"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/p/avl.Tree"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "5bc850fde6b94814d4ab7aac069a70adb6d90ab2",
//                         "ObjectID": "6bde79a5f04d2658d17cbc323e45df5fad216511:5"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
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
