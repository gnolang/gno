// Package oci pulls files out of a container image by digest over the
// registry HTTP API (the OCI distribution spec), with no container runtime.
// Every byte it uses is checked against a digest: the index against the
// digest the caller pins, each manifest and layer against the descriptor
// that names it.
package oci

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Media types the client reads. Docker's schema 2 types are accepted next to
// OCI's, since a registry may serve either for the same image.
const (
	MediaTypeOCIIndex      = "application/vnd.oci.image.index.v1+json"
	MediaTypeOCIManifest   = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeDockerList    = "application/vnd.docker.distribution.manifest.list.v2+json"
	MediaTypeDockerV2      = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeOCILayerGzip  = "application/vnd.oci.image.layer.v1.tar+gzip"
	MediaTypeOCILayer      = "application/vnd.oci.image.layer.v1.tar"
	MediaTypeDockerLayerGz = "application/vnd.docker.image.rootfs.diff.tar.gzip"

	maxManifestSize = 4 << 20
	// MaxLayerSize bounds one layer's descriptor size. gnoland images are
	// tens of megabytes; the pinned digest is what makes a layer trusted,
	// this only stops a malformed descriptor from filling the disk first.
	MaxLayerSize = 2 << 30
)

var digestRE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Descriptor points at a blob or manifest by digest.
type Descriptor struct {
	MediaType string    `json:"mediaType"`
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
	Platform  *Platform `json:"platform,omitempty"`
}

// Platform is the os/arch an image manifest is built for.
type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
}

type manifest struct {
	MediaType string       `json:"mediaType"`
	Manifests []Descriptor `json:"manifests"` // index
	Layers    []Descriptor `json:"layers"`    // image manifest
	size      int64
}

// Ref is an image reference split into what the registry API needs.
type Ref struct {
	Host       string // e.g. ghcr.io
	Repository string // e.g. gnolang/gno/gnoland
}

// ParseRef splits host/repository[:tag]. The tag is dropped: the client
// fetches by digest only, since a tag can be moved.
func ParseRef(ref string) (Ref, error) {
	host, rest, ok := strings.Cut(ref, "/")
	if !ok || !strings.ContainsAny(host, ".:") || rest == "" {
		return Ref{}, fmt.Errorf("image reference %q has no registry host", ref)
	}
	if i := strings.LastIndex(rest, "@"); i >= 0 {
		rest = rest[:i]
	}
	if i := strings.LastIndex(rest, ":"); i >= 0 {
		rest = rest[:i]
	}
	return Ref{Host: host, Repository: rest}, nil
}

// Client talks to one or more registries over HTTPS.
type Client struct {
	HTTP *http.Client
	// BaseURL, when set, replaces https://<host> for every reference: a
	// mirror, or a test server.
	BaseURL string
	// Attempts is how many times a request is tried before giving up on a
	// network error, a 5xx, a 429 or a body that fails its digest.
	Attempts int
	// Backoff is the wait before the second attempt; it doubles after.
	Backoff time.Duration
	// AttemptTimeout bounds one attempt, download included.
	AttemptTimeout time.Duration

	mu     sync.Mutex
	tokens map[string]string // base URL + repository -> bearer token
}

// NewClient returns a client with sensible defaults. There is no overall
// HTTP timeout, which would fail a large layer on a slow link on every
// attempt: a server that does not answer is cut by ResponseHeaderTimeout,
// and an attempt by AttemptTimeout.
func NewClient() *Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.ResponseHeaderTimeout = time.Minute
	return &Client{
		HTTP: &http.Client{
			Transport:     tr,
			CheckRedirect: noDowngrade,
		},
		Attempts:       4,
		Backoff:        2 * time.Second,
		AttemptTimeout: 30 * time.Minute,
	}
}

// noDowngrade follows redirects (ghcr.io sends blobs to a CDN) but never
// from https to anything else.
func noDowngrade(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("refusing a redirect from https to %s", req.URL.Scheme)
	}
	return nil
}

func (c *Client) base(r Ref) string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return "https://" + r.Host
}

