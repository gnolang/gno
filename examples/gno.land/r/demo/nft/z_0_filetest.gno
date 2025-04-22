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
// finalizerealm["gno.land/r/demo/nft"]
// finalizerealm["gno.land/r/demo/nft"]
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
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/nft.NFToken"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "564a9e78be869bd258fc3c9ad56f5a75ed68818f",
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
//                     "Hash": "b53ffc464e1b5655d19b9d5277f3491717c24aca",
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
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "b1d928b3716b147c92730e8d234162bec2f0f2fc",
//             "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:9"
//         }
//     }
// }
// u[67c479d3d51d4056b2f4111d5352912a00be311e:5]=
//     @@ -7,12 +7,22 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/demo/avl.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "b229b824842ec3e7f2341e33d0fa0ca77af2f480",
//     +                    "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:8"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:5",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:4",
//              "RefCount": "1"
//          }
// u[67c479d3d51d4056b2f4111d5352912a00be311e:4]=
//     @@ -2,6 +2,7 @@
//          "Fields": [
//              {},
//              {
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -14,7 +15,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "872ac4fb1c12756da1e64712e7496f1d8db57ee0",
//     +                "Hash": "1e0b9dddb406b4f50500a022266a4cb8a4ea38c6",
//                      "ObjectID": "67c479d3d51d4056b2f4111d5352912a00be311e:5"
//                  }
//              },
//     @@ -32,7 +33,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "67c479d3d51d4056b2f4111d5352912a00be311e:4",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "67c479d3d51d4056b2f4111d5352912a00be311e:3",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/demo/nft"]
// finalizerealm["gno.land/r/demo/nft_test"]
