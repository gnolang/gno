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
	d.e = 5
	x.b.c = &C{d: d}
}

// Realm:
// finalizerealm["gno.land/r/main"]
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](5)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "BQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -11,7 +11,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//              "LastObjectSize": "215",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:9",
//              "RefCount": "1"
//          }
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
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "f82b64740b75696af2804084c1c6553f542695af",
//     +            "Hash": "9adf50a17d5ecd43f48caaaa607daaf8654e2c72",
//                  "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:10"
//              }
//          }
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
//                     "Hash": "f4783626fc250fa516e5f3ff76df78641f2f37b2",
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
//             "Hash": "35c946879944ec292797680b43ec72253f2efc77",
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
//     +                    "Hash": "d19c41893833f71d630976276bd1ebd6a53414f3",
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
