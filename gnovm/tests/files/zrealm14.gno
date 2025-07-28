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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](-260)=
//     @@ -8,16 +8,6 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "ff7ee2df0e65a933be3dce7550bf361d29e83bc0",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              },
//              {
//     @@ -27,16 +17,6 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "ff020e571a6e197bf9ed4e90057ba19a576e0622",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "12",
//     -        "RefCount": "3"
//     +        "ModTime": "10",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:11](-335)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:12](-444)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:13](-335)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:14](-444)
