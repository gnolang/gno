// PKGPATH: gno.land/r/test
package test

var root any

func main(cur realm) {
	println(root)
	root = 1
	println(root)
}

// Output:
// undefined
// 1

// The below tests that the realm's block (of 1 variable) changed.  The first
// element image in the package (block) is for the "main" function, which
// appears first because function declarations are defined in a file before
// vars.

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](31)=
//     @@ -2,9 +2,15 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "N": "AQAAAAAAAAA=",
//     +        "T": {
//     +            "@type": "/gno.PrimitiveType",
//     +            "value": "32"
//     +        }
//     +    }
//      }
