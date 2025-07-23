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
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:14](256)={
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14",
//         "LastObjectSize": "256",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:13](484)={
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
//                 "Hash": "5de3e7ff11ff5edc7f205674996c7c47a0c45029",
//                 "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:13",
//         "LastObjectSize": "484",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:12](343)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:12",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/dom.Post"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "98e91a75929fc778d9f0f62886abd62f9152de1a",
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
//                     "ID": "gno.land/p/demo/dom.Post"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3c03811f000dfdd7e40014cc745f93b22e29d4b8",
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
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11",
//         "LastObjectSize": "682",
//         "ModTime": "0",
//         "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "RefCount": "1"
//     }
// }
// c[1ffd45e074aa1b8df562907c95ad97526b7ca187:10](342)={
//     "ObjectInfo": {
//         "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10",
//         "LastObjectSize": "342",
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
//             "Hash": "8e780dfd5043f7bd20a8898f2447cc1f130d436a",
//             "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:11"
//         }
//     }
// }
// u[1ffd45e074aa1b8df562907c95ad97526b7ca187:9](135)=
//     @@ -7,13 +7,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/p/demo/avl.Node"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "cdd32e91aaf44002f38742d156dc218b01f96341",
//     +                    "ObjectID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:10"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:9",
//              "LastObjectSize": "254",
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
//     -                "Hash": "26d848664a18b9656977e1a343672fa667890bac",
//     +                "Hash": "90966262bf536f809e84456091dcb7e547af4017",
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
//              "LastObjectSize": "448",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "1ffd45e074aa1b8df562907c95ad97526b7ca187:7",
//              "RefCount": "1"
//          }
