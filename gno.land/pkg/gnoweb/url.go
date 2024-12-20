package gnoweb

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

type PathKind byte

const (
	KindUnknown PathKind = 0
	KindRealm   PathKind = 'r'
	KindPure    PathKind = 'p'
)

// rePkgOrRealmPath matches and validates a realm or package path.
var rePkgOrRealmPath = regexp.MustCompile(`^/[a-z][a-zA-Z0-9_/.]*$`)

// GnoURL decomposes the parts of an URL to query a realm.
type GnoURL struct {
	// Example full path:
	// gno.land/r/demo/users:jae$help&a=b?c=d

	Domain   string     // gno.land
	Path     string     // /r/demo/users
	Args     string     // jae
	WebQuery url.Values // help&a=b
	Query    url.Values // c=d
}

// EncodeFlag is used to compose and encode URL components.
type EncodeFlag int

const (
	EncodePath EncodeFlag = 1 << iota
	EncodeArgs
	EncodeWebQuery
	EncodeQuery
	EncodeNoEscape // Disable escaping on arg
)

// Has checks if the EncodeFlag contains all the specified flags.
func (f EncodeFlag) Has(flags EncodeFlag) bool {
	return f&flags != 0
}

// Encode encodes the URL components based on the provided flags.
// Encode assums the URL is valid.
func (gnoURL GnoURL) Encode(encodeFlags EncodeFlag) string {
	var urlstr strings.Builder

	if encodeFlags.Has(EncodePath) {
		urlstr.WriteString(gnoURL.Path)
	}

	if encodeFlags.Has(EncodeArgs) && gnoURL.Args != "" {
		if encodeFlags.Has(EncodePath) {
			urlstr.WriteRune(':')
		}

		// XXX: Arguments should ideally always be escaped,
		// but this may require changes in some realms.
		args := gnoURL.Args
		if !encodeFlags.Has(EncodeNoEscape) {
			args = escapeDollarSign(url.PathEscape(args))
		}

		urlstr.WriteString(args)
	}

	if encodeFlags.Has(EncodeWebQuery) && len(gnoURL.WebQuery) > 0 {
		urlstr.WriteRune('$')
		urlstr.WriteString(gnoURL.WebQuery.Encode())
	}

	if encodeFlags.Has(EncodeQuery) && len(gnoURL.Query) > 0 {
		urlstr.WriteRune('?')
		urlstr.WriteString(gnoURL.Query.Encode())

	}

	return urlstr.String()
}

// EncodeArgs encodes the arguments and query parameters into a string.
// This function is intended to be passed as a realm `Render` argument.
func (gnoURL GnoURL) EncodeArgs() string {
	return gnoURL.Encode(EncodeArgs | EncodeQuery | EncodeNoEscape)
}

// EncodeURL encodes the path, arguments, and query parameters into a string.
// This function provides the full representation of the URL without the web query.
func (gnoURL GnoURL) EncodeURL() string {
	return gnoURL.Encode(EncodePath | EncodeArgs | EncodeQuery)
}

// EncodeWebURL encodes the path, package arguments, web query, and query into a string.
// This function provides the full representation of the URL.
func (gnoURL GnoURL) EncodeWebURL() string {
	return gnoURL.Encode(EncodePath | EncodeArgs | EncodeWebQuery | EncodeQuery)
}

// Kind determines the kind of path (invalid, realm, or pure) based on the path structure.
func (gnoURL GnoURL) Kind() PathKind {
	if len(gnoURL.Path) > 2 && gnoURL.Path[0] == '/' && gnoURL.Path[2] == '/' {
		switch k := PathKind(gnoURL.Path[1]); k {
		case KindPure, KindRealm:
			return k
		}
	}
	return KindUnknown
}

func (gnoURL GnoURL) IsValid() bool {
	return rePkgOrRealmPath.MatchString(gnoURL.Path)
}

// IsDir checks if the URL path represents a directory.
func (gnoURL GnoURL) IsDir() bool {
	return len(gnoURL.Path) > 0 && gnoURL.Path[len(gnoURL.Path)-1] == '/'
}

// IsFile checks if the URL path represents a file.
func (gnoURL GnoURL) IsFile() bool {
	return filepath.Ext(gnoURL.Path) != ""
}

var ErrURLInvalidPath = errors.New("invalid or malformed path")

// ParseGnoURL parses a URL into a GnoURL structure, extracting and validating its components.
func ParseGnoURL(u *url.URL) (*GnoURL, error) {
	var webargs string
	path, args, found := strings.Cut(u.EscapedPath(), ":")
	if found {
		args, webargs, _ = strings.Cut(args, "$")
	} else {
		path, webargs, _ = strings.Cut(path, "$")
	}

	// NOTE: `PathUnescape` should already unescape dollar signs.
	upath, err := url.PathUnescape(path)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape path %q: %w", path, err)
	}

	if !rePkgOrRealmPath.MatchString(upath) {
		return nil, fmt.Errorf("%w: %q", ErrURLInvalidPath, upath)
	}

	webquery := url.Values{}
	if len(webargs) > 0 {
		var parseErr error
		if webquery, parseErr = url.ParseQuery(webargs); parseErr != nil {
			return nil, fmt.Errorf("unable to parse webquery %q: %w", webargs, parseErr)
		}
	}

	uargs, err := url.PathUnescape(args)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape args %q: %w", args, err)
	}

	return &GnoURL{
		Path:     upath,
		Args:     uargs,
		WebQuery: webquery,
		Query:    u.Query(),
		Domain:   u.Hostname(),
	}, nil
}

// escapeDollarSign replaces dollar signs with their URL-encoded equivalent.
func escapeDollarSign(s string) string {
	return strings.ReplaceAll(s, "$", "%24")
}
