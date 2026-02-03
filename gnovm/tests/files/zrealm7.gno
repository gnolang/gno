// PKGPATH: gno.land/r/test
package test

import (
	"filetests/extern/timtadh/data_structures/tree/avl"
	"filetests/extern/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
	tree, _ = tree.Put(types.String("key2"), "value2")
}

func main(cur realm) {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 4

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:14](619)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "filetests/extern/timtadh/data_structures/types.String"
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
//                     "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "LastObjectSize": "619",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:13](376)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "LastObjectSize": "376",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d23620f4afd5ea68c8dcea8c04faaa6500cd8043",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12](135)=
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AgAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,13 +43,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "3b11426288ff1c536b4f7004debfbec94cbdd5c6",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "LastObjectSize": "619",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10](134)=
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
//     -                    "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//     -                }
//                  }
//              },
//              {
//     @@ -48,7 +39,24 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "db040c58097327c7918fdc57b6cec8f49230814b",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
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
//     +                    "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//     +                }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                      },
//                      "Index": "0",
//     @@ -59,7 +67,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "748",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:11](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "4a17484cc93c655365dd3aca8dd7b71f98efefc6",
//     +            "Hash": "0e5376edc0cabc86b08996b390e8280c38346b09",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](-124)=
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AwAAAAAAAAA=",
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,23 +43,13 @@
//                          "@type": "/gno.RefType",
//                          "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "c188a4b77bfb4d65e43b0a7cdc52d654aca701a2",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "746",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "LastObjectSize": "374",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b65adb52a839f07bf1db12b5889d4bff183f5134",
//     +            "Hash": "7d47756e4f866b266bbc0bf2013d6356c93db10d",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "373",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "78a0b6e28ada4afb447c17ffd7a956a7139793f7",
//     +            "Hash": "1059095280e98f907857fc5204f45e523eb27912",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "423",
//     -        "ModTime": "6",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b171dea0cd981edc859f2b1dea69ad0137f1c227",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "e096c094d6d895e5feee9e878292f88e2977fe54",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
