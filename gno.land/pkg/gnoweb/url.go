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
var rePkgOrRealmPath = regexp.MustCompile(`^/[a-z]/[a-zA-Z0-9_/.]*$`)

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

type EncodeFlag int

const (
	EncodePath EncodeFlag = 1 << iota
	EncodeArgs
	EncodeWebQuery
	EncodeQuery
)

// Encode encodes the URL components based on the provided flags.
func (gnoURL GnoURL) Encode(encodeFlags EncodeFlag) string {
	var urlstr strings.Builder

	if encodeFlags&EncodePath != 0 {
		urlstr.WriteString(gnoURL.Path)
	}

	if encodeFlags&EncodeArgs != 0 && gnoURL.Args != "" {
		if encodeFlags&EncodePath != 0 {
			urlstr.WriteString(":")
		}
		urlstr.WriteString(gnoURL.Args)
	}

	if encodeFlags&EncodeWebQuery != 0 && len(gnoURL.WebQuery) > 0 {
		urlstr.WriteString("$" + gnoURL.WebQuery.Encode())
	}

	if encodeFlags&EncodeQuery != 0 && len(gnoURL.Query) > 0 {
		urlstr.WriteString("?" + gnoURL.Query.Encode())
	}

	return urlstr.String()
}

// EncodeArgsQuery encodes the arguments and query parameters into a string.
// This function is intended to be passed as a realm `Render` argument.
func (gnoURL GnoURL) EncodeArgsQuery() string {
	return gnoURL.Encode(EncodeArgs | EncodeQuery)
}

// EncodePathArgsQuery encodes the path, arguments, and query parameters into a string.
// This function provides the full representation of the URL without the web query.
func (gnoURL GnoURL) EncodePathArgsQuery() string {
	return gnoURL.Encode(EncodePath | EncodeArgs | EncodeQuery)
}

// EncodeWebPath encodes the path, package arguments, web query, and query into a string.
// This function provides the full representation of the URL.
func (gnoURL GnoURL) EncodeWebPath() string {
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

// IsDir checks if the URL path represents a directory.
func (gnoURL GnoURL) IsDir() bool {
	return len(gnoURL.Path) > 0 && gnoURL.Path[len(gnoURL.Path)-1] == '/'
}

// IsFile checks if the URL path represents a file.
func (gnoURL GnoURL) IsFile() bool {
	return filepath.Ext(gnoURL.Path) != ""
}

var (
	ErrURLMalformedPath   = errors.New("malformed path")
	ErrURLInvalidPathKind = errors.New("invalid path kind")
)

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

	if !rePkgOrRealmPath.MatchString(upath) {
		return nil, fmt.Errorf("%w: %q", ErrURLMalformedPath, upath)
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
