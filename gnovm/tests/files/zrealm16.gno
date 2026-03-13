// PKGPATH: gno.land/r/test
package test

var (
	a  = &A{A: "here"}
	a2 = &A{A: "here"}
	b  [2]*B
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
	b[0].A = a2
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:15](5)=
//     @@ -13,7 +13,7 @@
//                      "Base": {
//                          "@type": "/gno.RefValue",
//                          "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -33,7 +33,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15",
//              "LastObjectSize": "444",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//     -        "RefCount": "1"
//     +        "ModTime": "17",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "340",
//     -        "ModTime": "15",
//     -        "RefCount": "3"
//     +        "ModTime": "17",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:6](8)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "LastObjectSize": "378",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "25af68699e6618f512440eee8a9b37a994931347",
//     +                "Escaped": true,
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:14](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14",
//              "LastObjectSize": "336",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "3c6ac6b66799a8816613ff8e2da5f6120f98678d",
//     +            "Hash": "c0abebbe29d5f091d88439645e6ed8deb902f8f3",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10](0)=
//     @@ -13,7 +13,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "c9f28b0a1b01eb5530d0fd9b3c7ca782aa29064e",
//     +                    "Hash": "e11eda824a4ed028305c300ddc9acf15c051e659",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14"
//                      },
//                      "Index": "0",
//     @@ -43,7 +43,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "583",
//     -        "ModTime": "13",
//     +        "ModTime": "17",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9](0)=
//     @@ -21,7 +21,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "655571ee53eabda24df8cd84d6acbce93141e624",
//     +            "Hash": "7bf3bb5e5b5203454045e870e5052f2d1b7046ae",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "42a79dfc7a1be8ad2239981db7f9b7417ab98879",
//     +                "Hash": "1ce0c2d8ac004045d97956a83b910e9b4c927280",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1d53138ea164a427121726fb40844462ad693ffa",
//     +                "Hash": "7e50c6bd50b04faa4875e1a0a6523524f9bad35e",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  }
//              },
