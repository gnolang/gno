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
	x.b.c = &C{}
	x.b.c.d = &D{e: 2}
}

// Realm:
// finalizerealm["gno.land/r/main"]
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:16](216)={
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
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:16",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:15",
//         "RefCount": "1"
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:15](336)={
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:15",
//         "LastObjectSize": "336",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:14",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/main.D"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "1dec28e121657267508856fefcbbcd7db14d7549",
//             "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:16"
//         }
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:14](379)={
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
//                     "Hash": "2ad35f30ae1c285d1778656149738d281a39dd84",
//                     "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:15"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:14",
//         "LastObjectSize": "379",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:13",
//         "RefCount": "1"
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:13](335)={
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:13",
//         "LastObjectSize": "335",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/main.C"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d60bdce59d90096915591cbee7506e6798bde04d",
//             "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:14"
//         }
//     }
// }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](6)=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "fea65b1c3974e6eb29ccc3de96f30b47f5fe41c8",
//     -                    "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:7"
//     +                    "Hash": "665b2ff7bc854d22d0c126ceb5b4d3a44045d54a",
//     +                    "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:13"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "RefCount": "1"
//          }
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-333)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-376)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:9](-334)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](-215)
