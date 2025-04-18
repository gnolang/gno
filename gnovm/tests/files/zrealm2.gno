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

func main() {
	switchrealm()

	root = InnerNode{
		Key: "new",
	}
}

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//     -        "ModTime": "3",
//     +        "ModTime": "4",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -25,8 +25,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d9cf719bd2589ae934ec781c5ede54ca30ed2189",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                "Hash": "61d4aa77a87c01e07038c6030d6aca299d0fdc1b",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
//              {
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]
