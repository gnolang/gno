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
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](5)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ad6d67fc38688886bc13c82a4bcb72b393c1d421",
//     +                "Hash": "ae8c8039711946765c5fa1e2eab697541cfd55b3",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:5"
//                  }
//              }
//     @@ -15,7 +15,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "331",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "5d3293bb598b65ca28e667e5d922bfc7f00c366d",
//     +            "Hash": "c7448fa6feb948e89835621349b4f213706be2bc",
//                  "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//              }
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "IsEscaped": true,
//              "LastObjectSize": "839",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -82,7 +82,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2faca1042190334ef66849c79b093880f82a3e10",
//     +                "Hash": "ae6accc34ac47c25864a90e84cc906e0c9c9ab20",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](-331)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-214)
