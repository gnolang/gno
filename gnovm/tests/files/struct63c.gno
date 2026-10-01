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
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "1b3c7d6e4cf175a2c9d623ac6f1dd98f633b4abb",
//     +            "Hash": "f51fb2944960939308ef90fde54e3f0aac191465",
//                  "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6"
//              }
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](5)=
//     @@ -12,7 +12,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "01365894281944ddba4ffafa4f0c598daa8bb4a9",
//     +                    "Hash": "1aa2ba7625255766d8311c0b67bc39957fc4bbc3",
//                          "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5"
//                      },
//                      "Index": "0",
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "23f351660e1b2a98a14e916a9d0a34d3586faa2a",
//     +            "Hash": "31973ad1121379fc9f49b5757adbeba2c46f938c",
//                  "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//              }
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "IsEscaped": true,
//              "LastObjectSize": "840",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -82,7 +82,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8a9f6565383dc7cc1d8c74d1754426ae9d6cbb64",
//     +                "Hash": "86e1459e1fb16c006b63164fce67a01f13ae6051",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-333)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-376)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:9](-334)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](-215)
