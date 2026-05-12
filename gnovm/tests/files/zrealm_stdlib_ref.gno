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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](199)=
//     @@ -2,9 +2,40 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
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
//     +            "Hash": "44d7cf6f34ea458657b8b6a40aae82fe8602216a",
//     +            "ObjectID": "c3ac87e8cb3cc9276228f74d38694a208cacb99b:869"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "537",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "77f76390147dcb12d9de104eb549c3a959934c1c",
//     +                "Hash": "e8bfa22926955db775477ccba7f9419291f5295f",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
