// PKGPATH: gno.land/r/test
package test

// Verify that deleting a key and re-adding it with the same value results in
// the map being marked dirty (both the deletion and re-insertion are recorded)
// as opposed to the behaviour of zrealm_map5.
var m map[string]int

func init() {
	m = map[string]int{"a": 1, "b": 2}
}

func main(cur realm) {
	delete(m, "b")
	m["b"] = 2 // re-add with same value
	println("len:", len(m))
	println("a:", m["a"])
	println("b:", m["b"])
}

// Output:
// len: 2
// a: 1
// b: 2

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -44,7 +44,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "382",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "371",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "039b5d03839f565cb51c1d968502b3e6cf609abd",
//     +            "Hash": "14c86668b123a9f8a2b3994e0d9f136e542347e1",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "35298f967a98cf2ed38de879dc10ea43eafb009d",
//     +                "Hash": "f047c16fa6d540480414f8f68e491299d39c5e17",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
