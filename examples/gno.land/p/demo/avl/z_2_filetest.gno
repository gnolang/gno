// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/demo/avl"
)

var tree avl.Tree

func init() {
	tree.Set("key0", "value0")
	tree.Set("key1", "value1")
}

func main(cur realm) {
	var updated bool
	updated = tree.Set("key2", "value2")
	println(updated, tree.Size())
}

// Output:
// false 3

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//              "LastObjectSize": "348",
//     -        "ModTime": "0",
//     +        "ModTime": "15",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//              "RefCount": "1"
//          },
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//              "LastObjectSize": "348",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//              "RefCount": "1"
//          },
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:19:0](549)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value2"
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:19:0",
//         "LastObjectSize": "549",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0](349)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0",
//         "LastObjectSize": "349",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c69f2ebb470214cab3ffa81de9f1ff0eed61c230",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:19:0"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0](759)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key2"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
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
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "e8b2fdd8fe2fb8295e79cff532606afefb56b596",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "61d82297f47ce157307e73cb76334afac888c228",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0",
//         "LastObjectSize": "759",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0](349)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0",
//         "LastObjectSize": "349",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "89a9769e36edabc3ea26594b42da664ed0197905",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0](759)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key1"
//             }
//         },
//         {},
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AwAAAAAAAAA=",
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
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "cf164d25b05e11db70fb5f3e90d37d182f935837",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "16acd90ab7f81aebbca1c12de6caf7694ed8cf97",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0",
//         "LastObjectSize": "759",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0](348)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0",
//         "LastObjectSize": "348",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3790ba8f901119782aa64514e58ad27c3c2267b7",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0](1)=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "b11293f3fdbc63efad34cac7138693147d7776d9",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0"
//     +                    "Hash": "0bb000f1a94bafe3860242af877b01d29594fe70",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0",
//              "LastObjectSize": "395",
//     -        "ModTime": "7",
//     +        "ModTime": "13",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "RefCount": "1"
//          }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0](-346)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0](-757)
