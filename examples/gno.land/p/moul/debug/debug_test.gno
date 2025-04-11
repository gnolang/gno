package debug

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestPackage(t *testing.T) {
	testing.SetRealm(std.NewUserRealm("g1user"))

	// no debug
	got := Render("")
	expected := ``
	uassert.Equal(t, got, expected)

	// debug without logs
	got = Render("?debug=1")
	expected = `<details><summary>debug</summary>

### Metadata
| Key | Value |
| --- | --- |
| ±std.CurrentRealm().PkgPath()± |  |
| ±std.CurrentRealm().Address()± | g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm |
| ±std.PreviousRealm().PkgPath()± |  |
| ±std.PreviousRealm().Address()± | g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm |
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

	uassert.Equal(t, got, expected)

	return

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
| ±std.CurrentRealm().PkgPath()± |  |
| ±std.CurrentRealm().Address()± | g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm |
| ±std.PreviousRealm().PkgPath()± |  |
| ±std.PreviousRealm().Address()± | g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm |
| ±std.ChainHeight()± | 123 |
| ±time.Now().Format(time.RFC3339)± | 2009-02-13T23:31:30Z |
		
</details>
`
	expected = strings.ReplaceAll(expected, "±", "`")
	uassert.Equal(t, got, expected)
}
