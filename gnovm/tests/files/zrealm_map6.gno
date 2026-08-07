// PKGPATH: gno.land/r/test
package test

// A declared key type over a primitive — the map[address]T shape. The map is
// built in init(), committed, then read and mutated in main() against a
// reloaded MapValue whose vmap is built on demand with the prefix omitted.

type Address string

var m map[Address]int

func init() {
	m = map[Address]int{"g1alice": 1, "g1bob": 2}
}

func main(cur realm) {
	println("alice:", m["g1alice"])
	m["g1carol"] = 3
	delete(m, "g1bob")
	println("len:", len(m))
	println("carol:", m["g1carol"])
	_, ok := m["g1bob"]
	println("bob gone:", !ok)
}

// Output:
// alice: 1
// len: 2
// carol: 3
// bob gone: true

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](7)=
//     @@ -28,11 +28,11 @@
//                          },
//                          "V": {
//                              "@type": "/gno.StringValue",
//     -                        "value": "g1bob"
//     +                        "value": "g1carol"
//                          }
//                      },
//                      "Value": {
//     -                    "N": "AgAAAAAAAAA=",
//     +                    "N": "AwAAAAAAAAA=",
//                          "T": {
//                              "@type": "/gno.PrimitiveType",
//                              "value": "32"
//     @@ -44,7 +44,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "426",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "388",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7e3199ea0453ec48a8490a3c73f5c016c614b9f1",
//     +            "Hash": "94db753ae151ce869a317eda0be15f0dde7e723f",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d154827cd28ca1612fa7e1eebedadd51c0176f6d",
//     +                "Hash": "bbc681a73ee32213463c91a5ca50433510532240",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
