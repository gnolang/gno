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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//             "Hash": "02df322f4d74a1b078106d37d0a99616ac4e9600",
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
//                     "Hash": "634933c69b1fc2340983008134fb1ca580c95d2e",
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//             "Hash": "0de90f76ee02ab7656b1765499bfbf2c287f9d3d",
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
//                     "Hash": "815e826cb0527620cb35ca0bbaeb30c15cdae3be",
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "b256504b4429432d570d59a910e24eb5c7d88102",
//                 "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:5"
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "27bc8ea689f6e5f467e7e6ae78ad2aff2f398bdb",
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
