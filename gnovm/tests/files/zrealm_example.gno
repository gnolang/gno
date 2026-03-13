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
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:14](257)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14",
//         "LastObjectSize": "257",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:13](485)={
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
//                 "ID": "gno.land/p/nt/avl/v0.Tree"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "850bf1e3c9734920700335f84e2f28ddab1e9058",
//                 "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "LastObjectSize": "485",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:12](346)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "LastObjectSize": "346",
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
//             "Hash": "722cf7135a5cc8c2d4c2780e25d301e323679363",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13"
//         }
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:11](687)={
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
//                     "Hash": "13619df96c2881a75db741d58eda5a8fc63f9b56",
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
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "LastObjectSize": "687",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:10](343)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "cd63fef940bf453a20c5b84fa2bc20437e8b689a",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//         }
//     }
// }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:8](5)=
//     @@ -17,11 +17,11 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "913587fc42b7ee6e72e442d449d09f28553f5b87",
//                      "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
//                  }
//              },
//              {
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -31,7 +31,7 @@
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "LastObjectSize": "449",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//              "RefCount": "1"
//          }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:9](135)=
//     @@ -7,13 +7,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/nt/avl/v0.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "50ee4992b63003a4160218ec1485e7e3cbc69f72",
//     +                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//              "LastObjectSize": "255",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "RefCount": "1"
//          }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//              "LastObjectSize": "343",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "c6843bf0fe0cc1546315f5b53a0e1217a5296b7e",
//     +            "Hash": "2d0d8ccbb6e189357736df7b998060b18505fbe3",
//                  "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8"
//              }
//          }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:3",
//              "LastObjectSize": "393",
//     -        "ModTime": "6",
//     +        "ModTime": "14",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "aa4247f0fc89c9747982dca6f88d57b5a860175a",
//     +                "Hash": "f0ac2be39dc1f9c89b6abce1155dbd685819aa2e",
//                      "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7"
//                  },
//                  "Index": "0",
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:2",
//              "IsEscaped": true,
//              "LastObjectSize": "680",
//     -        "ModTime": "9",
//     +        "ModTime": "14",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1fbaf4a3f57bd7462e872f8775fd0f90c3f9281a",
//     +                "Hash": "3a8d5652d0c8aeec53cf0961ec1a05fd44b305e0",
//                      "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:3"
//                  }
//              },
