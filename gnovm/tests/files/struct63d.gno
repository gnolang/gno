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
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:9](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:9",
//              "LastObjectSize": "334",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:8",
//              "RefCount": "1"
//          },
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14](378)={
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
//                     "Hash": "89bab966d4e4111e9325ae403d075de841a28dae",
//                     "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:9"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14",
//         "LastObjectSize": "378",
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
//             "Hash": "d58b6674054f460c7e570874230df8c19c4466b2",
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
//     +                    "Hash": "f91152997d3d07388ed6d2316ad143fdf16d4e71",
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
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "426adea303d4bbb2fa4f82a8ccec2d916ffd67df",
//     +            "Hash": "ac6b460a44cdadca951ba2298933c6441647daa9",
//                  "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6"
//              }
//          }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](5)=
//     @@ -12,7 +12,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "77ea9b0c2872ef550293b6708d5964b55db2c8fd",
//     +                    "Hash": "b37d01ff993d5d89b34915eff7266482d6ead693",
//                          "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5"
//                      },
//                      "Index": "0",
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7809bf2475549e40c85f533ff30a888ebe446b48",
//     +            "Hash": "6f6d2d8f2bc68541f9e181127d96d0e175e98e97",
//                  "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//              }
//          }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1253",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -219,7 +219,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "22a5909968065d0f398b10fdae7563e235b45ef2",
//     +                "Hash": "82fc6a886f60c778e112709d20fe3bf665584458",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-333)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-376)
