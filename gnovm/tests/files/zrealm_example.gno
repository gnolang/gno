// PKGPATH: gno.land/r/example
package example

import (
	"gno.land/p/demo/dom"
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
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:14:0](260)={
//     "Fields": [
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14:0",
//         "LastObjectSize": "260",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13:0",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:13:0](490)={
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
//                 "ID": "gno.land/p/demo/avl.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "af788c885718d0d9e70600639c5a306078fdc115",
//                 "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14:0"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13:0",
//         "LastObjectSize": "490",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12:0",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:12:0](349)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12:0",
//         "LastObjectSize": "349",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a3a55a8cba2bc2d5ba87e1b095f0dc21b02038bc",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13:0"
//         }
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:11:0](689)={
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
//                     "ID": "gno.land/p/demo/dom.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "a8d63d7f75a18152a5d151a5995ac176cf36bce8",
//                     "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12:0"
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11:0",
//         "LastObjectSize": "689",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10:0",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:10:0](348)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10:0",
//         "LastObjectSize": "348",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "bcaf8dd674a712da8e459fbbd62f92daccd5fa1d",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11:0"
//         }
//     }
// }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:9:0](138)=
//     @@ -7,13 +7,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/demo/avl.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "dc6b175496dd3c6b15b9f3794a8e93946567f963",
//     +                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10:0"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9:0",
//              "LastObjectSize": "258",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8:0",
//              "RefCount": "1"
//          }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:8:0](5)=
//     @@ -17,11 +17,12 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "878dfe3ae10e615c09dfda4009c4f421f90eb16c",
//     +                "Hash": "889a454042f7bc1d0a0e7451bb180b1a2de80772",
//                      "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9:0"
//                  }
//              },
//              {
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -31,7 +32,7 @@
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8:0",
//              "LastObjectSize": "454",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7:0",
//              "RefCount": "1"
//          }
