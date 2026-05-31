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
	d := x.b.c.d
	x.b.c = &C{d: d}
}

// Realm:
// finalizerealm["gno.land/r/main"]
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:9](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:9",
//              "LastObjectSize": "334",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:8",
//              "RefCount": "1"
//          },
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:14](378)={
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
//                     "Hash": "9dd8facc548ca98b55fa87afd0dad23b1074fda4",
//                     "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:9"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:14",
//         "LastObjectSize": "378",
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
//             "Hash": "5d2946fdd983d5a633a9caf417c769828193c278",
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
//     +                    "Hash": "58efd07df8536f0a82758135f5012fc916a2707f",
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
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "1b3c7d6e4cf175a2c9d623ac6f1dd98f633b4abb",
//     +            "Hash": "07855c9edc4df6a8b0a04a05b3b735dcbfad974e",
//                  "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6"
//              }
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](5)=
//     @@ -12,7 +12,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "01365894281944ddba4ffafa4f0c598daa8bb4a9",
//     +                    "Hash": "59688b2b118713a6469e3497a9eaa0709d68afc8",
//                          "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5"
//                      },
//                      "Index": "0",
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "23f351660e1b2a98a14e916a9d0a34d3586faa2a",
//     +            "Hash": "fd7d115b0a891d0d6ac2ca95ad92c91c05fd89ca",
//                  "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//              }
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "IsEscaped": true,
//              "LastObjectSize": "840",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -82,7 +82,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8a9f6565383dc7cc1d8c74d1754426ae9d6cbb64",
//     +                "Hash": "f18a61432c8e566cfecc6a7fe397cd5835fe36f8",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-333)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-376)
