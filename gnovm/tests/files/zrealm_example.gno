// PKGPATH: gno.land/r/example
package example

import (
	"gno.land/p/archive/dom"
)

var gPlot *dom.Plot

func init() {
	gPlot = &dom.Plot{Name: "First Plot"}
}

func main(cur realm) {
	gPlot.AddPost("TEST_TITLE", "TEST_BODY")
	println(gPlot.String())
}

// Output:
// # [plot] First Plot
//
// ## TEST_TITLE
// TEST_BODY

// Realm:
// finalizerealm["gno.land/r/example"]
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:14](254)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14",
//         "LastObjectSize": "254",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:13](482)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "TEST_TITLE"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "TEST_BODY"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/nt/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "a370ed7fbdb96069c46148384169030024821bec",
//                 "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "LastObjectSize": "482",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:12](347)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "LastObjectSize": "347",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/archive/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "7a953a4e6bf0c4c81e317ab3530a2fdc375cdf91",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13"
//         }
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:11](682)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/archive/dom.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "2d757e6db3d3c489ff8a492370ae0bdb35f40029",
//                     "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12"
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
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "LastObjectSize": "682",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:10](340)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "LastObjectSize": "340",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "72762edf4237f8a3a31b19af08420ecf9f21ebb3",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//         }
//     }
// }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:9](135)=
//     @@ -7,13 +7,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/nt/avl.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "6700cba0c327fa62c5f551329f3d586f6b4a3aaf",
//     +                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//              "LastObjectSize": "252",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "RefCount": "1"
//          }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:8](5)=
//     @@ -17,11 +17,12 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5f8c7561e9dd28d3c2189bb364f0c8a2a7b7188e",
//     +                "Hash": "c50c0225cb926e35e124cedd501d1624a0ebe064",
//                      "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
//                  }
//              },
//              {
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -31,7 +32,7 @@
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "LastObjectSize": "446",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//              "RefCount": "1"
//          }
