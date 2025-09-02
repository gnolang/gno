// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/crossrealm"
	"gno.land/r/tests/crossrealm_b"
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
// hello A cur=gno.land/r/tests/crossrealm_b prev=gno.land/r/tests/crossrealm
// hello B cur=gno.land/r/tests/crossrealm_b prev=gno.land/r/tests/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/tests/crossrealm"]
// u[83c395a7172ae8aebc131bd64cc048bbd2b31851:7](220)=
//     @@ -2,9 +2,28 @@
//          "ObjectInfo": {
//              "ID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:7",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "N": "UmVhRG9OTFk=",
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/tests/crossrealm_b.fooer"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:4"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:4](7)=
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "357",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/tests/crossrealm_b"]
// finalizerealm["gno.land/r/tests/crossrealm"]
// finalizerealm["gno.land/r/tests/crossrealm_b"]
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:5](5)=
//     @@ -7,14 +7,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "A"
//     +                "value": "B"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:5",
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "15",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:4",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/tests/crossrealm_b"]
// finalizerealm["gno.land/r/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
