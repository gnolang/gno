package markdown

import (
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ipfsRetiredGateways are public IPFS gateways that no longer serve
// hotlinked content (see gno.land/adr/pr6270_gnoweb_ipfs_gateway.md); URLs
// on them are rewritten to the configured gateway. The value reports
// whether the gateway also served the subdomain form <cid>.ipfs.<host>.
var ipfsRetiredGateways = map[string]bool{
	"ipfs.io":             false,
	"gateway.ipfs.io":     false,
	"cloudflare-ipfs.com": false,
	"cf-ipfs.com":         true,
	"dweb.link":           true,
	"nftstorage.link":     true,
	"w3s.link":            true,
}

// dnsLinkLabelDecoder reverses how subdomain gateways inline a DNSLink name
// into one DNS label: "--" stands for "-" and a lone "-" for ".", so
// "en-wikipedia--on--ipfs-org" is "en.wikipedia-on-ipfs.org".
var dnsLinkLabelDecoder = strings.NewReplacer("--", "-", "-", ".")

// ipfsTransformer rewrites IPFS destinations of links, autolinks and
// images to the gateway set in the parser context.
type ipfsTransformer struct{}

func (t *ipfsTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	gateway, _ := getIPFSGatewayFromContext(pc)
	if gateway == "" {
		return
	}

	type autolinkRewrite struct {
		node *ast.AutoLink
		dest []byte
	}
	var autolinks []autolinkRewrite
	source := reader.Source()
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Link:
			if dest, ok := rewriteIPFSDestination(gateway, n.Destination); ok {
				n.Destination = dest
			}
		case *ast.Image:
			if dest, ok := rewriteIPFSDestination(gateway, n.Destination); ok {
				n.Destination = dest
			}
		case *ast.AutoLink:
			if n.AutoLinkType != ast.AutoLinkURL {
				break
			}
			if dest, ok := rewriteIPFSDestination(gateway, n.URL(source)); ok {
				autolinks = append(autolinks, autolinkRewrite{n, dest})
			}
		}

		return ast.WalkContinue, nil
	})

	// An autolink has no destination to rewrite, so it is replaced by an
	// equivalent link. That happens after the walk: RemoveChild clears the
	// removed node's sibling pointers, which would end the walk of its
	// parent early.
	for _, a := range autolinks {
		parent := a.node.Parent()
		parent.ReplaceChild(parent, a.node, newLinkFromAutoLink(a.node, source, a.dest))
	}
}

// rewriteIPFSDestination rewrites a raw markdown destination and returns
// it escaped for markdown again (see escapeDestination).
func rewriteIPFSDestination(gateway string, dest []byte) ([]byte, bool) {
	resolved := trimLeadingControlAndSpace(resolveDestination(dest))
	if !mayBeIPFSURL(resolved) {
		return nil, false
	}

	out, ok := rewriteIPFSURL(gateway, string(resolved))
	if !ok {
		return nil, false
	}
	return escapeDestination(out), true
}

// mayBeIPFSURL is a cheap filter that spares url.Parse on the common,
// non-IPFS destination. Every URL rewriteIPFSURL rewrites contains "pfs"
// or "pns" in some ASCII case (ipfs://, /ipfs/, .ipfs.), or a percent
// escape that url.Parse decodes into one.
func mayBeIPFSURL(b []byte) bool {
	for i, c := range b {
		if c == '%' {
			return true
		}
		if c|0x20 == 'p' && i+2 < len(b) && (b[i+1]|0x20 == 'f' || b[i+1]|0x20 == 'n') && b[i+2]|0x20 == 's' {
			return true
		}
	}
	return false
}

