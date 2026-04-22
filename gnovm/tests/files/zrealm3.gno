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

func main(cur realm,) {
	root = &Node{
		Key: "new",
	}
}

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:10](390)={
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
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "390",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:9](337)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "LastObjectSize": "337",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "23d24aa0b5df1c2e38adde7e684a4bc2af3d19e6",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "386",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e8c07f1269c2f82666c0c8cac8fad9c98776dc0d",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "07c6d63dc74da34ca1e7dea31877f38a9af49df5",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:7](-336)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:8](-389)
