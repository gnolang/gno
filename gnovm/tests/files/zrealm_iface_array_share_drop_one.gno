// PKGPATH: gno.land/r/test
package test

// Companion to zrealm_iface_array_share.gno. The element shared by both
// arrays is persisted escaped with RefCount 2; dropping one of the two
// references must decrement to 1 and leave the object live, so the
// surviving array still reads it.

type S struct{ F int }

var (
	arr  [1]any
	arr2 [1]any
)

func init() {
	arr[0] = S{1}
	arr2 = arr // shared -> escaped, RefCount 2
}

func main(cur realm) {
	arr[0] = nil
	println("kept:", arr2[0].(S).F)
}

// Output:
// kept: 1

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](-144)=
//     @@ -1,21 +1,12 @@
//      {
//          "Data": null,
//          "List": [
//     -        {
//     -            "T": {
//     -                "@type": "/gno.RefType",
//     -                "ID": "gno.land/r/test.S"
//     -            },
//     -            "V": {
//     -                "@type": "/gno.RefValue",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//     -            }
//     -        }
//     +        {}
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "336",
//     -        "ModTime": "9",
//     +        "ModTime": "11",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](5)=
//     @@ -12,7 +12,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "IsEscaped": true,
//              "LastObjectSize": "217",
//     -        "ModTime": "0",
//     -        "RefCount": "2"
//     +        "ModTime": "11",
//     +        "RefCount": "1"
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "239853dee18a20c6db92f285fc47c690b2d2a24c",
//     +            "Hash": "10628b2e62a8ed820a76d90fcecd68f8212f499c",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d38e6c199d2d1521e8e683489efe9dde11d6d9b5",
//     +                "Hash": "ded75eda30081e329e6225387699b40c90280d2a",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
