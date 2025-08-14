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
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:11](256)={
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "LastObjectSize": "256",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:10](483)={
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
//                 "Hash": "f190df54e397e2006cee3fc525bcc1b4d556e4c4",
//                 "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "LastObjectSize": "483",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:9](341)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "LastObjectSize": "341",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a74fad6da10f1cec74ad3a8751490b4dca957761",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//         }
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:8](679)={
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
//                     "Hash": "8b11b3d07ddeb034f70a114c9433ec6bd5cbf899",
//                     "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//         "LastObjectSize": "679",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:7](340)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//         "LastObjectSize": "340",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "869abdac30a3ae78b2191806e1c894c48e399122",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8"
//         }
//     }
// }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:9](134)=
//     @@ -1,20 +1,30 @@
//      {
//     +    "Fields": [
//     +        {
//     +            "T": {
//     +                "@type": "/gno.PointerType",
//     +                "Elt": {
//     +                    "@type": "/gno.RefType",
//     +                    "ID": "gno.land/p/demo/avl.Node"
//     +                }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "f339f5f74cb0636a0eb703d5e28ac42e0b5df9c0",
//     +                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//     +            }
//     +        }
//     +    ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//     -        "LastObjectSize": "341",
//     -        "ModTime": "0",
//     +        "LastObjectSize": "254",
//     +        "ModTime": "6",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//              "RefCount": "1"
//     -    },
//     -    "Value": {
//     -        "T": {
//     -            "@type": "/gno.RefType",
//     -            "ID": "gno.land/p/demo/dom.Post"
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "a74fad6da10f1cec74ad3a8751490b4dca957761",
//     -            "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//     -        }
//          }
//      }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:8](5)=
//     @@ -7,32 +7,18 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "0"
//     +                "value": "First Plot"
//                  }
//              },
//              {
//                  "T": {
//     -                "@type": "/gno.PointerType",
//     -                "Elt": {
//     -                    "@type": "/gno.RefType",
//     -                    "ID": "gno.land/p/demo/dom.Post"
//     -                }
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/p/demo/avl.Tree"
//                  },
//                  "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "8b11b3d07ddeb034f70a114c9433ec6bd5cbf899",
//     -                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//     -            }
//     -        },
//     -        {
//     -            "T": {
//     -                "@type": "/gno.PrimitiveType",
//     -                "value": "64"
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "1b4c3cd50e265edc976073767beed2287bb9bf75",
//     +                "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9"
//                  }
//              },
//              {
//     @@ -41,30 +27,12 @@
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//                  }
//     -        },
//     -        {
//     -            "T": {
//     -                "@type": "/gno.PointerType",
//     -                "Elt": {
//     -                    "@type": "/gno.RefType",
//     -                    "ID": "gno.land/p/demo/avl.Node"
//     -                }
//     -            }
//     -        },
//     -        {
//     -            "T": {
//     -                "@type": "/gno.PointerType",
//     -                "Elt": {
//     -                    "@type": "/gno.RefType",
//     -                    "ID": "gno.land/p/demo/avl.Node"
//     -                }
//     -            }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:8",
//     -        "LastObjectSize": "679",
//     -        "ModTime": "0",
//     +        "LastObjectSize": "448",
//     +        "ModTime": "6",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//              "RefCount": "1"
//          }
