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
// c[67c479d3d51d4056b2f4111d5352912a00be311e:11]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
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
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:11",
//         "ModTime": "0",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:10",
//         "RefCount": "1"
//     }
// }
// c[67c479d3d51d4056b2f4111d5352912a00be311e:10]={
//     "ObjectInfo": {
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:10",
//         "ModTime": "0",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/nft.NFToken"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f6e0719bd3e659c4dc0bb9f971824b6af7ab17a2",
//             "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:11"
//         }
//     }
// }
// c[67c479d3d51d4056b2f4111d5352912a00be311e:9]={
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "3"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/nft.NFToken"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "18039c9262829759a15a77357969420cb4cb4fc2",
//                     "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:10"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:9",
//         "ModTime": "0",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:8",
//         "RefCount": "1"
//     }
// }
// c[67c479d3d51d4056b2f4111d5352912a00be311e:8]={
//     "ObjectInfo": {
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:8",
//         "ModTime": "0",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:5",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "26f9b96ec85f5f60d9c5f3ce00fbd0522fee6a68",
//             "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:9"
//         }
//     }
// }
// u[67c479d3d51d4056b2f4111d5352912a00be311e:5]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "1450e654c0ecbdbe6e2e8b2bc994fe94cc073107",
//                     "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:8"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:5",
//         "ModTime": "7",
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
//                 "Hash": "bb9077d3808e9ecb78aed49c7e97161fc9a822cc",
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
//         "ModTime": "7",
//         "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/nft"]
// switchrealm["gno.land/r/demo/nft_test"]
