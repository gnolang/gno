// PKGPATH: gno.land/r/test
package test

var root *Node

type Key any

type Node struct {
	Key   Key
	Left  *Node
	Right *Node
}

func init() {
	root = &Node{
		Key: "old",
	}
}

func main() {
	crossing()

	root = &Node{
		Key: "new",
	}
}

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]={
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
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "24da1ffa2b1135d506e4e79ad56ba790d5bb3d36",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -17,8 +17,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "477e1546db4d3a70091c73a992039811bc2666fe",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "d8185ab144b70c0fec8cbda653cabee29619a499",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]
