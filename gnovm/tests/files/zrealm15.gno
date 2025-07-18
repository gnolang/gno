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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](-270)=
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
//     @@ -27,23 +17,13 @@
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
//              "LastObjectSize": "596",
//     -        "ModTime": "11",
//     +        "ModTime": "15",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5:0](-135)=
//     @@ -7,23 +7,13 @@
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
//              "LastObjectSize": "391",
//     -        "ModTime": "11",
//     +        "ModTime": "15",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0](5)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//              "IsEscaped": true,
//              "LastObjectSize": "343",
//     -        "ModTime": "0",
//     -        "RefCount": "3"
//     +        "ModTime": "15",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0](0)=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "346",
//              "ModTime": "15",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "f6a97f2e653fdc2528aa855498d6741b601ada38",
//     +            "Hash": "78bf623f3ae3aa6030817832c6e3376f0acdda75",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5:0"
//              }
//          }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0](-343)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:15:0](-593)
