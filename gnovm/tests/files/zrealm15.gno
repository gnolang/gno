// PKGPATH: gno.land/r/test
package test

var (
	a = &A{}
	b [2]*B
	s *S
)

type S struct {
	S string
}
type A struct {
	A *S
}
type B struct {
	A *A
	B *S
}

func init() {
	s = &S{
		S: "c",
	}
	c := B{
		A: a,
		B: s,
	}
	b[0] = &c
	b[1] = &c
	a.A = s
}

func main() {
	switchrealm()

	b[0] = nil
	b[1] = nil
	a.A = nil
}

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]=
//     @@ -8,16 +8,6 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
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
//     -                    "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//     -        "ModTime": "6",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "2"
//     +        "ModTime": "10",
//     +        "RefCount": "0"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]=
//     @@ -7,22 +7,12 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.S"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     -        "ModTime": "6",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "3"
//     +        "ModTime": "10",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     @@ -3,7 +3,7 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "IsEscaped": true,
//              "ModTime": "10",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "38707548446d3e67a38939d1af16cbc2f3a311f9",
//     +            "Hash": "e82e410f22425e48d5f6c611160084a4dd50d3d1",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]
