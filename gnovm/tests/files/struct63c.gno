// PKGPATH: gno.land/r/main
package main

type A struct {
	b *B
}

type B struct {
	c *C
}

type C struct {
	d *D
}

type D struct {
	e int
}

var x = A{b: &B{c: &C{d: &D{e: 1}}}}

func main() {
	x.b.c = &C{d: &D{e: 2}}
}

// Realm:
// finalizerealm["gno.land/r/main"]
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:16](216)={
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
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:16",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:15",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:15](336)={
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:15",
//         "LastObjectSize": "336",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/main.D"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "db794299f99e46ec060ee261f9f37361d4cef0dd",
//             "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:16"
//         }
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14](379)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/main.D"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "2380c245d37e19f8c36e3618507be17822009e95",
//                     "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:15"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14",
//         "LastObjectSize": "379",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:13",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:13](335)={
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:13",
//         "LastObjectSize": "335",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/main.C"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5a14c88cc35094c1a4a1da28e8687d5a315bfa3c",
//             "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14"
//         }
//     }
// }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](6)=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "e9ab02374a9aeeec16cf1716ec13fa88690ebe33",
//     -                    "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:7"
//     +                    "Hash": "46af1d33174f442a16702bf0610d934e408b5f62",
//     +                    "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:13"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "RefCount": "1"
//          }
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-333)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-376)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:9](-334)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](-215)
