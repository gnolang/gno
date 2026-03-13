// PKGPATH: gno.land/r/test
package test

type S struct {
	A *int
}

var a, b *S

func init() {
	a = new(S)
	a.A = new(int)
	*a.A = 4
}

func main(cur realm,) {
	b = a
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "249",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.S"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//     +        "IsEscaped": true,
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     -        "RefCount": "1"
//     +        "ModTime": "10",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](3)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "383",
//     -        "ModTime": "7",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "0be689206f8a36b068b0ebaeaa8bf78052dc953c",
//     +                "Escaped": true,
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                  },
//                  "Index": "0",
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -81,7 +81,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1670513d52f692b73834d7a093b763b52e2e1c47",
//     +                "Hash": "2fd740361b49a08ef5710d5bb21893aa6a5e40ba",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -91,7 +91,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "da4c2d62873a0c7fa1f09f7bc5da0eed04a239d4",
//     +                "Hash": "b503587dcf988b27fbdf9d1a25da1ccd2b230fab",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
