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
	switchrealm()

	b = a
}

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//     -        "ModTime": "3",
//     +        "ModTime": "6",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -62,7 +62,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "4f679d9dc2ea37d8f31b020193e7afae0236c97c",
//     +                    "Escaped": true,
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                      },
//                      "Index": "0",
//     @@ -76,6 +76,16 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.S"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Escaped": true,
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              },
//              {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
