// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	switchrealm()

	f := crossrealm_b.Fooer
	withswitch(crossrealm.SetFooer)(f)
	withswitch(crossrealm.CallFooerFooSR2)()
	f.SetS("B")
	withswitch(crossrealm.CallFooerFooSR2)()
	print(".")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello B cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "IsEscaped": true,
//     -        "ModTime": "4",
//     +        "ModTime": "6",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -284,7 +284,26 @@
//                      }
//                  }
//              },
//     -        {},
//     +        {
//     +            "N": "UmVhRG9OTFk=",
//     +            "T": {
//     +                "@type": "/gno.PointerType",
//     +                "Elt": {
//     +                    "@type": "/gno.RefType",
//     +                    "ID": "gno.land/r/demo/tests/crossrealm_b.fooer"
//     +                }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Escaped": true,
//     +                    "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:3"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//     +            }
//     +        },
//              {
//                  "T": {
//                      "@type": "/gno.FuncType",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:3]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "6",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -7,13 +7,13 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "A"
//     +                "value": "B"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//              "RefCount": "1"
//          }
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