// rewriteIPFSURL maps an IPFS URL to the same content on gateway, a
// normalized origin:
//
//	ipfs://<cid>/<path>                     -> <gateway>/ipfs/<cid>/<path>
//	ipns://<name>/<path>                    -> <gateway>/ipns/<name>/<path>
//	https://ipfs.io/ipfs/<cid>/<path>       -> <gateway>/ipfs/<cid>/<path>
//	https://<cid>.ipfs.dweb.link/<path>     -> <gateway>/ipfs/<cid>/<path>
//	https://<label>.ipns.dweb.link/<path>   -> <gateway>/ipns/<name>/<path>
//
// Protocol-relative URLs (//ipfs.io/ipfs/...) are handled like https ones.
// Query and fragment are kept. It reports false, and leaves the URL alone,
// for anything else: other hosts, userinfo, explicit ports, a CID or name
// that is not plain alphanumerics (plus dots and hyphens for IPNS), a path
// with a "." or ".." segment, or a "uri" query parameter.
func rewriteIPFSURL(gateway, raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil {
		return "", false
	}

	var namespace, name, rest string
	switch u.Scheme { // url.Parse lowercases the scheme
	case "ipfs", "ipns":
		namespace, name, rest = u.Scheme, u.Host, u.EscapedPath()
	case "http", "https", "": // "" with a host: protocol-relative //host/...
		host := strings.ToLower(u.Host) // keeps any port, so ported hosts never match
		if _, retired := ipfsRetiredGateways[host]; retired {
			namespace, name, rest = splitGatewayPath(u.EscapedPath())
		} else {
			namespace, name = splitGatewaySubdomain(host)
			rest = u.EscapedPath()
		}
	default:
		return "", false
	}

	// Also rejects the empty name of an unmatched split, and any port or
	// opaque form, which leave a ':' in the name or no host at all.
	if !validIPFSName(namespace, name) {
		return "", false
	}
	// On the gateway the CID is a path segment rather than the host, so a
	// ".." after it would resolve to other content. And gateways built on
	// boxo (Kubo, Rainbow) redirect any request with a "uri" query
	// parameter to the content it names.
	if hasDotSegment(u.Path) || hasURIParam(u.RawQuery) {
		return "", false
	}

	out := gateway + "/" + namespace + "/" + name + rest
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		out += "#" + u.EscapedFragment()
	}
	return out, true
}

// splitGatewayPath splits "/ipfs/<cid>/<rest>" into its namespace, CID and
// "/<rest>". It returns empty strings for any other path.
func splitGatewayPath(p string) (namespace, name, rest string) {
	switch {
	case strings.HasPrefix(p, "/ipfs/"):
		namespace = "ipfs"
	case strings.HasPrefix(p, "/ipns/"):
		namespace = "ipns"
	default:
		return "", "", ""
	}

	name = p[len("/ipfs/"):]
	if i := strings.IndexByte(name, '/'); i >= 0 {
		name, rest = name[:i], name[i:]
	}
	return namespace, name, rest
}

// splitGatewaySubdomain splits "<label>.ipfs.<gateway>" or
// "<label>.ipns.<gateway>" for a retired subdomain gateway, decoding an
// inlined DNSLink name. It returns empty strings for any other host.
func splitGatewaySubdomain(host string) (namespace, name string) {
	label, suffix, _ := strings.Cut(host, ".")
	namespace, gateway, _ := strings.Cut(suffix, ".")
	if !ipfsRetiredGateways[gateway] {
		return "", ""
	}

	switch namespace {
	case "ipfs":
		return namespace, label
	case "ipns":
		return namespace, dnsLinkLabelDecoder.Replace(label)
	}
	return "", ""
}

// validIPFSName reports whether name is safe to place in a gateway path: a
// CID for /ipfs/ is plain alphanumerics; an IPNS name may also be a DNS
// name, so it can hold dots and hyphens, but must start with an
// alphanumeric and have no empty labels.
func validIPFSName(namespace, name string) bool {
	if name == "" || len(name) > 253 {
		return false
	}

	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
		case namespace == "ipns" && c == '-' && i > 0:
		case namespace == "ipns" && c == '.' && i > 0 && name[i-1] != '.' && i < len(name)-1:
		default:
			return false
		}
	}
	return true
}

// hasDotSegment reports whether the decoded path p has a "." or ".."
// segment. It splits at "\" too, because browsers treat it as "/" in https
// URLs, and the decoded path also exposes %2F, which a gateway may decode
// before it resolves the path.
func hasDotSegment(p string) bool {
	for _, seg := range strings.FieldsFunc(p, func(r rune) bool { return r == '/' || r == '\\' }) {
		if seg == "." || seg == ".." {
			return true
		}
	}
	return false
}

// hasURIParam reports whether rawQuery has a "uri" key, whatever its value.
// url.ParseQuery would drop a pair whose value has a malformed escape, but
// the renderer repairs that escape (%zz becomes %25zz), so the gateway
// would still see the parameter.
func hasURIParam(rawQuery string) bool {
	for _, pair := range strings.Split(rawQuery, "&") {
		key, _, _ := strings.Cut(pair, "=")
		if k, err := url.QueryUnescape(key); err == nil && k == "uri" {
			return true
		}
	}
	return false
}

type ipfsExtension struct{}

// ExtIPFS rewrites IPFS URLs to the gateway set in the parser context
// (GnoContext.IPFSGateway). It does nothing when no gateway is set.
var ExtIPFS = &ipfsExtension{}

// Extend adds the IPFS transformer at priority 400, before the link and
// image validator transformers (500), so both see the gateway URL.
func (e *ipfsExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&ipfsTransformer{}, 400),
	))
}
