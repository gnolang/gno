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

func main() {
	crossing()

	b = a
}

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -12,6 +12,16 @@
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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
