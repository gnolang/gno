package derive

import (
	"chain"
	"chain/runtime"
	"net/url"
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/leon/hor"
)

var cd = runtime.ChainDomain()

func Render(path string) string {
	out := "# Derive Realm Address\n\n"

	out += `<gno-form>
 	 <gno-input name="pkgpath" placeholder="gno.land/{r,p}/..." />
	</gno-form>
`

	if strings.Contains(path, "pkgpath") {
		parsed := parsePkgPath(path)
		if parsed == "" {
			out += "Please input a valid pkgpath."
			return out
		}
		out += ufmt.Sprintf("### [%s](%s) matches %s", parsed, strings.TrimPrefix(parsed, "gno.land"), chain.PackageAddress(parsed))
	}

	return out
}

func parsePkgPath(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return ""
	}

	pkgpath := u.Query().Get("pkgpath")
	if pkgpath == "" {
		return ""
	}

	switch {
	case strings.HasPrefix(pkgpath, "gno.land/r/"),
		strings.HasPrefix(pkgpath, "gno.land/p/"):
		// already absolute and valid
		return pkgpath

	case strings.HasPrefix(pkgpath, "r/"),
		strings.HasPrefix(pkgpath, "p/"):
		// relative path, normalize
		return "gno.land/" + pkgpath

	case strings.HasPrefix(pkgpath, "/r/"),
		strings.HasPrefix(pkgpath, "/p/"):
		// relative with leading slash
		return "gno.land" + pkgpath

	default:
		// reject invalid
		return ""
	}
}

func init() {
	hor.Register(cross, "Derive", "Quickly derive a Gno address from a pkg path.")
}
