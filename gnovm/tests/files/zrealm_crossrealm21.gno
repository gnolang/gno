// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFooCross(cross)
	f.SetS("B")
	crossrealm.CallFooerFooCross(cross)
	print(".")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello B cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:7]=
//     @@ -1,9 +1,28 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7",
//     -        "ModTime": "0",
//     +        "ModTime": "38",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "N": "UmVhRG9OTFk=",
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/demo/tests/crossrealm_b.fooer"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "38",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:8]=
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
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:8",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm_test"]
