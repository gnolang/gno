// PKGPATH: gno.land/r/test
package test

var root Node

type (
	Node any
	Key  any
)

type InnerNode struct {
	Key   Key
	Left  Node
	Right Node
}

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main(cur realm,) {
	root = InnerNode{
		Key: "new",
	}
}

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0](269)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "new"
//             }
//         },
//         {},
//         {}
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//         "LastObjectSize": "269",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "352",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "20595725486a167acb56adeee0173fb18e40cd33",
//     -            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +            "Hash": "6bd63c265cd4fcc626104192cbf474e32fc79af4",
//     +            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0"
//              }
//          }
//      }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](-269)
