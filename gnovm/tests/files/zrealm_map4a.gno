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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -44,7 +44,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "382",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "371",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "10c87852bc4e2427f6befea3083f84892cf17a03",
//     +            "Hash": "d398ee1392cd3908cea77490c21101fcafa419e7",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3b77cd040e71c669066b705edd76c01afbb365e0",
//     +                "Hash": "05695299dd581960d708c14d5bfad9781e965b22",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
