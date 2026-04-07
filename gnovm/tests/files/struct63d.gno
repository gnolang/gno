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
//                     "Hash": "6762ea9b9f3263aaf4a4a2c665ec38779a323eec",
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
//             "Hash": "4d1ddbecab6e5585cbff3b0348668ee36077380b",
//             "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:14"
//         }
//     }
// }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](6)=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "bd0670aa7dc0c1a702fe46f743714d7b9d0072a0",
//     -                    "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:7"
//     +                    "Hash": "c4855bdb25859f250a4405d0d7011ce16e4f7ea5",
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
