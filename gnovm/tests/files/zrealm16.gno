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
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "IsEscaped": true,
//              "LastObjectSize": "299",
//     -        "ModTime": "15",
//     -        "RefCount": "3"
//     +        "ModTime": "17",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
