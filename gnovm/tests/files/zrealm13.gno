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
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](-130)=
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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "12",
//     -        "RefCount": "3"
//     +        "ModTime": "10",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:11](-335)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:12](-444)
