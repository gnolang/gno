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

func main(cur realm,) {
	b[0] = nil
	b[1] = nil
	a.A = nil
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0]=
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
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0"
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
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//     -        "ModTime": "11",
//     +        "ModTime": "15",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5:0]=
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
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5:0",
//     -        "ModTime": "11",
//     +        "ModTime": "15",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0]=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "3"
//     +        "ModTime": "15",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0]=
//     @@ -3,7 +3,7 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0",
//              "IsEscaped": true,
//              "ModTime": "15",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "d841ecacc34604f3d3773824606e595aa9cfd396",
//     +            "Hash": "9cb1711da458ebe78ceaefcfe08a345fe4fb67c1",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5:0"
//              }
//          }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0]
