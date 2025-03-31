// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key0"), "value0-new")
	println(updated, tree.Size())
}

// Output:
// true 1

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5](6)=
//     @@ -17,7 +17,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "value0"
//     +                "value": "value0-new"
//                  }
//              },
//              {
//     @@ -48,7 +48,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](2)=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b819bc3b7c351c8e571654d4ce054a997a3ac50f",
//     +            "Hash": "d57ee28f5030eb8c612b374155fd545675633288",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//     -        "ModTime": "3",
//     +        "ModTime": "5",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -30,7 +30,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "7b9d58f40430bbbcbafd47eefb7a6dd342477f71",
//     +                    "Hash": "ef474e536e2d4229156a07798a88edc3767c0896",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                      },
//                      "Index": "0",