// ResolveManifest fetches the manifest digest pins and, if it is an index,
// the image manifest for platform ("os/arch"). It returns the layers in
// application order.
func (c *Client) ResolveManifest(ctx context.Context, r Ref, digest, platform string) ([]Descriptor, error) {
	if !digestRE.MatchString(digest) {
		return nil, fmt.Errorf("digest %q is not sha256:<64 hex>", digest)
	}
	m, mt, err := c.fetchManifest(ctx, r, digest)
	if err != nil {
		return nil, err
	}
	switch mt {
	case MediaTypeOCIIndex, MediaTypeDockerList:
		d, err := pickPlatform(m.Manifests, platform)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", digest, err)
		}
		m, mt, err = c.fetchManifest(ctx, r, d.Digest)
		if err != nil {
			return nil, err
		}
		if m.size != d.Size {
			return nil, fmt.Errorf("%s: platform manifest is %d bytes, its descriptor says %d", digest, m.size, d.Size)
		}
		if mt != MediaTypeOCIManifest && mt != MediaTypeDockerV2 {
			return nil, fmt.Errorf("%s: platform entry %s is a %s, not an image manifest", digest, d.Digest, mt)
		}
	case MediaTypeOCIManifest, MediaTypeDockerV2:
		// A single-platform image: the caller pinned it, so trust the pin.
	default:
		return nil, fmt.Errorf("%s: unsupported manifest media type %q", digest, mt)
	}
	if len(m.Layers) == 0 {
		return nil, fmt.Errorf("%s: image manifest has no layers", digest)
	}
	for _, l := range m.Layers {
		if !digestRE.MatchString(l.Digest) || l.Size < 0 || l.Size > MaxLayerSize {
			return nil, fmt.Errorf("%s: layer descriptor %q/%d is malformed", digest, l.Digest, l.Size)
		}
	}
	return m.Layers, nil
}

func pickPlatform(ds []Descriptor, platform string) (Descriptor, error) {
	wantOS, wantArch, ok := strings.Cut(platform, "/")
	if !ok {
		return Descriptor{}, fmt.Errorf("platform %q is not os/arch", platform)
	}
	for _, d := range ds {
		if d.Platform != nil && d.Platform.OS == wantOS && d.Platform.Architecture == wantArch {
			return d, nil
		}
	}
	return Descriptor{}, fmt.Errorf("image index has no %s manifest", platform)
}

func (c *Client) fetchManifest(ctx context.Context, r Ref, digest string) (*manifest, string, error) {
	accept := strings.Join([]string{MediaTypeOCIIndex, MediaTypeDockerList, MediaTypeOCIManifest, MediaTypeDockerV2}, ",")
	var body []byte
	err := c.retry(ctx, func(ctx context.Context) error {
		resp, err := c.get(ctx, r, "/manifests/"+digest, accept)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(io.LimitReader(resp.Body, maxManifestSize+1))
		if err != nil {
			return retryable(err)
		}
		if len(body) > maxManifestSize {
			return fmt.Errorf("manifest %s is larger than %d bytes", digest, maxManifestSize)
		}
		if got := sha256Digest(body); got != digest {
			return retryable(fmt.Errorf("manifest %s: body has digest %s", digest, got))
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	var m manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, "", fmt.Errorf("manifest %s: %w", digest, err)
	}
	m.size = int64(len(body))
	// Only the body is covered by the digest, so the media type comes from
	// it, never from the Content-Type header. Both fields are optional in
	// the OCI spec; without one, the shape decides.
	mt := m.MediaType
	if mt == "" {
		if len(m.Manifests) > 0 {
			mt = MediaTypeOCIIndex
		} else {
			mt = MediaTypeOCIManifest
		}
	}
	return &m, mt, nil
}

// FetchBlob downloads the blob d names into a new file in dir, checking its
// size and digest, and returns the file's path. A truncated or altered
// download is retried, then refused; no partial file is left behind.
func (c *Client) FetchBlob(ctx context.Context, r Ref, d Descriptor, dir string) (string, error) {
	if d.Size > MaxLayerSize {
		return "", fmt.Errorf("blob %s is %d bytes, more than %d", d.Digest, d.Size, MaxLayerSize)
	}
	var path string
	err := c.retry(ctx, func(ctx context.Context) error {
		f, err := os.CreateTemp(dir, "blob-*")
		if err != nil {
			return err
		}
		ok := false
		defer func() {
			f.Close()
			if !ok {
				os.Remove(f.Name())
			}
		}()
		resp, err := c.get(ctx, r, "/blobs/"+d.Digest, "")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		h := sha256.New()
		n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, d.Size+1))
		if err != nil {
			return retryable(fmt.Errorf("blob %s: %w", d.Digest, err))
		}
		if n != d.Size {
			return retryable(fmt.Errorf("blob %s: got %d bytes, want %d", d.Digest, n, d.Size))
		}
		if got := "sha256:" + hex.EncodeToString(h.Sum(nil)); got != d.Digest {
			return retryable(fmt.Errorf("blob %s: body has digest %s", d.Digest, got))
		}
		if err := f.Close(); err != nil {
			return err
		}
		ok, path = true, f.Name()
		return nil
	})
	return path, err
}

