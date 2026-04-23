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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](-260)=
//     @@ -8,15 +8,6 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              },
//              {
//     @@ -26,22 +17,13 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.B"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "582",
//     -        "ModTime": "11",
//     +        "ModTime": "15",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](-130)=
//     @@ -7,22 +7,13 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.S"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "382",
//     -        "ModTime": "11",
//     +        "ModTime": "15",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:12](5)=
//     @@ -3,8 +3,8 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "IsEscaped": true,
//              "LastObjectSize": "337",
//     -        "ModTime": "0",
//     -        "RefCount": "3"
//     +        "ModTime": "15",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "340",
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
//     -            "Hash": "11139afc54f70ddb4300ec4f8dbfa084876353c0",
//     +            "Hash": "90bf83c06410281c3afa2e7f3eab7f368d4de377",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//              }
//          }
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:14](-337)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:15](-579)
