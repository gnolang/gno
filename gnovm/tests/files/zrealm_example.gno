// PKGPATH: gno.land/r/example
package example

import (
	"gno.land/p/demo/dom"
)

var gPlot *dom.Plot

func init() {
	gPlot = &dom.Plot{Name: "First Plot"}
}

func main() {
	switchrealm()

	gPlot.AddPost("TEST_TITLE", "TEST_BODY")
	println(gPlot.String())
}

// Output:
// # [plot] First Plot
//
// ## TEST_TITLE
// TEST_BODY

// Realm:
// switchrealm["gno.land/r/example"]
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:12]={
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:11]={
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
//                 "Hash": "77fa0a02c9bf6d2ceb653733c5d3b8e096b8f419",
//                 "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:10]={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "37e8a2c1e11ffa853de0e4d88d38007040b1b919",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//         }
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:9]={
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
//                     "Hash": "12c2dde22875f18e829e62c7bd46ac6226373a20",
//                     "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:8]={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3bb997509749fd8aae09b2d6f3489cb5d70c18b3",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
//         }
//     }
// }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:7]=
//     @@ -7,12 +7,22 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/demo/avl.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "d36523c60c0aaee4fccb497a457095403b77e59b",
//     +                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:6",
//              "RefCount": "1"
//          }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:6]=
//     @@ -17,11 +17,12 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f0e284183fac3e381b521c6c29a8a0c51bec1f72",
//     +                "Hash": "8bbaca14feedd6a41cfea0af0deefc1f98c713b7",
//                      "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7"
//                  }
//              },
//              {
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -30,7 +31,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:6",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:5",
//              "RefCount": "1"
//          }
