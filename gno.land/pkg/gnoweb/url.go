package gnoweb

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrURLInvalidPath = errors.New("invalid path")

// rePkgOrRealmPath matches and validates a flexible path.
var rePkgOrRealmPath = regexp.MustCompile(`^/[a-z][a-z0-9_/]*$`)

// GnoURL decomposes the parts of an URL to query a realm.
type GnoURL struct {
	// Example full path:
	// gno.land/r/demo/users/render.gno:jae$help&a=b?c=d

	Domain   string     // gno.land
	Path     string     // /r/demo/users
	Args     string     // jae
	WebQuery url.Values // help&a=b
	Query    url.Values // c=d
	File     string     // render.gno
}

// EncodeFlag is used to specify which URL components to encode.
type EncodeFlag int

const (
	EncodePath     EncodeFlag = 1 << iota // Encode the path component
	EncodeArgs                            // Encode the arguments component
	EncodeWebQuery                        // Encode the web query component
	EncodeQuery                           // Encode the query component
	EncodeNoEscape                        // Disable escaping of arguments
)

// Has checks if the EncodeFlag contains all the specified flags.
func (f EncodeFlag) Has(flags EncodeFlag) bool {
	return f&flags != 0
}

// Encode encodes the URL components based on the provided flags.
func (gnoURL GnoURL) Encode(encodeFlags EncodeFlag) string {
	var urlstr strings.Builder

	if encodeFlags.Has(EncodePath) {
		path := gnoURL.Path
		if !encodeFlags.Has(EncodeNoEscape) {
			path = url.PathEscape(path)
		}

		urlstr.WriteString(gnoURL.Path)
	}

	if len(gnoURL.File) > 0 {
		urlstr.WriteRune('/')
		urlstr.WriteString(gnoURL.File)
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

func escapeDollarSign(s string) string {
	return strings.ReplaceAll(s, "$", "%24")
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

// IsPure checks if the URL path represents a pure path.
func (gnoURL GnoURL) IsPure() bool {
	return strings.HasPrefix(gnoURL.Path, "/p/")
}

// IsRealm checks if the URL path represents a realm path.
func (gnoURL GnoURL) IsRealm() bool {
	return strings.HasPrefix(gnoURL.Path, "/r/")
}

// IsFile checks if the URL path represents a file.
func (gnoURL GnoURL) IsFile() bool {
	return gnoURL.File != ""
}

// IsDir checks if the URL path represents a directory.
func (gnoURL GnoURL) IsDir() bool {
	return !gnoURL.IsFile() &&
		len(gnoURL.Path) > 0 && gnoURL.Path[len(gnoURL.Path)-1] == '/'
}

func (gnoURL GnoURL) IsValid() bool {
	return rePkgOrRealmPath.MatchString(gnoURL.Path)
}

// ParseGnoURL parses a URL into a GnoURL structure, extracting and validating its components.
func ParseGnoURL(u *url.URL) (*GnoURL, error) {
	var webargs string
	path, args, found := strings.Cut(u.EscapedPath(), ":")
	if found {
		args, webargs, _ = strings.Cut(args, "$")
	} else {
		path, webargs, _ = strings.Cut(path, "$")
	}

	upath, err := url.PathUnescape(path)
	if err != nil {
		return nil, fmt.Errorf("unable to unescape path %q: %w", path, err)
	}

	var file string
	if ext := filepath.Ext(upath); ext != "" {
		file = filepath.Base(upath)
		upath = strings.TrimSuffix(upath, file)

		// Trim last slash
		if i := strings.LastIndexByte(upath, '/'); i > 0 {
			upath = upath[:i]
		}
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
		File:     file,
	}, nil
}
