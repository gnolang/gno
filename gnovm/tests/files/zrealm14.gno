// PKGPATH: gno.land/r/test
package test

var (
	a = &A{A: "here"}
	b [2]*B
)

type A struct {
	A string
}
type B struct {
	A *A
	B string
}

func init() {
	c := B{
		A: a,
		B: "c",
	}
	b[0] = &c

	d := B{
		A: a,
		B: "d",
	}
	b[1] = &d
}

func main(cur realm,) {
	b[0] = nil
	b[1] = nil
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](-260)=
//     @@ -8,16 +8,6 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "039c758da189de11bb723e68a3ef3f2f59bae535",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              },
//              {
//     @@ -27,23 +17,13 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "054c315a527d298c5e5919d78ced1705080b6fc3",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "582",
//     -        "ModTime": "10",
//     +        "ModTime": "14",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "12",
//     -        "RefCount": "3"
//     +        "ModTime": "14",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:6](0)=
//     @@ -21,7 +21,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "e967d26f7dc6cad11f3a7cc86c56cab106be8637",
//     +            "Hash": "ec97cf562f62097dfe3d3c9c5e53163d52888d9d",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f73e3304fbcaae9aa841c88c54240e25bd5f5279",
//     +                "Hash": "4d105ed1ed2d08d7e76e7a79274120f1e56afe9a",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:11](-335)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:12](-441)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:13](-335)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:14](-441)
