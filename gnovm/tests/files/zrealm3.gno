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
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0](394)={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//         "LastObjectSize": "394",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0](343)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "b4d2041543476f76c7a169a63fd6ebcae2af731e",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "393",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e29105cd5ba238194723643a08c3b265d26a5f4c",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +                "Hash": "23cefd5fff29e9e131aa052c8318ccbdcb64f1d4",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
//                  },
//                  "Index": "0",
//                  "TV": null
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](-342)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0](-393)
