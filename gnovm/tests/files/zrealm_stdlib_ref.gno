// PKGPATH: gno.land/r/test
package test

import "strconv"

// Store a reference to an existing stdlib object (not a copy).
var ref interface{}

func main(cur realm) {
	// This should store a reference to strconv.Itoa itself,
	// which is a FuncValue owned by the strconv package.
	// Does this incref the stdlib object?
	ref = strconv.Itoa
	println(ref.(func(int) string)(42))
}

// Output:
// 42

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](199)=
//     @@ -2,9 +2,40 @@
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
//     +        "T": {
//     +            "@type": "/gno.FuncType",
//     +            "Params": [
//     +                {
//     +                    "Embedded": false,
//     +                    "Name": "i",
//     +                    "Tag": "",
//     +                    "Type": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "32"
//     +                    }
//     +                }
//     +            ],
//     +            "Results": [
//     +                {
//     +                    "Embedded": false,
//     +                    "Name": ".res.0",
//     +                    "Tag": "",
//     +                    "Type": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "16"
//     +                    }
//     +                }
//     +            ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "aa2f3c8cd23e4e15c3eebc5463ee00c87348472d",
//     +            "ObjectID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:869"
//     +        }
//     +    }
//      }
