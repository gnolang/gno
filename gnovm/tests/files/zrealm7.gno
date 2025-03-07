// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
	tree, _ = tree.Put(types.String("key2"), "value2")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 4

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/types.String"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key3"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value3"
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
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "42cd813e173ad23c7873e9605901e8bea1176c96",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]=
//     --- old
//     +++ new
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AgAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,12 +43,22 @@
//                          "@type": "/gno.RefType",
//                          "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "4f88fcdc73a4a94905e8e4044aa50c2ec7bf2227",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]=
//     --- old
//     +++ new
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AwAAAAAAAAA=",
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,22 +43,12 @@
//                          "@type": "/gno.RefType",
//                          "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "e47c9fe5fa842d7ec1bbbc99317cb157a4174877",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]=
//     --- old
//     +++ new
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "916668cd920db8d3551f75ba22194d2e1e1cc804",
//     +            "Hash": "d7fbb234dca9f194f35fe5409a62db9daf39b0fc",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]=
//     --- old
//     +++ new
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "17ce788733dc5f31545164a039c5eee6b30e3b25",
//     +            "Hash": "2c172bbe0183ccc73c59d9acb196c45b0331c39e",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]=
//     --- old
//     +++ new
//     @@ -21,19 +21,10 @@
//                  }
//              },
//              {
//     -            "N": "AgAAAAAAAAA=",
//     +            "N": "AwAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     -            }
//     -        },
//     -        {
//     -            "T": {
//     -                "@type": "/gno.PointerType",
//     -                "Elt": {
//     -                    "@type": "/gno.RefType",
//     -                    "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//     -                }
//                  }
//              },
//              {
//     @@ -48,7 +39,26 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "750e6705842b4af3b3602ccaaca3d089f8402679",
//     +                    "Hash": "76a40dcf03d32c312c2213265c14d4de1b12a810",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//     +            }
//     +        },
//     +        {
//     +            "T": {
//     +                "@type": "/gno.PointerType",
//     +                "Elt": {
//     +                    "@type": "/gno.RefType",
//     +                    "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//     +                }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "43f69f24b7827a331921b4af0f667346d186e0c3",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                      },
//                      "Index": "0",
//     @@ -58,7 +68,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:6]=
//     --- old
//     +++ new
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "71038fd71391c2de4859dedffaaf6feae2282dae",
//     +            "Hash": "92b2f4ebab764951f64086bce480f898f755de5a",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]=
//     --- old
//     +++ new
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//     -        "ModTime": "3",
//     +        "ModTime": "9",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -30,8 +30,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "a4fa9bdf45caf8c6b5be7a3752704423817b3ef2",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                    "Hash": "a303dd292d2104ce5461b1e795b9855540fbe179",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                      },
//                      "Index": "0",
//                      "TV": null
