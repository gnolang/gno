// PKGPATH: gno.land/r/test
package test

// Verify that plain byte-index assignment (bs[i] = v) on a Data-backed []byte
// realm variable marks the backing ArrayValue dirty via DidUpdate, so the
// mutation persists across transactions.
var bs = []byte("aabb")

func main(cur realm) {
	bs[0] = 'A'
	bs[1] = 'B'
	println(string(bs))
}

// Output:
// ABbb

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "YWFiYg==",
//     +    "Data": "QUJiYg==",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "181",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
