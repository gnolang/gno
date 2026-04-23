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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "249",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,15 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.S"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:8](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//     +        "IsEscaped": true,
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     -        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     -        "RefCount": "1"
//     +        "ModTime": "10",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "383",
//     -        "ModTime": "7",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2a8e4e6d1fc5cad9a7f23c526cb5ffb097b8c493",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                  },
//                  "Index": "0",
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "50c6f1f4029c4dde84a232f7c968487683f3a0f4",
//     +                "Hash": "c331f7a0d093cc1d6539a5dd81195c33d08ff01b",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "81e0c6ca227ff564c90d2b560c954fe5543c9769",
//     +                "Hash": "e2f95c616ca1189d7e775fa87c98dfba951bb32f",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
