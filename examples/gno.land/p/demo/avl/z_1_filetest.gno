// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/demo/avl"
)

var node *avl.Node

func init() {
	node = avl.NewNode("key0", "value0")
	node, _ = node.Set("key1", "value1")
}

func main(cur realm) {
	var updated bool
	node, updated = node.Set("key2", "value2")
	// println(node, updated)
	println(updated, node.Size())
}

// Output:
// false 3

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//              "RefCount": "1"
//          },
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//              "RefCount": "1"
//          },
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "8858fa597cb8f7361adb30a2b5e6ee98ed4ff2c2",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:18:0"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0]={
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
//                     "Hash": "e220aeff9de378e09cc8f57341773803667a7fb7",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0"
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
//                     "Hash": "d4c41b05ec95727c10a47c6c40124050d5ce8ffb",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:17:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "7aef739ee9b7a445979ce3a32f64126702fec474",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:16:0"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0]={
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
//                     "Hash": "597fe00715d44e6b9a8b685ebaa5d1a5569afc76",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
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
//                     "Hash": "1c7a9466f59430336efc4647412ef56cd626e408",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "4e062e752d1ba68f4803dc444c8da5fb13bef73a",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//     -        "ModTime": "6",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -17,8 +17,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f361ff031186372d3e18ba1ad715e6c89a976c22",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +                "Hash": "906b35baa325f1024f809bbcfa80d593c04910bc",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0"
//                  },
//                  "Index": "0",
//                  "TV": null
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0]
