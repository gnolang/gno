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
// c[0ffd45e074aa1b8df562907c95ad97526b7ca187:14](257)={
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
//         "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:14",
//         "LastObjectSize": "257",
//         "ModTime": "0",
//         "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "RefCount": "1"
//     }
// }
// c[0ffd45e074aa1b8df562907c95ad97526b7ca187:13](485)={
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
//                 "Hash": "f29452900c9e9bf5475fd7e589f1bd8714958f3d",
//                 "ObjectID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "LastObjectSize": "485",
//         "ModTime": "0",
//         "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "RefCount": "1"
//     }
// }
// c[0ffd45e074aa1b8df562907c95ad97526b7ca187:12](346)={
//     "ObjectInfo": {
//         "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "LastObjectSize": "346",
//         "ModTime": "0",
//         "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/archive/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "187c2d3f28c73395ee97bf19409147e94389ff96",
//             "ObjectID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:13"
//         }
//     }
// }
// c[0ffd45e074aa1b8df562907c95ad97526b7ca187:11](687)={
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
//                     "Hash": "82b9b0210cb8614983737976800b64daad781994",
//                     "ObjectID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:12"
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
//         "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "LastObjectSize": "687",
//         "ModTime": "0",
//         "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[0ffd45e074aa1b8df562907c95ad97526b7ca187:10](343)={
//     "ObjectInfo": {
//         "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "ac7d986ec6e1ed131a9838b1258062756c7c4078",
//             "ObjectID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//         }
//     }
// }
// u[0ffd45e074aa1b8df562907c95ad97526b7ca187:8](5)=
//     @@ -17,11 +17,11 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "0f0e314e82aa24001576643128a2ca8c88a694bc",
//                      "ObjectID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:9"
//                  }
//              },
//              {
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -31,7 +31,7 @@
//          "ObjectInfo": {
//              "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "LastObjectSize": "449",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//              "RefCount": "1"
//          }
// u[0ffd45e074aa1b8df562907c95ad97526b7ca187:9](135)=
//     @@ -7,13 +7,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/nt/avl/v0.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "0f72acf6f18bb025291bdefef9706e3917b96972",
//     +                    "ObjectID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//              "LastObjectSize": "255",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "0ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "RefCount": "1"
//          }
