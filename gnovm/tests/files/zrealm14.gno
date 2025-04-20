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

func main() {
	crossing()

	b[0] = nil
	b[1] = nil
}

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]=
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
//     @@ -27,22 +17,12 @@
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
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     -        "ModTime": "10",
//     +        "ModTime": "14",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//     -        "ModTime": "12",
//     -        "RefCount": "3"
//     +        "ModTime": "14",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:12]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:13]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:14]
