// PKGPATH: gno.land/r/docs/events/filetests/zemit

package zemit

import "gno.land/r/docs/events"

func init(cur realm) {
	events.Emit(cross(cur), "hello")
}

func main() {}

// Events:
// [
//   {
//     "type": "ExampleEvent",
//     "attrs": [
//       {
//         "key": "key",
//         "value": "hello"
//       }
//     ],
//     "pkg_path": "gno.land/r/docs/events"
//   }
// ]
