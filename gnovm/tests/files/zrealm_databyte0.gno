// PKGPATH: gno.land/r/test
package test

// Verify that plain byte-index assignment (bs[i] = v) on a Data-backed []byte
// realm variable marks the backing ArrayValue dirty via DidUpdate, so the
// mutation persists across transactions.
var bs = []byte("aabb")

func main(cur realm) {
	bs[0] = 'A'
	bs[1] = 'B'
	println(string(bs))
}

// Output:
// ABbb

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "YWFiYg==",
//     +    "Data": "QUJiYg==",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "181",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "369",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "bb3aef6fabbc2581ab64c2544349179712b17180",
//     +                "Hash": "29b0d5cb086622b4dc2491343b970d1da9982304",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  },
//                  "Length": "4",
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "537",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "540bc520c52d33138aec2717bfe0428b8af4ad93",
//     +                "Hash": "1007b3bcb38cc5ef099e12bf912ca885d5e819b8",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
