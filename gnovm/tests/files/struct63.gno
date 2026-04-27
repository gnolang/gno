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
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:11](216)={
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
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:11",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//         "RefCount": "1"
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](333)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/main.D"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "3eb2453a7fb8536b3ad3e3b9e837944addfd7a98",
//                 "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:11"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//         "LastObjectSize": "333",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//         "RefCount": "1"
//     }
// }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5](6)=
//     @@ -7,15 +7,15 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1274a7ff3e8fe80d0f715689b512ed03897e1a01",
//     -                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6"
//     +                "Hash": "e156e889763b28f60647501965a3f2613d3b67cc",
//     +                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "LastObjectSize": "331",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "RefCount": "1"
//          }
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](-331)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-214)