// get issues one GET, answering a bearer challenge once.
func (c *Client) get(ctx context.Context, r Ref, path, accept string) (*http.Response, error) {
	u := c.base(r) + "/v2/" + r.Repository + path
	do := func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		if accept != "" {
			req.Header.Set("Accept", accept)
		}
		if tok := c.token(r); tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, retryable(err)
		}
		return resp, nil
	}
	resp, err := do()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		challenge := resp.Header.Get("WWW-Authenticate")
		resp.Body.Close()
		if err := c.authenticate(ctx, r, challenge); err != nil {
			return nil, err
		}
		if resp, err = do(); err != nil {
			return nil, err
		}
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		err := fmt.Errorf("GET %s: %s", u, resp.Status)
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			return nil, retryable(err)
		}
		return nil, err
	}
	return resp, nil
}

func (c *Client) tokenKey(r Ref) string { return c.base(r) + "/" + r.Repository }

func (c *Client) token(r Ref) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tokens[c.tokenKey(r)]
}

// authenticate answers a "Bearer realm=...,service=...,scope=..." challenge
// with an anonymous token request, as public registries such as ghcr.io
// expect for pulls.
func (c *Client) authenticate(ctx context.Context, r Ref, challenge string) error {
	scheme, params, _ := strings.Cut(challenge, " ")
	if !strings.EqualFold(scheme, "Bearer") {
		return fmt.Errorf("registry asks for %q authentication; only anonymous bearer tokens are supported", scheme)
	}
	p := parseChallenge(params)
	realm := p["realm"]
	if realm == "" {
		return errors.New("bearer challenge has no realm")
	}
	q := url.Values{}
	if s := p["service"]; s != "" {
		q.Set("service", s)
	}
	scope := p["scope"]
	if scope == "" {
		scope = "repository:" + r.Repository + ":pull"
	}
	q.Set("scope", scope)
	u, err := url.Parse(realm)
	if err != nil {
		return fmt.Errorf("bearer realm: %w", err)
	}
	// The token is sent back to the registry, so the realm must not be a
	// way to reach (or downgrade to) anything but an https token service. A
	// plain-http registry (a local mirror, a test) may use a plain-http realm.
	if u.Scheme != "https" && !(u.Scheme == "http" && strings.HasPrefix(c.base(r), "http://")) {
		return fmt.Errorf("bearer realm %q is not https", realm)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return retryable(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request: %s", resp.Status)
	}
	var tr struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tr); err != nil {
		return fmt.Errorf("token response: %w", err)
	}
	tok := tr.Token
	if tok == "" {
		tok = tr.AccessToken
	}
	if tok == "" {
		return errors.New("token response has no token")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tokens == nil {
		c.tokens = map[string]string{}
	}
	c.tokens[c.tokenKey(r)] = tok
	return nil
}

func parseChallenge(s string) map[string]string {
	out := map[string]string{}
	for s != "" {
		var kv string
		// Values are quoted and may contain commas (scope lists).
		k, rest, ok := strings.Cut(s, "=")
		if !ok {
			break
		}
		k = strings.TrimSpace(k)
		if strings.HasPrefix(rest, `"`) {
			end := strings.Index(rest[1:], `"`)
			if end < 0 {
				break
			}
			kv, s = rest[1:end+1], strings.TrimPrefix(strings.TrimSpace(rest[end+2:]), ",")
		} else {
			kv, s, _ = strings.Cut(rest, ",")
		}
		out[strings.ToLower(k)] = kv
		s = strings.TrimSpace(s)
	}
	return out
}

type retryableError struct{ error }

func (e retryableError) Unwrap() error { return e.error }

func retryable(err error) error { return retryableError{err} }

// retry runs f until it succeeds, fails for good, or runs out of attempts;
// each attempt gets its own deadline.
func (c *Client) retry(ctx context.Context, f func(ctx context.Context) error) error {
	attempts := max(c.Attempts, 1)
	wait := c.Backoff
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			wait *= 2
		}
		err = c.attempt(ctx, f)
		var re retryableError
		if err == nil || !errors.As(err, &re) || ctx.Err() != nil {
			return err
		}
	}
	return fmt.Errorf("after %d attempts: %w", attempts, err)
}

func (c *Client) attempt(ctx context.Context, f func(ctx context.Context) error) error {
	if c.AttemptTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.AttemptTimeout)
		defer cancel()
	}
	return f(ctx)
}

func sha256Digest(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}
