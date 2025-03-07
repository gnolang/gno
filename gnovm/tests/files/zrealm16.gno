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

func main() {
	b[0].A = a2
}

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]=
//     --- old
//     +++ new
//     @@ -13,7 +13,7 @@
//                      "Base": {
//                          "@type": "/gno.RefValue",
//                          "Escaped": true,
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -32,7 +32,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]=
//     --- old
//     +++ new
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "11",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     --- old
//     +++ new
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "IsEscaped": true,
//     -        "ModTime": "10",
//     -        "RefCount": "3"
//     +        "ModTime": "11",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
