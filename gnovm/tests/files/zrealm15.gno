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

func main(cur realm) {
	b[0] = nil
	b[1] = nil
	a.A = nil
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](-266)=
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
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14"
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
//     -                    "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5](-133)=
//     @@ -7,16 +7,6 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.S"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12](5)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "IsEscaped": true,
//              "LastObjectSize": "337",
//     -        "ModTime": "0",
//     -        "RefCount": "3"
//     +        "ModTime": "11",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "15",
//     -        "RefCount": "2"
//     +        "ModTime": "11",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "2934b1ba91d3fcb0bfe845af8fecab7fe559beb8",
//     +            "Hash": "a8e6dfcf3e8dc8d903cf5ccd0664cf494c8af4d7",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//              }
//          }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:14](-337)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:15](-585)
