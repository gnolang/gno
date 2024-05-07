// PKGPATH: gno.land/r/demo/nft_test
package nft_test

import (
	"gno.land/p/demo/testutils"
	"gno.land/r/demo/nft"
)

func main() {
	addr1 := testutils.TestAddress("addr1")
	// addr2 := testutils.TestAddress("addr2")
	grc721 := nft.GetToken()
	tid := grc721.Mint(addr1, "NFT#1")
	println(grc721.OwnerOf(tid))
	println(addr1)
}

// Output:
// g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5
// g1v9jxgu33ta047h6lta047h6lta047h6l43dqc5

// Realm:
// switchrealm["gno.land/r/demo/nft"]
// switchrealm["gno.land/r/demo/nft"]
// c[67c479d3d51d4056b2f4111d5352912a00be311e:8]={
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
//                 "ID": "gno.land/p/demo/grc/grc721.TokenID"
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
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:8",
//         "ModTime": "0",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:7",
//         "RefCount": "1"
//     }
// }
// c[67c479d3d51d4056b2f4111d5352912a00be311e:7]={
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
//                     "ID": "gno.land/r/demo/nft.NFToken"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/nft.NFToken"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "c06f58d0ff2bc26ad3e65e953b127a0d03353e97",
//                         "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:8"
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:7",
//         "ModTime": "0",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:5",
//         "RefCount": "1"
//     }
// }
// u[67c479d3d51d4056b2f4111d5352912a00be311e:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/avl.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "45a64533aa57b49b6b4a1d3f6de79db8bea3a710",
//                         "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:7"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:5",
//         "ModTime": "6",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:4",
//         "RefCount": "1"
//     }
// }
// u[67c479d3d51d4056b2f4111d5352912a00be311e:4]={
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
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "dad3106a54e1facb92bce473898b8aec0eb930ff",
//                 "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:5"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "05ab6746ea84b55ca133806af215d99a1c4b045e",
//                 "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:6"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:4",
//         "ModTime": "6",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:2",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/nft"]
// switchrealm["gno.land/r/demo/nft_test"]
