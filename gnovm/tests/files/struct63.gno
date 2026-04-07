// PKGPATH: gno.land/r/main
package main

type A struct {
	b B
}

type B struct {
	c C
}

type C struct {
	d D
}

type D struct {
	e int
}

var x = A{b: B{c: C{d: D{e: 1}}}}

func main() {
	x.b.c = C{d: D{e: 2}}
}

// Realm:
// finalizerealm["gno.land/r/main"]
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:11](216)={
//     "Fields": [
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:11",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](333)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/main.D"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "9ecaaba55c8400d0a9b7f90804afeebeccd80a79",
//                 "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:11"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//         "LastObjectSize": "333",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//         "RefCount": "1"
//     }
// }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5](6)=
//     @@ -7,15 +7,15 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1fd36417177aea9a1c497135645769b1228a43ee",
//     -                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6"
//     +                "Hash": "4f2e1ce0434fbc3f3dd49c10c8823bed0c407ab4",
//     +                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "LastObjectSize": "331",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "RefCount": "1"
//          }
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](-331)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-214)
