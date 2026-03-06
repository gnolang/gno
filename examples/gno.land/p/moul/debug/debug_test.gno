package debug

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestPackage(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm("g1user"))

	testPackage(t)
}

func testPackage(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

	// no debug
	got := Render("")
	expected := ``
	uassert.Equal(t, expected, got)

	// debug without logs
	got = Render("?debug=1")
	expected = `<details><summary>debug</summary>

### Metadata
| Key | Value |
| --- | --- |
| ±std.CurrentRealm().PkgPath()± | gno.land/r/test/test |
| ±std.CurrentRealm().Address()± | g1z7fga7u94pdmamlvcrtvsfwxgsye0qv3rres7n |
| ±std.PreviousRealm().PkgPath()± |  |
| ±std.PreviousRealm().Address()± | g1user |
| ±std.ChainHeight()± | 123 |
| ±time.Now().Format(time.RFC3339)± | 2009-02-13T23:31:30Z |

</details>
`
	expected = strings.ReplaceAll(expected, "±", "`")

	println("###################")
	println(got)
	println("###################")
	println(expected)
	println("###################")

	uassert.Equal(t, expected, got)

	// debug with logs
	var d Debug
	d.Log("hello world!")
	d.Log("foobar")
	got = d.Render("?debug=1")
	expected = `<details><summary>debug</summary>

### Logs
- hello world!
- foobar
### Metadata
| Key | Value |
| --- | --- |
| ±std.CurrentRealm().PkgPath()± | gno.land/r/test/test |
| ±std.CurrentRealm().Address()± | g1z7fga7u94pdmamlvcrtvsfwxgsye0qv3rres7n |
| ±std.PreviousRealm().PkgPath()± |  |
| ±std.PreviousRealm().Address()± | g1user |
| ±std.ChainHeight()± | 123 |
| ±time.Now().Format(time.RFC3339)± | 2009-02-13T23:31:30Z |

</details>
`
	expected = strings.ReplaceAll(expected, "±", "`")
	uassert.Equal(t, got, expected)
}
