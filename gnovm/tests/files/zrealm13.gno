// PKGPATH: gno.land/r/test
package test

var (
	a = &A{A: "here"}
	b [2]*B
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

func main() {
	b[0] = nil
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
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              },
//              {
//     @@ -42,7 +32,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//     -        "ModTime": "6",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]=
//     @@ -3,8 +3,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "2"
//     +        "ModTime": "9",
//     +        "RefCount": "1"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
